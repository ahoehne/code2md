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
	defaultAllowedLanguages = ".php,.go,.js,.ts,.py,.sh,.java"
	defaultIgnoredPatterns  = "*.yaml,*.yml,*.xml"
)

type Config struct {
	InputFolder      string
	OutputMarkdown   string
	AllowedFileNames map[string]bool
	IgnorePatterns   []string
	Help             bool
	Version          bool
}

var allowedLanguages = map[string]bool{
	".php":  true,
	".go":   true,
	".js":   true,
	".ts":   true,
	".py":   true,
	".sh":   true,
	".java": true,
	".md":   false,
	".html": false,
	".scss": false,
	".css":  false,
	".json": false,
	".yaml": false,
	".yml":  false,
	".xml":  false,
}

func InitializeConfigFromFlags() Config {
	inputFolder := flag.String("input", "", "Input folder to scan")
	outputMarkdown := flag.String("output", "", "Output Markdown file")
	languages := flag.String("languages", "", "Comma-separated list of allowed languages")
	ignorePatterns := flag.String("ignore", "", "Comma-separated list of files and/or search patterns to ignore")
	help := flag.Bool("help", false, "Show help")
	v := flag.Bool("version", false, "Show version information")

	flag.StringVar(inputFolder, "i", "", "Input folder to scan (shorthand)")
	flag.StringVar(outputMarkdown, "o", "", "Output Markdown file (shorthand)")
	flag.StringVar(languages, "l", defaultAllowedLanguages, "languages (shorthand)")
	flag.StringVar(ignorePatterns, "I", defaultIgnoredPatterns, "ignore patterns (shorthand)")
	flag.BoolVar(help, "h", false, "help (shorthand)")
	flag.BoolVar(v, "v", false, "version (shorthand)")

	flag.Parse()

	updateLanguagesFilter(*languages)
	if *ignorePatterns == "" {
		*ignorePatterns = *outputMarkdown
	} else {
		*ignorePatterns += "," + *outputMarkdown
	}

	ignorePatternsList := parseIgnorePatterns(*ignorePatterns)

	return Config{
		InputFolder:      *inputFolder,
		OutputMarkdown:   *outputMarkdown,
		AllowedFileNames: GetAllowedFileNames(allowedLanguages),
		IgnorePatterns:   ignorePatternsList,
		Help:             *help,
		Version:          *v,
	}
}

func GetAllowedLanguages() map[string]bool {
	result := make(map[string]bool)
	for k, v := range allowedLanguages {
		result[k] = v
	}
	return result
}

func IsConfigValid(config Config) bool {
	return config.InputFolder != ""
}

func updateLanguagesFilter(languages string) {
	selectedLanguages := strings.Split(languages, ",")
	for langKey, langVal := range selectedLanguages {
		if !strings.HasPrefix(langVal, ".") {
			langVal = "." + langVal
			selectedLanguages[langKey] = "." + langVal
		}
		selectedLanguages[langKey] = strings.ToLower(langVal)
	}
	for ext := range allowedLanguages {
		allowedLanguages[ext] = sliceContains(selectedLanguages, ext)
	}
}

func GetAllowedFileNames(allowedLanguages map[string]bool) map[string]bool {
	allowedFileNames := make(map[string]bool)
	if allowedLanguages[".go"] {
		allowedFileNames["go.mod"] = true
	}
	if allowedLanguages[".php"] {
		allowedFileNames["composer.json"] = true
	}
	if allowedLanguages[".js"] {
		allowedFileNames["package.json"] = true
	}
	if allowedLanguages[".java"] {
		allowedFileNames["pom.xml"] = true
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
	if allowedLanguages[".css"] {
		ignorePatternsList = append(ignorePatternsList, "**.min.css")
	} else if allowedLanguages[".scss"] {
		ignorePatternsList = append(ignorePatternsList, "**.css")
	}
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
	for lang, v := range allowedLanguages {
		if v {
			active = append(active, strings.TrimPrefix(lang, "."))
		}
	}
	return active
}

func GetInactiveLanguages() []string {
	var inactive []string
	for lang, v := range allowedLanguages {
		if !v {
			inactive = append(inactive, strings.TrimPrefix(lang, "."))
		}
	}
	return inactive
}
