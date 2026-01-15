package main

import (
	"code2md/c2mConfig"
	"code2md/language"
	"code2md/patternMatcher"
	"code2md/processor"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
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

	err = processor.ProcessDirectory(
		processor.Options{
			InputFolder:      config.InputFolder,
			AllowedLanguages: config.AllowedLanguages,
			AllowedFileNames: config.AllowedFileNames,
			IgnorePatterns:   patternMatcher.CompilePatterns(config.IgnorePatterns),
			MaxFileSize:      config.MaxFileSize,
		}, outputWriter,
	)
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
		activeLangs := language.GetActiveLanguages(config.AllowedLanguages)
		inactiveLangs := language.GetInactiveLanguages(config.AllowedLanguages)
		fmt.Printf("By default, these languages are activated: %v\n", activeLangs)
		fmt.Printf("Supported languages that need to be activated manually: %v\n", inactiveLangs)
	}
}
