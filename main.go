package main

import (
	"code2md/c2mConfig"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	config := c2mConfig.InitializeConfigFromFlags()
	if !c2mConfig.IsConfigValid(config) {
		displayUsageInstructions(c2mConfig.GetActiveLanguages(), c2mConfig.GetInactiveLanguages())
		return
	}

	outputFile, err := os.Create(config.OutputMarkdown)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer outputFile.Close()

	err = processDirectory(config.InputFolder, outputFile, config.IgnorePatterns, c2mConfig.AllowedLanguages, config.AllowedFileNames)
	if err != nil {
		fmt.Printf("Error walking through the directory: %v\n", err)
	}
}

func displayUsageInstructions(activeLangs, inactiveLangs []string) {
	fmt.Println("Error: You must provide both an input folder and an output file.")
	fmt.Println("Usage: go run script.go -i <input_folder> -o <output_markdown> [--languages <languages>] [--ignore <ignore_patterns>]")
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
