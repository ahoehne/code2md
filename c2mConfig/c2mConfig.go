package c2mConfig

import (
	"code2md/language"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

var flagAliases = []struct{ name, short string }{
	{"input", "i"},
	{"output", "o"},
	{"languages", "l"},
	{"ignore", "I"},
	{"max-file-size", "m"},
	{"help", "h"},
	{"version", "v"},
}

const (
	defaultIgnoredPatterns = "*.yaml,*.yml,*.xml"
	defaultMaxFileSize     = 100 * 1024 * 1024
)

type sizeValue int64

func (s *sizeValue) String() string { return strconv.FormatInt(int64(*s), 10) }

func (s *sizeValue) Set(value string) error {
	number := strings.ToUpper(value)
	multiplier := int64(1)
	for _, unit := range []struct {
		suffix     string
		multiplier int64
	}{{"KB", 1024}, {"MB", 1024 * 1024}, {"GB", 1024 * 1024 * 1024}, {"B", 1}} {
		if strings.HasSuffix(number, unit.suffix) {
			number = strings.TrimSuffix(number, unit.suffix)
			multiplier = unit.multiplier
			break
		}
	}
	n, err := strconv.ParseInt(number, 10, 64)
	if err != nil || n <= 0 || n > math.MaxInt64/multiplier {
		return fmt.Errorf("expected a positive size like 512KB or 10MB, got %q", value)
	}
	*s = sizeValue(n * multiplier)
	return nil
}

type Config struct {
	InputFolder           string
	OutputMarkdown        string
	AllowedLanguages      map[string]bool
	AllowedFileNames      map[string]bool
	UserIgnorePatterns    []string
	DefaultIgnorePatterns []string
	MaxFileSize           int64
	Help                  bool
	Version               bool
}

func InitializeConfigFromFlags() (*Config, error) {
	inputFolder := flag.String("input", "", "Input `directory` to scan (required)")
	outputMarkdown := flag.String("output", "", "Output Markdown `file` (default: stdout)")
	languages := flag.String("languages", "", "Comma-separated language `names` or extensions (empty uses defaults)")
	var ignorePatterns string
	flag.StringVar(&ignorePatterns, "ignore", defaultIgnoredPatterns, "Comma-separated ignore `patterns` (replaces defaults: "+defaultIgnoredPatterns+")")
	maxFileSize := sizeValue(defaultMaxFileSize)
	flag.Var(&maxFileSize, "max-file-size", "Maximum `size` of each file, e.g. 512KB or 10MB (default: 100MB)")
	help := flag.Bool("help", false, "Show help")
	v := flag.Bool("version", false, "Show version information")

	for _, alias := range flagAliases {
		f := flag.Lookup(alias.name)
		flag.Var(f.Value, alias.short, f.Usage)
	}
	flag.CommandLine.Usage = func() { PrintUsage(flag.CommandLine.Output()) }
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

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
		MaxFileSize:           int64(maxFileSize),
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

func PrintUsage(output io.Writer) {
	fmt.Fprintln(output, "Usage: code2md -i <directory> [options]")
	fmt.Fprintln(output, "\nOptions:")
	w := tabwriter.NewWriter(output, 0, 4, 2, ' ', 0)
	for _, alias := range flagAliases {
		f := flag.Lookup(alias.name)
		valueName, usage := flag.UnquoteUsage(f)
		label := fmt.Sprintf("  -%s, --%s", alias.short, alias.name)
		if valueName != "" {
			label += " " + valueName
		}
		fmt.Fprintf(w, "%s\t%s\n", label, usage)
	}
	w.Flush()
	defaults := language.ParseLanguages("")
	fmt.Fprintf(output, "\nDefault languages: %s\n", strings.Join(language.LanguageNames(defaults, true), ", "))
	fmt.Fprintf(output, "Opt-in languages (-l): %s\n", strings.Join(language.LanguageNames(defaults, false), ", "))
}
