package language

import (
	"reflect"
	"testing"
)

func langMap(enabled ...string) map[string]bool {
	m := make(map[string]bool)
	for k := range supportedLanguages {
		m[k] = false
	}
	for _, e := range enabled {
		m[e] = true
	}
	return m
}

func TestIsDockerfile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"Dockerfile", true},
		{"dockerfile", true},
		{"DOCKERFILE", true},
		{"DockerFile", true},
		{"Dockerfile.dev", true},
		{"Dockerfile.prod", true},
		{"Dockerfile.test", true},
		{"dockerfile.local", true},
		{"DOCKERFILE.CI", true},
		{"Dockerfile.multi.stage", true},
		{"NotDockerfile", false},
		{"MyDockerfile", false},
		{"Dockerfile-dev", false},
		{"docker-compose.yml", false},
		{".dockerfile", false},
		{"file.dockerfile", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if got := isDockerfile(tt.filename); got != tt.want {
				t.Errorf("isDockerfile(%q) = %v; want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIsFileAllowed(t *testing.T) {
	allowedLanguages := map[string]bool{
		".go":         true,
		".json":       false,
		".dockerfile": true,
	}

	allowedFileNames := map[string]bool{
		"go.mod":  true,
		"pom.xml": true,
	}

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"allowed extension", "file.go", true},
		{"uppercase extension", "MAIN.GO", true},
		{"multi-dot extension", "file.with.multiple.dots.go", true},
		{"hidden file", ".hidden.go", true},
		{"unknown extension", "file.txt", false},
		{"disabled extension", "file.json", false},
		{"uppercase disabled extension", "FILE.JSON", false},
		{"no extension", "file_without_extension", false},
		{"manifest without extension", "go.mod", true},
		{"manifest with disallowed extension", "pom.xml", true},
		{"Dockerfile", "Dockerfile", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFileAllowed(tt.filename, allowedLanguages, allowedFileNames); got != tt.want {
				t.Errorf("IsFileAllowed(%q) = %v; want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIsFileAllowedDockerfileDisabled(t *testing.T) {
	allowedLanguages := map[string]bool{
		".go":         true,
		".dockerfile": false,
	}
	allowedFileNames := map[string]bool{
		"go.mod": true,
	}

	tests := []struct {
		filename string
		want     bool
	}{
		{"Dockerfile", false},
		{"dockerfile", false},
		{"Dockerfile.dev", false},
		{"main.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if got := IsFileAllowed(tt.filename, allowedLanguages, allowedFileNames); got != tt.want {
				t.Errorf("IsFileAllowed(%q) = %v; want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestGetMarkdownLanguage(t *testing.T) {
	allowedFileNames := map[string]bool{
		"go.mod":     true,
		"Pipfile":    true,
		"Cargo.toml": true,
	}

	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{"extension fallback", "main.go", "go"},
		{"uppercase extension", "SCRIPT.PY", "py"},
		{"multi-dot extension", "file.test.js", "js"},
		{"hidden file with extension", ".eslintrc.json", "json"},
		{"no extension", "README", "plaintext"},
		{"md file", "README.md", "md"},
		{"uppercase md", "README.MD", "md"},
		{"build.gradle", "build.gradle", "gradle"},
		{"go.mod special", "go.mod", "go"},
		{"Pipfile special", "Pipfile", "toml"},
		{"Cargo.toml special", "Cargo.toml", "toml"},
		{"Dockerfile", "Dockerfile", "dockerfile"},
		{"c header maps to c", "header.h", "c"},
		{"cpp header maps to cpp", "header.hpp", "cpp"},
		{"cc maps to cpp", "util.cc", "cpp"},
		{"cxx maps to cpp", "util.cxx", "cpp"},
		{"hh maps to cpp", "header.hh", "cpp"},
		{"cs maps to csharp", "Program.cs", "csharp"},
		{"mjs maps to js", "util.mjs", "js"},
		{"cjs maps to js", "util.cjs", "js"},
		{"mts maps to ts", "types.mts", "ts"},
		{"cts maps to ts", "types.cts", "ts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMarkdownLanguage(tt.filename, allowedFileNames); got != tt.want {
				t.Errorf("GetMarkdownLanguage(%q) = %v; want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestParseLanguages(t *testing.T) {
	defaultsExpected := make(map[string]bool)
	for k, v := range supportedLanguages {
		defaultsExpected[k] = v
	}

	tests := []struct {
		name     string
		input    string
		expected map[string]bool
	}{
		{"empty string uses defaults", "", defaultsExpected},
		{"single language", ".go", langMap(".go")},
		{"multiple languages", ".go,.js,.php", langMap(".go", ".js", ".php")},
		{"languages without dots", "go,js", langMap(".go", ".js")},
		{"uppercase languages", "GO,JS", langMap(".go", ".js")},
		{"languages with spaces", " go , js ", langMap(".go", ".js")},
		{"unsupported language ignored", "go,ruby,js", langMap(".go", ".js")},
		{"trailing comma ignored", "go,", langMap(".go")},
		{"blank token ignored", "go, ,js", langMap(".go", ".js")},
		{"dockerfile explicitly enabled", "dockerfile", langMap(".dockerfile")},
		{"dockerfile uppercase", "DOCKERFILE", langMap(".dockerfile")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseLanguages(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseLanguages(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetAllowedFileNames(t *testing.T) {
	tests := []struct {
		name             string
		allowedLanguages map[string]bool
		expected         map[string]bool
	}{
		{
			name: "union of enabled languages, dockerfile contributes nothing",
			allowedLanguages: map[string]bool{
				".go":         true,
				".java":       true,
				".dockerfile": true,
			},
			expected: map[string]bool{
				"go.mod":          true,
				"pom.xml":         true,
				"build.gradle":    true,
				"settings.gradle": true,
			},
		},
		{
			name: "disabled language contributes nothing",
			allowedLanguages: map[string]bool{
				".ts": true,
				".py": false,
			},
			expected: map[string]bool{
				"package.json":  true,
				"tsconfig.json": true,
			},
		},
		{
			name:             "no languages",
			allowedLanguages: map[string]bool{},
			expected:         map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAllowedFileNames(tt.allowedLanguages)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetAllowedFileNames() = %v; want %v", result, tt.expected)
			}
		})
	}
}

func TestLanguageNames(t *testing.T) {
	allowedLanguages := map[string]bool{
		".go":         true,
		".js":         true,
		".php":        false,
		".py":         false,
		".dockerfile": true,
	}

	t.Run("enabled languages sorted", func(t *testing.T) {
		if got := LanguageNames(allowedLanguages, true); !reflect.DeepEqual(got, []string{"dockerfile", "go", "js"}) {
			t.Errorf("LanguageNames(m, true) = %v; want [dockerfile go js]", got)
		}
	})

	t.Run("disabled languages sorted", func(t *testing.T) {
		if got := LanguageNames(allowedLanguages, false); !reflect.DeepEqual(got, []string{"php", "py"}) {
			t.Errorf("LanguageNames(m, false) = %v; want [php py]", got)
		}
	})
}
