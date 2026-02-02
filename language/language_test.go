package language

import (
	"reflect"
	"sort"
	"testing"
)

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
		".php":        true,
		".go":         true,
		".js":         true,
		".ts":         true,
		".java":       true,
		".json":       false,
		".dockerfile": true,
	}

	allowedFileNames := map[string]bool{
		"go.mod":        true,
		"composer.json": true,
		"package.json":  true,
		"pom.xml":       true,
	}

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"php file", "file.php", true},
		{"go file", "file.go", true},
		{"js file", "file.js", true},
		{"ts file", "file.ts", true},
		{"java file", "file.java", true},
		{"java file uppercase", "File.java", true},
		{"special file pom.xml", "pom.xml", true},
		{"txt file not allowed", "file.txt", false},
		{"special file go.mod", "go.mod", true},
		{"special file composer.json", "composer.json", true},
		{"Dockerfile capitalized", "Dockerfile", true},
		{"dockerfile lowercase", "dockerfile", true},
		{"DOCKERFILE uppercase", "DOCKERFILE", true},
		{"Dockerfile.dev multi-stage", "Dockerfile.dev", true},
		{"Dockerfile.prod multi-stage", "Dockerfile.prod", true},
		{"dockerfile.test lowercase multi-stage", "dockerfile.test", true},
		{"multi-dot js", "file.with.multiple.dots.js", true},
		{"no extension", "file_without_extension", false},
		{"multi-dot php", "file.with.multiple.dots.php", true},
		{"multi-dot go", "file.with.multiple.dots.go", true},
		{"multi-dot ts", "file.with.multiple.dots.ts", true},
		{"json file disabled", "file.json", false},
		{"hidden go file", ".hidden.go", true},
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
		"go.mod": true,
	}

	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{"go file", "main.go", "go"},
		{"js file", "script.js", "js"},
		{"go.mod special", "go.mod", "go"},
		{"Dockerfile capitalized", "Dockerfile", "dockerfile"},
		{"dockerfile lowercase", "dockerfile", "dockerfile"},
		{"DOCKERFILE uppercase", "DOCKERFILE", "dockerfile"},
		{"Dockerfile.dev multi-stage", "Dockerfile.dev", "dockerfile"},
		{"Dockerfile.prod multi-stage", "Dockerfile.prod", "dockerfile"},
		{"dockerfile.test multi-stage", "dockerfile.test", "dockerfile"},
		{"no extension", "README", "plaintext"},
		{"md file", "README.md", "md"},
		{"php file", "index.php", "php"},
		{"ts file", "app.ts", "ts"},
		{"java file", "Main.java", "java"},
		{"multi-dot extension", "file.test.js", "js"},
		{"hidden file with extension", ".eslintrc.json", "json"},
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
	tests := []struct {
		name     string
		input    string
		expected map[string]bool
	}{
		{
			name:  "empty string uses defaults",
			input: "",
			expected: map[string]bool{
				".go": true, ".php": true, ".js": true, ".ts": true,
				".py": true, ".sh": true, ".java": true, ".dockerfile": true,
				".md": false, ".html": false, ".scss": false, ".css": false,
				".json": false, ".yaml": false, ".yml": false, ".xml": false,
			},
		},
		{
			name:  "single language",
			input: ".go",
			expected: map[string]bool{
				".go": true, ".php": false, ".js": false, ".ts": false,
				".py": false, ".sh": false, ".java": false, ".dockerfile": false,
				".md": false, ".html": false, ".scss": false, ".css": false,
				".json": false, ".yaml": false, ".yml": false, ".xml": false,
			},
		},
		{
			name:  "multiple languages",
			input: ".go,.js,.php",
			expected: map[string]bool{
				".go": true, ".php": true, ".js": true, ".ts": false,
				".py": false, ".sh": false, ".java": false, ".dockerfile": false,
				".md": false, ".html": false, ".scss": false, ".css": false,
				".json": false, ".yaml": false, ".yml": false, ".xml": false,
			},
		},
		{
			name:  "languages without dots",
			input: "go,js",
			expected: map[string]bool{
				".go": true, ".js": true, ".php": false, ".ts": false,
				".py": false, ".sh": false, ".java": false, ".dockerfile": false,
				".md": false, ".html": false, ".scss": false, ".css": false,
				".json": false, ".yaml": false, ".yml": false, ".xml": false,
			},
		},
		{
			name:  "uppercase languages",
			input: "GO,JS",
			expected: map[string]bool{
				".go": true, ".js": true, ".php": false, ".ts": false,
				".py": false, ".sh": false, ".java": false, ".dockerfile": false,
				".md": false, ".html": false, ".scss": false, ".css": false,
				".json": false, ".yaml": false, ".yml": false, ".xml": false,
			},
		},
		{
			name:  "languages with spaces",
			input: " go , js ",
			expected: map[string]bool{
				".go": true, ".js": true, ".php": false, ".ts": false,
				".py": false, ".sh": false, ".java": false, ".dockerfile": false,
				".md": false, ".html": false, ".scss": false, ".css": false,
				".json": false, ".yaml": false, ".yml": false, ".xml": false,
			},
		},
		{
			name:  "unsupported language ignored",
			input: "go,ruby,js",
			expected: map[string]bool{
				".go": true, ".js": true, ".php": false, ".ts": false,
				".py": false, ".sh": false, ".java": false, ".dockerfile": false,
				".md": false, ".html": false, ".scss": false, ".css": false,
				".json": false, ".yaml": false, ".yml": false, ".xml": false,
			},
		},
		{
			name:  "dockerfile explicitly enabled",
			input: "dockerfile",
			expected: map[string]bool{
				".go": false, ".js": false, ".php": false, ".ts": false,
				".py": false, ".sh": false, ".java": false, ".dockerfile": true,
				".md": false, ".html": false, ".scss": false, ".css": false,
				".json": false, ".yaml": false, ".yml": false, ".xml": false,
			},
		},
		{
			name:  "dockerfile uppercase",
			input: "DOCKERFILE",
			expected: map[string]bool{
				".go": false, ".js": false, ".php": false, ".ts": false,
				".py": false, ".sh": false, ".java": false, ".dockerfile": true,
				".md": false, ".html": false, ".scss": false, ".css": false,
				".json": false, ".yaml": false, ".yml": false, ".xml": false,
			},
		},
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
			name: "all languages except dockerfile",
			allowedLanguages: map[string]bool{
				".go":         true,
				".php":        true,
				".js":         true,
				".ts":         true,
				".java":       true,
				".dockerfile": true,
			},
			expected: map[string]bool{
				"go.mod":        true,
				"composer.json": true,
				"package.json":  true,
				"tsconfig.json": true,
				"pom.xml":       true,
			},
		},
		{
			name: "only go",
			allowedLanguages: map[string]bool{
				".go": true,
			},
			expected: map[string]bool{
				"go.mod": true,
			},
		},
		{
			name: "only ts includes package.json and tsconfig.json",
			allowedLanguages: map[string]bool{
				".ts": true,
			},
			expected: map[string]bool{
				"package.json":  true,
				"tsconfig.json": true,
			},
		},
		{
			name: "js includes package.json",
			allowedLanguages: map[string]bool{
				".js": true,
			},
			expected: map[string]bool{
				"package.json": true,
			},
		},
		{
			name: "only dockerfile returns empty",
			allowedLanguages: map[string]bool{
				".dockerfile": true,
			},
			expected: map[string]bool{},
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

func TestGetActiveLanguages(t *testing.T) {
	t.Run("returns active languages", func(t *testing.T) {
		allowedLanguages := map[string]bool{
			".go":         true,
			".js":         true,
			".php":        false,
			".dockerfile": true,
		}

		active := GetActiveLanguages(allowedLanguages)
		sort.Strings(active)

		expected := []string{"dockerfile", "go", "js"}
		sort.Strings(expected)

		if !reflect.DeepEqual(active, expected) {
			t.Errorf("GetActiveLanguages() = %v; want %v", active, expected)
		}
	})

	t.Run("nil map returns empty slice", func(t *testing.T) {
		active := GetActiveLanguages(nil)
		if len(active) != 0 {
			t.Errorf("GetActiveLanguages(nil) = %v; want empty slice", active)
		}
	})

	t.Run("no active languages", func(t *testing.T) {
		allowedLanguages := map[string]bool{
			".go":  false,
			".js":  false,
			".php": false,
		}

		active := GetActiveLanguages(allowedLanguages)
		if len(active) != 0 {
			t.Errorf("GetActiveLanguages() = %v; want empty slice", active)
		}
	})
}

func TestGetInactiveLanguages(t *testing.T) {
	t.Run("returns inactive languages", func(t *testing.T) {
		allowedLanguages := map[string]bool{
			".go":         true,
			".js":         true,
			".php":        false,
			".py":         false,
			".dockerfile": true,
		}

		inactive := GetInactiveLanguages(allowedLanguages)
		sort.Strings(inactive)

		expected := []string{"php", "py"}
		sort.Strings(expected)

		if !reflect.DeepEqual(inactive, expected) {
			t.Errorf("GetInactiveLanguages() = %v; want %v", inactive, expected)
		}
	})

	t.Run("nil map returns empty slice", func(t *testing.T) {
		inactive := GetInactiveLanguages(nil)
		if len(inactive) != 0 {
			t.Errorf("GetInactiveLanguages(nil) = %v; want empty slice", inactive)
		}
	})

	t.Run("all languages active", func(t *testing.T) {
		allowedLanguages := map[string]bool{
			".go": true,
			".js": true,
		}

		inactive := GetInactiveLanguages(allowedLanguages)
		if len(inactive) != 0 {
			t.Errorf("GetInactiveLanguages() = %v; want empty slice", inactive)
		}
	})
}

func TestGetDefaultLanguages(t *testing.T) {
	defaults := GetDefaultLanguages()

	expectedDefaults := []string{"php", "go", "js", "ts", "py", "sh", "java", "dockerfile"}
	sort.Strings(defaults)
	sort.Strings(expectedDefaults)

	if len(defaults) != len(expectedDefaults) {
		t.Errorf("GetDefaultLanguages() returned %d languages; want %d", len(defaults), len(expectedDefaults))
	}

	if !reflect.DeepEqual(defaults, expectedDefaults) {
		t.Errorf("GetDefaultLanguages() = %v; want %v", defaults, expectedDefaults)
	}
}

func TestGetSupportedLanguages(t *testing.T) {
	supported := GetSupportedLanguages()

	expectedSupported := []string{"php", "go", "js", "ts", "py", "sh", "java", "dockerfile", "md", "html", "scss", "css", "json", "yaml", "yml", "xml"}

	if len(supported) != len(expectedSupported) {
		t.Errorf("GetSupportedLanguages() returned %d languages; want %d", len(supported), len(expectedSupported))
	}

	sort.Strings(supported)
	sort.Strings(expectedSupported)

	if !reflect.DeepEqual(supported, expectedSupported) {
		t.Errorf("GetSupportedLanguages() = %v; want %v", supported, expectedSupported)
	}
}
