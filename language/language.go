package language

import (
	"path/filepath"
	"strings"
)

var SupportedLanguages = map[string]bool{
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

var specialFileLanguages = map[string]string{
	"go.mod": "go",
}

func ParseLanguages(languages string) map[string]bool {
	result := make(map[string]bool)

	if languages == "" {
		for lang, defaultEnabled := range SupportedLanguages {
			result[lang] = defaultEnabled
		}
		return result
	}

	for lang := range SupportedLanguages {
		result[lang] = false
	}

	selectedLanguages := strings.Split(languages, ",")
	for _, lang := range selectedLanguages {
		lang = strings.TrimSpace(lang)
		if !strings.HasPrefix(lang, ".") {
			lang = "." + lang
		}
		lang = strings.ToLower(lang)

		if _, exists := SupportedLanguages[lang]; exists {
			result[lang] = true
		}
	}

	return result
}

func GetAllowedFileNames(allowedLanguages map[string]bool) map[string]bool {
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

func GetMarkdownLanguage(filename string, allowedFileNames map[string]bool) string {
	lang := strings.TrimPrefix(filepath.Ext(filename), ".")
	if allowedFileNames[filename] && specialFileLanguages[filename] != "" {
		lang = specialFileLanguages[filename]
	}
	if lang == "" {
		lang = "plaintext"
	}
	return lang
}

func IsFileAllowed(filename string, allowedLanguages, allowedFileNames map[string]bool) bool {
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

func GetDefaultLanguages() []string {
	var defaults []string
	for lang, defaultEnabled := range SupportedLanguages {
		if defaultEnabled {
			defaults = append(defaults, strings.TrimPrefix(lang, "."))
		}
	}
	return defaults
}

func GetSupportedLanguages() []string {
	var languages []string
	for lang := range SupportedLanguages {
		languages = append(languages, strings.TrimPrefix(lang, "."))
	}
	return languages
}
