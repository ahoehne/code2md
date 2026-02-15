package processor

import (
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
	AllowedFileNames map[string]bool
	IgnorePatterns   []patternMatcher.CompiledPattern
	MaxFileSize      int64
}

func ProcessDirectory(opts Options, output io.Writer) error {
	found := false

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

		if patternMatcher.IsPathIgnored(relPath, opts.IgnorePatterns) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() && language.IsFileAllowed(d.Name(), opts.AllowedLanguages, opts.AllowedFileNames) {
			found = true
			lang := language.GetMarkdownLanguage(d.Name(), opts.AllowedFileNames)
			return WriteMarkdown(path, relPath, output, lang, opts.MaxFileSize)
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

func WriteMarkdown(path string, displayPath string, output io.Writer, lang string, maxFileSize int64) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stating file %s: %w", path, err)
	}

	if fileInfo.Size() > maxFileSize {
		fmt.Fprintf(os.Stderr, "Warning: skipping large file %s (%d bytes)\n", displayPath, fileInfo.Size())
		return nil
	}

	var buf strings.Builder
	buf.WriteString("# ")
	buf.WriteString(displayPath)
	buf.WriteString("\n")

	if lang != "md" {
		buf.WriteString("```" + lang + "\n")
	}

	if _, err := io.WriteString(output, buf.String()); err != nil {
		return fmt.Errorf("writing header for %s: %w", path, err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file %s: %w", path, err)
	}

	if _, err := output.Write(content); err != nil {
		return fmt.Errorf("writing content from %s: %w", path, err)
	}

	suffix := ""
	if lang != "md" {
		if len(content) == 0 || content[len(content)-1] != '\n' {
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
