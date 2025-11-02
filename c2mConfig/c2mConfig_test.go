package c2mConfig

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetActiveLanguages(t *testing.T) {
	active := GetActiveLanguages()
	expected := 6

	if len(active) != 6 {
		t.Errorf("GetActiveLanguages() = %v (count: %d); want %d", active, len(active), expected)
	}
}

func TestGetInactiveLanguages(t *testing.T) {
	inactive := GetInactiveLanguages()
	expected := 6

	if len(inactive) != expected {
		t.Errorf("GetInactiveLanguages() = %v (count: %d); want %d", inactive, len(inactive), expected)
	}
}

func TestInitializeConfigFromFlags(t *testing.T) {
	// Set up test flags
	os.Args = []string{"cmd", "-input", "test_input", "-output", "test_output", "-languages", ".go,.js", "-ignore", ".txt,.log"}
	config := InitializeConfigFromFlags()

	expectedConfig := Config{
		InputFolder:    "test_input",
		OutputMarkdown: "test_output",
		AllowedFileNames: map[string]bool{
			"go.mod":       true,
			"package.json": true,
		},
		IgnorePatterns: []string{".txt", ".log"},
	}

	if !reflect.DeepEqual(config, expectedConfig) {
		t.Errorf("InitializeConfigFromFlags() = %v; want %v", config, expectedConfig)
	}
}

func TestIsConfigValid(t *testing.T) {
	tests := []struct {
		config Config
		want   bool
	}{
		{Config{InputFolder: "input", OutputMarkdown: "output"}, true},
		{Config{InputFolder: "", OutputMarkdown: "output"}, false},
		{Config{InputFolder: "input", OutputMarkdown: ""}, false},
		{Config{InputFolder: "", OutputMarkdown: ""}, false},
	}

	for _, tt := range tests {
		if got := IsConfigValid(tt.config); got != tt.want {
			t.Errorf("IsConfigValid(%v) = %v; want %v", tt.config, got, tt.want)
		}
	}
}

func TestUpdateLanguagesFilter(t *testing.T) {
	updateLanguagesFilter(".go,.js")
	expected := map[string]bool{
		".php":  false,
		".go":   true,
		".js":   true,
		".ts":   false,
		".py":   false,
		".sh":   false,
		".html": false,
		".scss": false,
		".css":  false,
		".json": false,
		".yaml": false,
		".yml":  false,
	}

	if !reflect.DeepEqual(GetAllowedLanguages(), expected) {
		t.Errorf("updateLanguagesFilter() = %v; want %v", GetAllowedLanguages(), expected)
	}
}

func TestFetchAllowedFileNames(t *testing.T) {
	updateLanguagesFilter(".go,.php,.js")
	allowedFileNames := fetchAllowedFileNames(GetAllowedLanguages())
	expected := map[string]bool{
		"go.mod":        true,
		"composer.json": true,
		"package.json":  true,
	}
	if !reflect.DeepEqual(allowedFileNames, expected) {
		t.Errorf("fetchAllowedFileNames() = %v; want %v", allowedFileNames, expected)
	}

}
func TestFetchAllowedFileNamesCapitalLetters(t *testing.T) {
	updateLanguagesFilter("GO,PHP")
	allowedFileNames := fetchAllowedFileNames(GetAllowedLanguages())
	expected := map[string]bool{
		"go.mod":        true,
		"composer.json": true,
	}
	if !reflect.DeepEqual(allowedFileNames, expected) {
		t.Errorf("fetchAllowedFileNames (Capital Letters): %v; want %v", allowedFileNames, expected)
	}
}

func TestParseIgnorePatterns(t *testing.T) {
	patterns := parseIgnorePatterns(".txt,.log")
	expected := []string{".txt", ".log"}

	if !reflect.DeepEqual(patterns, expected) {
		t.Errorf("parseIgnorePatterns() = %v; want %v", patterns, expected)
	}
}

func TestParseIgnorePatternsEmpty(t *testing.T) {
	patterns := parseIgnorePatterns("")
	expected := []string{}

	if !reflect.DeepEqual(patterns, expected) {
		t.Errorf("parseIgnorePatterns() = %v; want %v", patterns, expected)
	}
}

func TestLoadGitignorePatterns(t *testing.T) {
	tempDir := t.TempDir()
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	err := os.WriteFile(gitignorePath, []byte("*.txt\n*.log\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write .gitignore file: %v", err)
	}

	patterns, err := LoadGitignorePatterns(gitignorePath)
	if err != nil {
		t.Errorf("LoadGitignorePatterns() threw an unexpected error: %v", err)
	}
	expected := []string{"*.txt", "*.log"}

	if !reflect.DeepEqual(patterns, expected) {
		t.Errorf("LoadGitignorePatterns() = %v; want %v", patterns, expected)
	}
}

func TestSliceContains(t *testing.T) {
	tests := []struct {
		slice []string
		item  string
		want  bool
	}{
		{[]string{"go", "php", "js"}, "go", true},
		{[]string{"go", "php", "js"}, "java", false},
		{[]string{}, "go", false},
		{[]string{"go", "php", "js"}, "Go", false},
		{[]string{"go", "php", "js", "Go"}, "Go", true},
	}

	for _, tt := range tests {
		if got := sliceContains(tt.slice, tt.item); got != tt.want {
			t.Errorf("sliceContains(%v, %q) = %v; want %v", tt.slice, tt.item, got, tt.want)
		}
	}
}
