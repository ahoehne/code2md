package c2mConfig

import (
	"bufio"
	"flag"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultIgnoredPatterns = "*.yaml,*.yml,*.xml"
	defaultMaxFileSize     = 100 * 1024 * 1024
)

type Config struct {
	InputFolder      string
	OutputMarkdown   string
	AllowedLanguages map[string]bool
	AllowedFileNames map[string]bool
	IgnorePatterns   []string
	MaxFileSize      int64
	Help             bool
	Version          bool
}

var supportedLanguages = map[string]bool{
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

func InitializeConfigFromFlags() (*Config, error) {
	inputFolder := flag.String("input", "", "Input folder to scan")
	outputMarkdown := flag.String("output", "", "Output Markdown file")
	languages := flag.String("languages", "", "Comma-separated list of allowed languages (empty = use defaults)")
	ignorePatterns := flag.String("ignore", "", "Comma-separated list of files and/or search patterns to ignore")
	maxFileSize := flag.Int64("max-file-size", defaultMaxFileSize, "Maximum file size in bytes to process")
	help := flag.Bool("help", false, "Show help")
	v := flag.Bool("version", false, "Show version information")

	flag.StringVar(inputFolder, "i", "", "Input folder to scan (shorthand)")
	flag.StringVar(outputMarkdown, "o", "", "Output Markdown file (shorthand)")
	flag.StringVar(languages, "l", "", "languages (shorthand)")
	flag.StringVar(ignorePatterns, "I", defaultIgnoredPatterns, "ignore patterns (shorthand)")
	flag.Int64Var(maxFileSize, "m", defaultMaxFileSize, "max file size (shorthand)")
	flag.BoolVar(help, "h", false, "help (shorthand)")
	flag.BoolVar(v, "v", false, "version (shorthand)")

	flag.Parse()

	allowedLanguages := parseLanguages(*languages)

	var ignorePatternsList []string
	if *ignorePatterns == "" {
		ignorePatternsList = []string{*outputMarkdown}
	} else {
		ignorePatternsList = append(strings.Split(*ignorePatterns, ","), *outputMarkdown)
	}

	if allowedLanguages[".css"] || allowedLanguages[".scss"] {
		ignorePatternsList = append(ignorePatternsList, "**.min.css")
	}

	gitignorePatterns, err := LoadGitignorePatterns(filepath.Join(".", ".gitignore"))
	if err != nil {
		return nil, err
	}
	ignorePatternsList = append(gitignorePatterns, ignorePatternsList...)

	return &Config{
		InputFolder:      *inputFolder,
		OutputMarkdown:   *outputMarkdown,
		AllowedLanguages: allowedLanguages,
		AllowedFileNames: getMapOfAllowedFileNames(allowedLanguages),
		IgnorePatterns:   ignorePatternsList,
		MaxFileSize:      *maxFileSize,
		Help:             *help,
		Version:          *v,
	}, nil
}

func parseLanguages(languages string) map[string]bool {
	result := make(map[string]bool)

	if languages == "" {
		for lang, defaultEnabled := range supportedLanguages {
			result[lang] = defaultEnabled
		}
		return result
	}

	for lang := range supportedLanguages {
		result[lang] = false
	}

	selectedLanguages := strings.Split(languages, ",")
	for _, lang := range selectedLanguages {
		lang = strings.TrimSpace(lang)
		if !strings.HasPrefix(lang, ".") {
			lang = "." + lang
		}
		lang = strings.ToLower(lang)

		if _, exists := supportedLanguages[lang]; exists {
			result[lang] = true
		}
	}

	return result
}

func IsConfigValid(config *Config) bool {
	return config != nil && config.InputFolder != ""
}

func getMapOfAllowedFileNames(allowedLanguages map[string]bool) map[string]bool {
	allowedFileNames := make(map[string]bool)
	if allowedLanguages[".go"] {
		allowedFileNames["go.mod"] = true
	}
	if allowedLanguages[".php"] {
		allowedFileNames["composer.json"] = true
	}
	if allowedLanguages[".js"] || allowedLanguages[".ts"] {
		allowedFileNames["package.json"] = true
	}
	if allowedLanguages[".ts"] {
		allowedFileNames["tsconfig.json"] = true
	}
	if allowedLanguages[".java"] {
		allowedFileNames["pom.xml"] = true
	}
	return allowedFileNames
}

func LoadGitignorePatterns(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
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

func GetActiveLanguages(config *Config) []string {
	if config == nil {
		return []string{}
	}

	var active []string
	for lang, enabled := range config.AllowedLanguages {
		if enabled {
			active = append(active, strings.TrimPrefix(lang, "."))
		}
	}
	return active
}

func GetInactiveLanguages(config *Config) []string {
	if config == nil {
		return []string{}
	}

	var inactive []string
	for lang, enabled := range config.AllowedLanguages {
		if !enabled {
			inactive = append(inactive, strings.TrimPrefix(lang, "."))
		}
	}
	return inactive
}

func GetDefaultLanguages() []string {
	var defaults []string
	for lang, defaultEnabled := range supportedLanguages {
		if defaultEnabled {
			defaults = append(defaults, strings.TrimPrefix(lang, "."))
		}
	}
	return defaults
}

func GetSupportedLanguages() []string {
	var languages []string
	for lang := range supportedLanguages {
		languages = append(languages, strings.TrimPrefix(lang, "."))
	}
	return languages
}
