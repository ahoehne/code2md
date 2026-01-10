package c2mConfig

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseLanguages(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]bool
	}{
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
	allowedLanguages := map[string]bool{
		".go":   true,
		".php":  true,
		".js":   true,
		".java": true,
	}

	allowedFileNames := getMapOfAllowedFileNames(allowedLanguages)
	expected := map[string]bool{
		"go.mod":        true,
		"composer.json": true,
		"package.json":  true,
		"pom.xml":       true,
	}

	if !reflect.DeepEqual(allowedFileNames, expected) {
		t.Errorf("getMapOfAllowedFileNames() = %v; want %v", allowedFileNames, expected)
	}
}

func TestLoadGitignorePatterns(t *testing.T) {
	tempDir := t.TempDir()
	gitignorePath := filepath.Join(tempDir, ".gitignore")

	content := "*.txt\n*.log\n# comment\n\n"
	err := os.WriteFile(gitignorePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write .gitignore file: %v", err)
	}

	patterns, err := LoadGitignorePatterns(gitignorePath)
	if err != nil {
		t.Errorf("LoadGitignorePatterns() error: %v", err)
	}

	expected := []string{"*.txt", "*.log"}
	if !reflect.DeepEqual(patterns, expected) {
		t.Errorf("LoadGitignorePatterns() = %v; want %v", patterns, expected)
	}
}

func TestLoadGitignorePatternsNonExistent(t *testing.T) {
	patterns, err := LoadGitignorePatterns("/nonexistent/.gitignore")
	if err != nil {
		t.Errorf("LoadGitignorePatterns() should not error on non-existent file: %v", err)
	}

	if len(patterns) != 0 {
		t.Errorf("LoadGitignorePatterns() = %v; want empty slice", patterns)
	}
}

func TestGetActiveLanguages(t *testing.T) {
	config := &Config{
		AllowedLanguages: map[string]bool{
			".go":  true,
			".js":  true,
			".php": false,
		},
	}

	active := GetActiveLanguages(config)

	if len(active) != 2 {
		t.Errorf("GetActiveLanguages() returned %d languages; want 2", len(active))
	}

	hasGo := false
	hasJs := false
	for _, lang := range active {
		if lang == "go" {
			hasGo = true
		}
		if lang == "js" {
			hasJs = true
		}
	}

	if !hasGo || !hasJs {
		t.Errorf("GetActiveLanguages() = %v; want [go, js]", active)
	}
}

func TestGetInactiveLanguages(t *testing.T) {
	config := &Config{
		AllowedLanguages: map[string]bool{
			".go":  true,
			".js":  true,
			".php": false,
			".py":  false,
		},
	}

	inactive := GetInactiveLanguages(config)

	if len(inactive) != 2 {
		t.Errorf("GetInactiveLanguages() returned %d languages; want 2", len(inactive))
	}
}
