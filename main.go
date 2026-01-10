package main

import (
	"code2md/c2mConfig"
	"code2md/patternMatcher"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

var VersionNumber string

func main() {
	config, err := c2mConfig.InitializeConfigFromFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
		os.Exit(1)
	}

	if config.Version {
		displayVersion()
		return
	}

	if !c2mConfig.IsConfigValid(config) || config.Help {
		displayUsageInstructions(config, !config.Help)
		return
	}

	if err := run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(config *c2mConfig.Config) error {
	var err error
	outputWriter := os.Stdout

	if config.OutputMarkdown != "" {
		outputDir := filepath.Dir(config.OutputMarkdown)
		if outputDir != "." && outputDir != "" {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("creating output directory: %w", err)
			}
		}

		outputWriter, err = os.Create(config.OutputMarkdown)
		if err != nil {
			return fmt.Errorf("creating output file %s: %w", config.OutputMarkdown, err)
		}
		defer func() {
			if closeErr := outputWriter.Close(); closeErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to close output file: %v\n", closeErr)
			}
		}()
	}

	compiledPatterns := patternMatcher.CompilePatterns(config.IgnorePatterns)

	err = processDirectory(config.InputFolder, outputWriter, compiledPatterns, config.AllowedLanguages, config.AllowedFileNames, config.MaxFileSize)
	if err != nil {
		return fmt.Errorf("processing directory %s: %w", config.InputFolder, err)
	}

	return nil
}

func displayVersion() {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok || buildInfo.GoVersion == "" {
		fmt.Println("Error determining Go version")
		return
	}

	if VersionNumber == "" {
		fmt.Println("code2md development-version")
		fmt.Println(buildInfo.GoVersion)
		return
	}

	fmt.Printf("code2md %s\n", VersionNumber)
	fmt.Println(buildInfo.GoVersion)
}

func displayUsageInstructions(config *c2mConfig.Config, showError bool) {
	if showError {
		fmt.Println("Error: You have to provide an input folder.")
	}
	fmt.Println("Usage: code2md -i <input_folder> -o <output_markdown> [--languages <languages>] [--ignore <ignore_patterns>]")
	fmt.Println("| Flag              | Short | Description                                                     |")
	fmt.Println("| ----------------- | ----- | --------------------------------------------------------------- |")
	fmt.Println("| `--input`         | `-i`  | Input directory to scan (required)                              |")
	fmt.Println("| `--output`        | `-o`  | Output Markdown file (optional, defaults to stdout)             |")
	fmt.Println("| `--languages`     | `-l`  | Comma-separated list of allowed languages (extensions or names) |")
	fmt.Println("| `--ignore`        | `-I`  | Comma-separated ignore patterns                                 |")
	fmt.Println("| `--max-file-size` | `-m`  | Maximum file size in bytes (default: 100MB)                     |")
	fmt.Println("| `--help`          | `-h`  | Show help                                                       |")
	fmt.Println("| `--version`       | `-v`  | Show version information                                        |")

	if config != nil {
		activeLangs := c2mConfig.GetActiveLanguages(config)
		inactiveLangs := c2mConfig.GetInactiveLanguages(config)
		fmt.Printf("By default, these languages are activated: %v\n", activeLangs)
		fmt.Printf("Supported languages that need to be activated manually: %v\n", inactiveLangs)
	}
}

func processDirectory(inputFolder string, outputFile *os.File, patterns []patternMatcher.CompiledPattern, allowedLanguages, allowedFileNames map[string]bool, maxFileSize int64) error {
	specialFileLanguages := map[string]string{
		"go.mod": "go",
	}
	processed := 0

	ret := filepath.WalkDir(inputFolder, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				fmt.Fprintf(os.Stderr, "Warning: permission denied: %s\n", path)
				return nil
			}
			return fmt.Errorf("accessing path %s: %w", path, err)
		}

		relPath, err := filepath.Rel(inputFolder, path)
		if err != nil {
			return fmt.Errorf("getting relative path for %s: %w", path, err)
		}

		if patternMatcher.IsPathIgnored(relPath, patterns) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() && isFileAllowed(d.Name(), allowedLanguages, allowedFileNames) {
			processed++
			return writeMarkdown(path, outputFile, getMdLang(d.Name(), allowedFileNames, specialFileLanguages), maxFileSize)
		}

		return nil
	})

	if ret != nil {
		return ret
	}

	if processed == 0 {
		return errors.New("no files processed - file list is empty")
	}

	return nil
}

func writeMarkdown(path string, outputFile *os.File, lang string, maxFileSize int64) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stating file %s: %w", path, err)
	}

	if fileInfo.Size() > maxFileSize {
		fmt.Fprintf(os.Stderr, "Warning: skipping large file %s (%d bytes)\n", path, fileInfo.Size())
		return nil
	}

	var buf strings.Builder
	buf.WriteString("# ")
	buf.WriteString(path)
	buf.WriteString("\n")

	if lang != "md" {
		buf.WriteString("```" + lang + "\n")
	}

	if _, err := outputFile.WriteString(buf.String()); err != nil {
		return fmt.Errorf("writing header for %s: %w", path, err)
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening file %s: %w", path, err)
	}
	defer file.Close()

	if _, err := io.Copy(outputFile, file); err != nil {
		return fmt.Errorf("copying content from %s: %w", path, err)
	}

	suffix := ""
	if lang != "md" {
		suffix = "\n```"
	}
	suffix += "\n\n"

	if _, err := outputFile.WriteString(suffix); err != nil {
		return fmt.Errorf("writing suffix for %s: %w", path, err)
	}

	return nil
}

func getMdLang(filename string, allowedFileNames map[string]bool, specialFileLanguages map[string]string) string {
	lang := strings.TrimPrefix(filepath.Ext(filename), ".")
	if allowedFileNames[filename] && specialFileLanguages[filename] != "" {
		lang = specialFileLanguages[filename]
	}
	if lang == "" {
		lang = "plaintext"
	}
	return lang
}

func isFileAllowed(filename string, allowedLanguages, allowedFileNames map[string]bool) bool {
	return allowedFileNames[filename] || allowedLanguages[filepath.Ext(filename)]
}
