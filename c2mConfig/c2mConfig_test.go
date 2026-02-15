package c2mConfig

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestIsConfigValid(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{"valid config", &Config{InputFolder: "input", OutputMarkdown: "output", MaxFileSize: 1024}, true},
		{"empty input folder", &Config{InputFolder: "", OutputMarkdown: "output", MaxFileSize: 1024}, false},
		{"valid without output", &Config{InputFolder: "input", OutputMarkdown: "", MaxFileSize: 1024}, true},
		{"nil config", nil, false},
		{"zero max file size", &Config{InputFolder: "input", MaxFileSize: 0}, false},
		{"negative max file size", &Config{InputFolder: "input", MaxFileSize: -1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsConfigValid(tt.config); got != tt.want {
				t.Errorf("IsConfigValid(%v) = %v; want %v", tt.config, got, tt.want)
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
