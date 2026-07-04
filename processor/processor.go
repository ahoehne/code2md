package processor

import (
	"bufio"
	"code2md/language"
	"code2md/patternMatcher"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Options struct {
	InputFolder      string
	AllowedLanguages map[string]bool
	AllowedFileNames      map[string]bool
	UserIgnorePatterns    []patternMatcher.CompiledPattern
	DefaultIgnorePatterns []patternMatcher.CompiledPattern
	AbsOutputFilePath     string
	MaxFileSize           int64
}

type gitignoreScope struct {
	dir      string
	patterns []patternMatcher.CompiledPattern
}

func ProcessDirectory(opts Options, output io.Writer) error {
	found := false
	var scopes []gitignoreScope

	absInputFolder, err := filepath.Abs(opts.InputFolder)
	if err != nil {
		return fmt.Errorf("resolving input path %s: %w", opts.InputFolder, err)
	}

	ret := filepath.WalkDir(opts.InputFolder, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				fmt.Fprintf(os.Stderr, "Warning: permission denied: %s\n", path)
				return nil
			}
			return fmt.Errorf("accessing path %s: %w", path, err)
		}

		relPath, err := filepath.Rel(opts.InputFolder, path)
		if err != nil {
			return fmt.Errorf("getting relative path for %s: %w", path, err)
		}
		rel := filepath.ToSlash(relPath)

		if rel == "." && d.IsDir() {
			scope, err := loadGitignoreScope(path, rel)
			if err != nil {
				return err
			}
			if len(scope.patterns) > 0 {
				scopes = append(scopes, scope)
			}
			return nil
		}

		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		if !d.IsDir() && opts.AbsOutputFilePath != "" &&
			filepath.Join(absInputFolder, relPath) == opts.AbsOutputFilePath {
			return nil
		}

		ignored, byDefaultPattern := isIgnored(rel, d.IsDir(), opts, scopes)
		if ignored {
			if d.IsDir() {
				return filepath.SkipDir
			}
			manifestBypassesIgnore := byDefaultPattern && opts.AllowedFileNames[d.Name()]
			if !manifestBypassesIgnore {
				return nil
			}
		}

		if d.IsDir() {
			scope, err := loadGitignoreScope(path, rel)
			if err != nil {
				return err
			}
			if len(scope.patterns) > 0 {
				scopes = append(scopes, scope)
			}
			return nil
		}

		if language.IsFileAllowed(d.Name(), opts.AllowedLanguages, opts.AllowedFileNames) {
			found = true
			lang := language.GetMarkdownLanguage(d.Name(), opts.AllowedFileNames)
			return writeMarkdown(path, relPath, output, lang, opts.MaxFileSize)
		}

		return nil
	})

	if ret != nil {
		return ret
	}

	if !found {
		return errors.New("no files processed - file list is empty")
	}

	return nil
}

func isIgnored(rel string, isDir bool, opts Options, scopes []gitignoreScope) (ignored, byDefaultPattern bool) {
	if r := patternMatcher.MatchPath(rel, isDir, opts.UserIgnorePatterns); r != patternMatcher.NoMatch {
		return r == patternMatcher.Ignored, false
	}

	for i := len(scopes) - 1; i >= 0; i-- {
		sub, ok := pathWithinScope(rel, scopes[i].dir)
		if !ok {
			continue
		}
		if r := patternMatcher.MatchPath(sub, isDir, scopes[i].patterns); r != patternMatcher.NoMatch {
			return r == patternMatcher.Ignored, false
		}
	}

	return patternMatcher.IsPathIgnored(rel, isDir, opts.DefaultIgnorePatterns), true
}

func pathWithinScope(rel, scopeDir string) (string, bool) {
	if scopeDir == "." {
		return rel, true
	}
	if strings.HasPrefix(rel, scopeDir+"/") {
		return rel[len(scopeDir)+1:], true
	}
	return "", false
}

func loadGitignoreScope(dir, rel string) (gitignoreScope, error) {
	patterns, err := loadGitignorePatterns(filepath.Join(dir, ".gitignore"))
	if err != nil {
		return gitignoreScope{}, err
	}
	return gitignoreScope{dir: rel, patterns: patternMatcher.CompilePatterns(patterns)}, nil
}

func loadGitignorePatterns(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}

func writeMarkdown(path string, displayPath string, output io.Writer, lang string, maxFileSize int64) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stating file %s: %w", path, err)
	}

	if fileInfo.Size() > maxFileSize {
		fmt.Fprintf(os.Stderr, "Warning: skipping large file %s (%d bytes)\n", displayPath, fileInfo.Size())
		return nil
	}

	if _, err := io.WriteString(output, "# "+displayPath+"\n"); err != nil {
		return fmt.Errorf("writing header for %s: %w", path, err)
	}

	if lang != "md" {
		if _, err := io.WriteString(output, "```"+lang+"\n"); err != nil {
			return fmt.Errorf("writing header for %s: %w", path, err)
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("reading file %s: %w", path, err)
	}
	defer file.Close()

	if _, err := io.Copy(output, file); err != nil {
		return fmt.Errorf("writing content from %s: %w", path, err)
	}

	suffix := ""
	if lang != "md" {
		needsNewline := true
		if fileInfo.Size() > 0 {
			lastByte := make([]byte, 1)
			if _, err := file.ReadAt(lastByte, fileInfo.Size()-1); err == nil {
				needsNewline = lastByte[0] != '\n'
			}
		}
		if needsNewline {
			suffix = "\n"
		}
		suffix += "```"
	}
	suffix += "\n\n"

	if _, err := io.WriteString(output, suffix); err != nil {
		return fmt.Errorf("writing suffix for %s: %w", path, err)
	}

	return nil
}
