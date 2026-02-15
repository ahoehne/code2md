package c2mConfig

import (
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestInitializeConfigFromFlags(t *testing.T) {
	t.Run("default ignore patterns present when no flags passed", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		origArgs := os.Args
		defer func() { os.Args = origArgs }()

		tempDir := t.TempDir()
		origDir, _ := os.Getwd()
		os.Chdir(tempDir)
		defer os.Chdir(origDir)

		os.Args = []string{"cmd", "-i", tempDir}
		config, err := InitializeConfigFromFlags()
		if err != nil {
			t.Fatalf("InitializeConfigFromFlags() error: %v", err)
		}

		defaults := strings.Split(defaultIgnoredPatterns, ",")
		for _, d := range defaults {
			found := false
			for _, p := range config.IgnorePatterns {
				if p == d {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected default ignore pattern %q in %v", d, config.IgnorePatterns)
			}
		}
	})

	t.Run("explicit ignore overrides defaults", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		origArgs := os.Args
		defer func() { os.Args = origArgs }()

		tempDir := t.TempDir()
		origDir, _ := os.Getwd()
		os.Chdir(tempDir)
		defer os.Chdir(origDir)

		os.Args = []string{"cmd", "-i", tempDir, "--ignore", "custom.txt,other.log"}
		config, err := InitializeConfigFromFlags()
		if err != nil {
			t.Fatalf("InitializeConfigFromFlags() error: %v", err)
		}

		for _, expected := range []string{"custom.txt", "other.log"} {
			found := false
			for _, p := range config.IgnorePatterns {
				if p == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected ignore pattern %q in %v", expected, config.IgnorePatterns)
			}
		}

		for _, d := range strings.Split(defaultIgnoredPatterns, ",") {
			for _, p := range config.IgnorePatterns {
				if p == d {
					t.Errorf("default pattern %q should not be present when --ignore is explicit, got %v", d, config.IgnorePatterns)
				}
			}
		}
	})

	t.Run("output file added to ignore patterns", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		origArgs := os.Args
		defer func() { os.Args = origArgs }()

		tempDir := t.TempDir()
		origDir, _ := os.Getwd()
		os.Chdir(tempDir)
		defer os.Chdir(origDir)

		os.Args = []string{"cmd", "-i", tempDir, "-o", "output.md"}
		config, err := InitializeConfigFromFlags()
		if err != nil {
			t.Fatalf("InitializeConfigFromFlags() error: %v", err)
		}

		found := false
		for _, p := range config.IgnorePatterns {
			if p == "output.md" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected output file 'output.md' in ignore patterns %v", config.IgnorePatterns)
		}
	})
}

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
