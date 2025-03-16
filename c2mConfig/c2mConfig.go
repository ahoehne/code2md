package c2mConfig

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultAllowedLanguages = ".php,.go,.js,.ts,.py,.sh"
	defaultIgnoredPatterns  = "*.yaml,*.yml"
)

type Config struct {
	InputFolder      string
	OutputMarkdown   string
	AllowedFileNames map[string]bool
	IgnorePatterns   []string
}

var AllowedLanguages = map[string]bool{
	".php":  true,
	".go":   true,
	".js":   true,
	".ts":   true,
	".py":   true,
	".sh":   true,
	".json": false,
	".yaml": false,
	".yml":  false,
}

func InitializeConfigFromFlags() Config {
	inputFolder := flag.String("input", "", "Input folder to scan")
	outputMarkdown := flag.String("output", "", "Output Markdown file")
	languages := flag.String("languages", "", "Comma-separated list of allowed languages")
	ignorePatterns := flag.String("ignore", "", "Comma-separated list of files and/or search patterns to ignore")

	flag.StringVar(inputFolder, "i", "", "Input folder to scan (shorthand)")
	flag.StringVar(outputMarkdown, "o", "", "Output Markdown file (shorthand)")
	flag.StringVar(languages, "l", defaultAllowedLanguages, "languages (shorthand)")
	flag.StringVar(ignorePatterns, "I", defaultIgnoredPatterns, "ignore patterns (shorthand)")

	flag.Parse()

	updateLanguagesFilter(*languages)
	ignorePatternsList := parseIgnorePatterns(*ignorePatterns)

	return Config{
		InputFolder:      *inputFolder,
		OutputMarkdown:   *outputMarkdown,
		AllowedFileNames: fetchAllowedFileNames(AllowedLanguages),
		IgnorePatterns:   ignorePatternsList,
	}
}

func IsConfigValid(config Config) bool {
	return config.InputFolder != "" && config.OutputMarkdown != ""
}

func updateLanguagesFilter(languages string) {
	selectedLanguages := strings.Split(languages, ",")
	for ext := range AllowedLanguages {
		AllowedLanguages[ext] = sliceContains(selectedLanguages, ext)
	}
}

func fetchAllowedFileNames(AllowedLanguages map[string]bool) map[string]bool {
	allowedFileNames := make(map[string]bool)
	if AllowedLanguages[".go"] {
		allowedFileNames["go.mod"] = true
	}
	if AllowedLanguages[".php"] {
		allowedFileNames["composer.json"] = true
	}
	if AllowedLanguages[".js"] {
		allowedFileNames["package.json"] = true
	}
	return allowedFileNames
}

func sliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func parseIgnorePatterns(patterns string) []string {
	ignorePatternsList := strings.Split(patterns, ",")
	if len(ignorePatternsList) == 1 && ignorePatternsList[0] == "" {
		return []string{}
	}
	gitignorePatterns, err := LoadGitignorePatterns(filepath.Join(".", ".gitignore"))
	if err != nil {
		fmt.Println("unknown error with loading .gitignore: ", err)
		return ignorePatternsList
	}
	return append(gitignorePatterns, ignorePatternsList...)
}

func LoadGitignorePatterns(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("error reading .gitignore file: %w", err)
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" && !strings.HasPrefix(line, "#") {
			patterns = append(patterns, line)
		}
	}
	return patterns, scanner.Err()
}

func GetActiveLanguages() []string {
	var active []string
	for lang, v := range AllowedLanguages {
		if v {
			active = append(active, strings.TrimPrefix(lang, "."))
		}
	}
	return active
}

func GetInactiveLanguages() []string {
	var inactive []string
	for lang, v := range AllowedLanguages {
		if !v {
			inactive = append(inactive, strings.TrimPrefix(lang, "."))
		}
	}
	return inactive
}
