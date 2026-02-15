package c2mConfig

import (
	"bufio"
	"code2md/language"
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

func InitializeConfigFromFlags() (*Config, error) {
	inputFolder := flag.String("input", "", "Input folder to scan")
	outputMarkdown := flag.String("output", "", "Output Markdown file")
	languages := flag.String("languages", "", "Comma-separated list of allowed languages (empty = use defaults)")
	var ignorePatterns string
	flag.StringVar(&ignorePatterns, "ignore", defaultIgnoredPatterns, "Comma-separated list of files and/or search patterns to ignore")
	maxFileSize := flag.Int64("max-file-size", defaultMaxFileSize, "Maximum file size in bytes to process")
	help := flag.Bool("help", false, "Show help")
	v := flag.Bool("version", false, "Show version information")

	flag.StringVar(inputFolder, "i", "", "Input folder to scan (shorthand)")
	flag.StringVar(outputMarkdown, "o", "", "Output Markdown file (shorthand)")
	flag.StringVar(languages, "l", "", "languages (shorthand)")
	flag.StringVar(&ignorePatterns, "I", defaultIgnoredPatterns, "ignore patterns (shorthand)")
	flag.Int64Var(maxFileSize, "m", defaultMaxFileSize, "max file size (shorthand)")
	flag.BoolVar(help, "h", false, "help (shorthand)")
	flag.BoolVar(v, "v", false, "version (shorthand)")

	flag.Parse()

	allowedLanguages := language.ParseLanguages(*languages)

	var ignorePatternsList []string
	for _, p := range strings.Split(ignorePatterns, ",") {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			ignorePatternsList = append(ignorePatternsList, trimmed)
		}
	}
	if *outputMarkdown != "" {
		ignorePatternsList = append(ignorePatternsList, *outputMarkdown)
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
		AllowedFileNames: language.GetAllowedFileNames(allowedLanguages),
		IgnorePatterns:   ignorePatternsList,
		MaxFileSize:      *maxFileSize,
		Help:             *help,
		Version:          *v,
	}, nil
}

func IsConfigValid(config *Config) bool {
	return config != nil && config.InputFolder != "" && config.MaxFileSize > 0
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
