package main

import (
	"code2md/c2mConfig"
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

	if config.Help {
		c2mConfig.PrintUsage(os.Stdout)
		return
	}
	if config.InputFolder == "" {
		fmt.Fprint(os.Stderr, "Error: provide an input directory with -i or --input\n\n")
		c2mConfig.PrintUsage(os.Stderr)
		os.Exit(1)
	}

	if err := run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(config *c2mConfig.Config) error {
	var err error
	outputWriter := os.Stdout
	absOutputPath := ""

	if config.OutputMarkdown != "" {
		absOutputPath, err = filepath.Abs(config.OutputMarkdown)
		if err != nil {
			return fmt.Errorf("resolving output path %s: %w", config.OutputMarkdown, err)
		}

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
			InputFolder:           config.InputFolder,
			AllowedLanguages:      config.AllowedLanguages,
			AllowedFileNames:      config.AllowedFileNames,
			UserIgnorePatterns:    patternMatcher.CompilePatterns(config.UserIgnorePatterns),
			DefaultIgnorePatterns: patternMatcher.CompilePatterns(config.DefaultIgnorePatterns),
			AbsOutputFilePath:     absOutputPath,
			MaxFileSize:           config.MaxFileSize,
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
