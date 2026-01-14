package c2mConfig

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

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
				".py": true, ".sh": true, ".java": true, ".md": false,
				".html": false, ".scss": false, ".css": false, ".json": false,
				".yaml": false, ".yml": false, ".xml": false,
			},
		},
		{
			name:  "single language",
			input: ".go",
			expected: map[string]bool{
				".go": true, ".php": false, ".js": false, ".ts": false,
				".py": false, ".sh": false, ".java": false, ".md": false,
				".html": false, ".scss": false, ".css": false, ".json": false,
				".yaml": false, ".yml": false, ".xml": false,
			},
		},
		{
			name:  "multiple languages",
			input: ".go,.js,.php",
			expected: map[string]bool{
				".go": true, ".php": true, ".js": true, ".ts": false,
				".py": false, ".sh": false, ".java": false, ".md": false,
				".html": false, ".scss": false, ".css": false, ".json": false,
				".yaml": false, ".yml": false, ".xml": false,
			},
		},
		{
			name:  "languages without dots",
			input: "go,js",
			expected: map[string]bool{
				".go": true, ".js": true, ".php": false, ".ts": false,
				".py": false, ".sh": false, ".java": false, ".md": false,
				".html": false, ".scss": false, ".css": false, ".json": false,
				".yaml": false, ".yml": false, ".xml": false,
			},
		},
		{
			name:  "uppercase languages",
			input: "GO,JS",
			expected: map[string]bool{
				".go": true, ".js": true, ".php": false, ".ts": false,
				".py": false, ".sh": false, ".java": false, ".md": false,
				".html": false, ".scss": false, ".css": false, ".json": false,
				".yaml": false, ".yml": false, ".xml": false,
			},
		},
		{
			name:  "languages with spaces",
			input: " go , js ",
			expected: map[string]bool{
				".go": true, ".js": true, ".php": false, ".ts": false,
				".py": false, ".sh": false, ".java": false, ".md": false,
				".html": false, ".scss": false, ".css": false, ".json": false,
				".yaml": false, ".yml": false, ".xml": false,
			},
		},
		{
			name:  "unsupported language ignored",
			input: "go,ruby,js",
			expected: map[string]bool{
				".go": true, ".js": true, ".php": false, ".ts": false,
				".py": false, ".sh": false, ".java": false, ".md": false,
				".html": false, ".scss": false, ".css": false, ".json": false,
				".yaml": false, ".yml": false, ".xml": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLanguages(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseLanguages(%q) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsConfigValid(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{"valid config", &Config{InputFolder: "input", OutputMarkdown: "output"}, true},
		{"empty input folder", &Config{InputFolder: "", OutputMarkdown: "output"}, false},
		{"valid without output", &Config{InputFolder: "input", OutputMarkdown: ""}, true},
		{"nil config", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsConfigValid(tt.config); got != tt.want {
				t.Errorf("IsConfigValid(%v) = %v; want %v", tt.config, got, tt.want)
			}
		})
	}
}

func TestGetMapOfAllowedFileNames(t *testing.T) {
	tests := []struct {
		name             string
		allowedLanguages map[string]bool
		expected         map[string]bool
	}{
		{
			name: "all languages",
			allowedLanguages: map[string]bool{
				".go":   true,
				".php":  true,
				".js":   true,
				".ts":   true,
				".java": true,
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
			name:             "no languages",
			allowedLanguages: map[string]bool{},
			expected:         map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMapOfAllowedFileNames(tt.allowedLanguages)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("getMapOfAllowedFileNames() = %v; want %v", result, tt.expected)
			}
		})
	}
}

func TestLoadGitignorePatterns(t *testing.T) {
	t.Run("valid gitignore", func(t *testing.T) {
		tempDir := t.TempDir()
		gitignorePath := filepath.Join(tempDir, ".gitignore")

		content := "*.txt\n*.log\n# comment\n\n   \n  spaced  \n"
		err := os.WriteFile(gitignorePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write .gitignore file: %v", err)
		}

		patterns, err := LoadGitignorePatterns(gitignorePath)
		if err != nil {
			t.Errorf("LoadGitignorePatterns() error: %v", err)
		}

		expected := []string{"*.txt", "*.log", "spaced"}
		if !reflect.DeepEqual(patterns, expected) {
			t.Errorf("LoadGitignorePatterns() = %v; want %v", patterns, expected)
		}
	})

	t.Run("non-existent file returns empty slice", func(t *testing.T) {
		patterns, err := LoadGitignorePatterns("/nonexistent/.gitignore")
		if err != nil {
			t.Errorf("LoadGitignorePatterns() should not error on non-existent file: %v", err)
		}
		if len(patterns) != 0 {
			t.Errorf("LoadGitignorePatterns() = %v; want empty slice", patterns)
		}
	})

	t.Run("permission denied returns error", func(t *testing.T) {
		tempDir := t.TempDir()
		gitignorePath := filepath.Join(tempDir, ".gitignore")

		err := os.WriteFile(gitignorePath, []byte("*.txt"), 0644)
		if err != nil {
			t.Fatalf("Failed to write .gitignore file: %v", err)
		}

		os.Chmod(gitignorePath, 0000)
		defer os.Chmod(gitignorePath, 0644)

		_, err = LoadGitignorePatterns(gitignorePath)
		if err == nil {
			t.Skip("Test requires permission denied error, skipping on systems that allow root access")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		tempDir := t.TempDir()
		gitignorePath := filepath.Join(tempDir, ".gitignore")

		err := os.WriteFile(gitignorePath, []byte(""), 0644)
		if err != nil {
			t.Fatalf("Failed to write .gitignore file: %v", err)
		}

		patterns, err := LoadGitignorePatterns(gitignorePath)
		if err != nil {
			t.Errorf("LoadGitignorePatterns() error: %v", err)
		}
		if len(patterns) != 0 {
			t.Errorf("LoadGitignorePatterns() = %v; want empty slice", patterns)
		}
	})

	t.Run("only comments and whitespace", func(t *testing.T) {
		tempDir := t.TempDir()
		gitignorePath := filepath.Join(tempDir, ".gitignore")

		content := "# comment\n\n   \n# another comment\n"
		err := os.WriteFile(gitignorePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write .gitignore file: %v", err)
		}

		patterns, err := LoadGitignorePatterns(gitignorePath)
		if err != nil {
			t.Errorf("LoadGitignorePatterns() error: %v", err)
		}
		if len(patterns) != 0 {
			t.Errorf("LoadGitignorePatterns() = %v; want empty slice", patterns)
		}
	})
}

func TestGetActiveLanguages(t *testing.T) {
	t.Run("returns active languages", func(t *testing.T) {
		config := &Config{
			AllowedLanguages: map[string]bool{
				".go":  true,
				".js":  true,
				".php": false,
			},
		}

		active := GetActiveLanguages(config)
		sort.Strings(active)

		expected := []string{"go", "js"}
		sort.Strings(expected)

		if !reflect.DeepEqual(active, expected) {
			t.Errorf("GetActiveLanguages() = %v; want %v", active, expected)
		}
	})

	t.Run("nil config returns empty slice", func(t *testing.T) {
		active := GetActiveLanguages(nil)
		if len(active) != 0 {
			t.Errorf("GetActiveLanguages(nil) = %v; want empty slice", active)
		}
	})

	t.Run("no active languages", func(t *testing.T) {
		config := &Config{
			AllowedLanguages: map[string]bool{
				".go":  false,
				".js":  false,
				".php": false,
			},
		}

		active := GetActiveLanguages(config)
		if len(active) != 0 {
			t.Errorf("GetActiveLanguages() = %v; want empty slice", active)
		}
	})
}

func TestGetInactiveLanguages(t *testing.T) {
	t.Run("returns inactive languages", func(t *testing.T) {
		config := &Config{
			AllowedLanguages: map[string]bool{
				".go":  true,
				".js":  true,
				".php": false,
				".py":  false,
			},
		}

		inactive := GetInactiveLanguages(config)
		sort.Strings(inactive)

		expected := []string{"php", "py"}
		sort.Strings(expected)

		if !reflect.DeepEqual(inactive, expected) {
			t.Errorf("GetInactiveLanguages() = %v; want %v", inactive, expected)
		}
	})

	t.Run("nil config returns empty slice", func(t *testing.T) {
		inactive := GetInactiveLanguages(nil)
		if len(inactive) != 0 {
			t.Errorf("GetInactiveLanguages(nil) = %v; want empty slice", inactive)
		}
	})

	t.Run("all languages active", func(t *testing.T) {
		config := &Config{
			AllowedLanguages: map[string]bool{
				".go": true,
				".js": true,
			},
		}

		inactive := GetInactiveLanguages(config)
		if len(inactive) != 0 {
			t.Errorf("GetInactiveLanguages() = %v; want empty slice", inactive)
		}
	})
}

func TestGetDefaultLanguages(t *testing.T) {
	defaults := GetDefaultLanguages()

	expectedDefaults := []string{"php", "go", "js", "ts", "py", "sh", "java"}
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

	expectedSupported := []string{"php", "go", "js", "ts", "py", "sh", "java", "md", "html", "scss", "css", "json", "yaml", "yml", "xml"}

	if len(supported) != len(expectedSupported) {
		t.Errorf("GetSupportedLanguages() returned %d languages; want %d", len(supported), len(expectedSupported))
	}

	sort.Strings(supported)
	sort.Strings(expectedSupported)

	if !reflect.DeepEqual(supported, expectedSupported) {
		t.Errorf("GetSupportedLanguages() = %v; want %v", supported, expectedSupported)
	}
}
