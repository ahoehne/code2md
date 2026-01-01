package main

import (
	"code2md/c2mConfig"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

var VersionNumber string

func main() {
	config := c2mConfig.InitializeConfigFromFlags()
	if config.Version {
		displayVersion()
		return
	}
	if !c2mConfig.IsConfigValid(config) || config.Help {
		displayUsageInstructions(c2mConfig.GetActiveLanguages(), c2mConfig.GetInactiveLanguages(), !config.Help)
		return
	}

	if err := run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(config c2mConfig.Config) error {
	outputFile, err := os.Create(config.OutputMarkdown)
	if err != nil {
		return fmt.Errorf("creating output file %s: %w", config.OutputMarkdown, err)
	}
	defer func() {
		if closeErr := outputFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close output file: %v\n", closeErr)
		}
	}()

	err = processDirectory(config.InputFolder, outputFile, config.IgnorePatterns, c2mConfig.GetAllowedLanguages(), config.AllowedFileNames)
	if err != nil {
		return fmt.Errorf("processing directory %s: %w", config.InputFolder, err)
	}

	return nil
}

func displayVersion() {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok && buildInfo.GoVersion == "" {
		println("error determining go version")
		return
	}
	if ok && VersionNumber == "" {
		fmt.Println("code2md development-version")
		fmt.Print(buildInfo)
		return
	}
	fmt.Printf("code2md %s\n", VersionNumber)
	println(buildInfo.GoVersion)

}

func displayUsageInstructions(activeLangs, inactiveLangs []string, showError bool) {
	if showError {
		fmt.Println("Error: You must provide both an input folder and an output file.")
	}
	fmt.Println("Usage: code2md -i <input_folder> -o <output_markdown> [--languages <languages>] [--ignore <ignore_patterns>]")
	fmt.Printf("By default, these languages are activated: %v\n", activeLangs)
	fmt.Printf("Supported languages that need to be activated manually: %v\n", inactiveLangs)
}

func processDirectory(inputFolder string, outputFile *os.File, patterns []string, allowedLanguages, allowedFileNames map[string]bool) error {
	specialFileLanguages := map[string]string{
		"go.mod": "go",
	}
	return filepath.WalkDir(inputFolder, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(inputFolder, path)
		if err != nil {
			return err
		}
		if isPathIgnored(relPath, patterns) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() && isFileAllowed(d.Name(), allowedLanguages, allowedFileNames) {
			return writeMarkdown(path, outputFile, getMdLang(d.Name(), allowedFileNames, specialFileLanguages))
		}
		return nil
	})
}

func writeMarkdown(path string, outputFile *os.File, lang string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	title := fmt.Sprintf("# %s\n\n```%s\n", path, lang)
	outputFile.WriteString(title)
	outputFile.Write(content)
	outputFile.WriteString("\n```\n\n")
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

func doesPathMatchPattern(path, pattern string) bool {
	if pattern == path {
		return true
	}
	if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
		return strings.HasPrefix(path, strings.TrimPrefix(pattern, "/"))
	}
	if strings.HasPrefix(pattern, "**") {
		// Handle globstar matching for the start of the pattern (e.g., **.min.css)
		// TODO: check out https://github.com/bmatcuk/doublestar
		subPattern := strings.TrimPrefix(pattern, "**/")
		if subPattern == pattern {
			subPattern = strings.TrimPrefix(pattern, "**")
			_, fileName := filepath.Split(path)
			matched, _ := filepath.Match("*"+subPattern, fileName)
			return matched
		}

		parts := strings.Split(path, string(filepath.Separator))
		for i := 0; i < len(parts); i++ {
			remainingPath := strings.Join(parts[i:], string(filepath.Separator))
			matched, _ := filepath.Match(subPattern, remainingPath)
			if matched {
				return true
			}
		}
		return false
	}
	if strings.Contains(pattern, "*") {
		matched, _ := filepath.Match(pattern, path)
		return matched
	}
	return false
}

func isPathIgnored(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.HasSuffix(pattern, "/") && strings.HasPrefix(path, pattern) {
			return true
		}
		if doesPathMatchPattern(path, pattern) {
			return true
		}
	}
	return false
}

func isFileAllowed(filename string, allowedLanguages, allowedFileNames map[string]bool) bool {
	return allowedFileNames[filename] || allowedLanguages[filepath.Ext(filename)]
}
