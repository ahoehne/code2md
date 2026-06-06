package language

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var supportedLanguages = map[string]bool{
	".php":        true,
	".go":         true,
	".js":         true,
	".ts":         true,
	".jsx":        true,
	".tsx":        true,
	".py":         true,
	".sh":         true,
	".java":       true,
	".c":          true,
	".h":          true,
	".cpp":        true,
	".cc":         true,
	".cxx":        true,
	".hpp":        true,
	".cs":         true,
	".rs":         true,
	".dockerfile": true,
	".md":         false,
	".html":       false,
	".scss":       false,
	".css":        false,
	".json":       false,
	".yaml":       false,
	".yml":        false,
	".xml":        false,
	".toml":       false,
	".sql":        false,
}

var specialFileLanguages = map[string]string{
	"go.mod":     "go",
	"Pipfile":    "toml",
	"Cargo.toml": "toml",
}

var extensionMarkdownLanguages = map[string]string{
	".h":   "c",
	".hpp": "cpp",
	".cc":  "cpp",
	".cxx": "cpp",
}

var languageManifests = map[string][]string{
	".php":  {"composer.json"},
	".go":   {"go.mod"},
	".js":   {"package.json"},
	".ts":   {"package.json", "tsconfig.json"},
	".py":   {"pyproject.toml", "requirements.txt", "setup.py", "setup.cfg", "Pipfile"},
	".java": {"pom.xml", "build.gradle", "settings.gradle"},
	".rs":   {"Cargo.toml"},
	".jsx":  {"package.json"},
	".tsx":  {"package.json", "tsconfig.json"},
}

func isDockerfile(filename string) bool {
	lower := strings.ToLower(filename)
	return lower == "dockerfile" || strings.HasPrefix(lower, "dockerfile.")
}

func ParseLanguages(languages string) map[string]bool {
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
		} else {
			fmt.Fprintf(os.Stderr, "Warning: unrecognized language %q, skipping\n", strings.TrimPrefix(lang, "."))
		}
	}

	return result
}

func GetAllowedFileNames(allowedLanguages map[string]bool) map[string]bool {
	allowedFileNames := make(map[string]bool)
	for ext, files := range languageManifests {
		if allowedLanguages[ext] {
			for _, f := range files {
				allowedFileNames[f] = true
			}
		}
	}
	return allowedFileNames
}

func GetMarkdownLanguage(filename string, allowedFileNames map[string]bool) string {
	if isDockerfile(filename) {
		return "dockerfile"
	}
	if lang, exists := specialFileLanguages[filename]; exists && allowedFileNames[filename] {
		return lang
	}
	if lang, exists := extensionMarkdownLanguages[filepath.Ext(filename)]; exists {
		return lang
	}
	lang := strings.TrimPrefix(filepath.Ext(filename), ".")
	if lang == "" {
		return "plaintext"
	}
	return lang
}

func IsFileAllowed(filename string, allowedLanguages, allowedFileNames map[string]bool) bool {
	if allowedLanguages[".dockerfile"] && isDockerfile(filename) {
		return true
	}
	return allowedFileNames[filename] || allowedLanguages[filepath.Ext(filename)]
}

func GetActiveLanguages(allowedLanguages map[string]bool) []string {
	var active []string
	for lang, enabled := range allowedLanguages {
		if enabled {
			active = append(active, strings.TrimPrefix(lang, "."))
		}
	}
	return active
}

func GetInactiveLanguages(allowedLanguages map[string]bool) []string {
	var inactive []string
	for lang, enabled := range allowedLanguages {
		if !enabled {
			inactive = append(inactive, strings.TrimPrefix(lang, "."))
		}
	}
	return inactive
}

