package c2mConfig

import (
	"code2md/language"
	"flag"
	"fmt"
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
	AllowedFileNames      map[string]bool
	UserIgnorePatterns    []string
	DefaultIgnorePatterns []string
	MaxFileSize           int64
	Help                  bool
	Version               bool
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

	if flag.NArg() > 0 {
		return nil, fmt.Errorf("unexpected arguments: %s", strings.Join(flag.Args(), " "))
	}

	ignoreExplicitlySet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "ignore" || f.Name == "I" {
			ignoreExplicitlySet = true
		}
	})

	allowedLanguages := language.ParseLanguages(*languages)

	var userIgnoreList []string
	var defaultIgnoreList []string
	for _, p := range strings.Split(ignorePatterns, ",") {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		if ignoreExplicitlySet {
			userIgnoreList = append(userIgnoreList, trimmed)
		} else if !isExtensionPatternForEnabledLanguage(trimmed, allowedLanguages) {
			defaultIgnoreList = append(defaultIgnoreList, trimmed)
		}
	}

	if allowedLanguages[".css"] || allowedLanguages[".scss"] {
		defaultIgnoreList = append(defaultIgnoreList, "**.min.css")
	}

	return &Config{
		InputFolder:           *inputFolder,
		OutputMarkdown:        *outputMarkdown,
		AllowedLanguages:      allowedLanguages,
		AllowedFileNames:      language.GetAllowedFileNames(allowedLanguages),
		UserIgnorePatterns:    userIgnoreList,
		DefaultIgnorePatterns: defaultIgnoreList,
		MaxFileSize:           *maxFileSize,
		Help:                  *help,
		Version:               *v,
	}, nil
}

func isExtensionPatternForEnabledLanguage(pattern string, allowedLanguages map[string]bool) bool {
	if !strings.HasPrefix(pattern, "*.") {
		return false
	}
	return allowedLanguages[strings.TrimPrefix(pattern, "*")]
}

func IsConfigValid(config *Config) bool {
	return config != nil && config.InputFolder != "" && config.MaxFileSize > 0
}
