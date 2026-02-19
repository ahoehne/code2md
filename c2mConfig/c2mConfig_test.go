package c2mConfig

import (
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func setupFlagTest(t *testing.T, args ...string) (cleanup func()) {
	t.Helper()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	origArgs := os.Args
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tempDir)
	os.Args = append([]string{"cmd"}, args...)
	return func() {
		os.Args = origArgs
		os.Chdir(origDir)
	}
}

func sliceContains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

func TestInitializeConfigFromFlags(t *testing.T) {
	t.Run("default ignore patterns present when no flags passed", func(t *testing.T) {
		tempDir := t.TempDir()
		cleanup := setupFlagTest(t, "-i", tempDir)
		defer cleanup()

		config, err := InitializeConfigFromFlags()
		if err != nil {
			t.Fatalf("InitializeConfigFromFlags() error: %v", err)
		}

		for _, d := range strings.Split(defaultIgnoredPatterns, ",") {
			if !sliceContains(config.IgnorePatterns, d) {
				t.Errorf("expected default ignore pattern %q in %v", d, config.IgnorePatterns)
			}
		}
	})

	t.Run("explicit ignore overrides defaults", func(t *testing.T) {
		tempDir := t.TempDir()
		cleanup := setupFlagTest(t, "-i", tempDir, "--ignore", "custom.txt,other.log")
		defer cleanup()

		config, err := InitializeConfigFromFlags()
		if err != nil {
			t.Fatalf("InitializeConfigFromFlags() error: %v", err)
		}

		for _, expected := range []string{"custom.txt", "other.log"} {
			if !sliceContains(config.IgnorePatterns, expected) {
				t.Errorf("expected ignore pattern %q in %v", expected, config.IgnorePatterns)
			}
		}

		for _, d := range strings.Split(defaultIgnoredPatterns, ",") {
			if sliceContains(config.IgnorePatterns, d) {
				t.Errorf("default pattern %q should not be present when --ignore is explicit, got %v", d, config.IgnorePatterns)
			}
		}
	})

	t.Run("output file added to ignore patterns", func(t *testing.T) {
		tempDir := t.TempDir()
		cleanup := setupFlagTest(t, "-i", tempDir, "-o", "output.md")
		defer cleanup()

		config, err := InitializeConfigFromFlags()
		if err != nil {
			t.Fatalf("InitializeConfigFromFlags() error: %v", err)
		}

		if !sliceContains(config.IgnorePatterns, "output.md") {
			t.Errorf("expected output file 'output.md' in ignore patterns %v", config.IgnorePatterns)
		}
	})
}

func TestInitializeConfigFromFlags_InputFolderGitignore(t *testing.T) {
	t.Run("loads gitignore from input folder", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		origArgs := os.Args
		defer func() { os.Args = origArgs }()

		cwdDir := t.TempDir()
		inputDir := t.TempDir()

		os.WriteFile(filepath.Join(cwdDir, ".gitignore"), []byte("cwd_pattern\n"), 0644)
		os.WriteFile(filepath.Join(inputDir, ".gitignore"), []byte("input_pattern\n"), 0644)

		origDir, _ := os.Getwd()
		os.Chdir(cwdDir)
		defer os.Chdir(origDir)

		os.Args = []string{"cmd", "-i", inputDir}
		config, err := InitializeConfigFromFlags()
		if err != nil {
			t.Fatalf("InitializeConfigFromFlags() error: %v", err)
		}

		if !sliceContains(config.IgnorePatterns, "cwd_pattern") {
			t.Errorf("expected cwd gitignore pattern in %v", config.IgnorePatterns)
		}
		if !sliceContains(config.IgnorePatterns, "input_pattern") {
			t.Errorf("expected input folder gitignore pattern in %v", config.IgnorePatterns)
		}
	})

	t.Run("does not duplicate when input is cwd", func(t *testing.T) {
		tempDir := t.TempDir()
		cleanup := setupFlagTest(t, "-i", ".")
		defer cleanup()

		os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte("some_pattern\n"), 0644)

		origDir, _ := os.Getwd()
		os.Chdir(tempDir)
		defer os.Chdir(origDir)

		config, err := InitializeConfigFromFlags()
		if err != nil {
			t.Fatalf("InitializeConfigFromFlags() error: %v", err)
		}

		count := 0
		for _, p := range config.IgnorePatterns {
			if p == "some_pattern" {
				count++
			}
		}
		if count != 1 {
			t.Errorf("expected pattern once, found %d times in %v", count, config.IgnorePatterns)
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

		patterns, err := loadGitignorePatterns(gitignorePath)
		if err != nil {
			t.Errorf("loadGitignorePatterns() error: %v", err)
		}

		expected := []string{"*.txt", "*.log", "spaced"}
		if !reflect.DeepEqual(patterns, expected) {
			t.Errorf("loadGitignorePatterns() = %v; want %v", patterns, expected)
		}
	})

	t.Run("non-existent file returns empty slice", func(t *testing.T) {
		patterns, err := loadGitignorePatterns("/nonexistent/.gitignore")
		if err != nil {
			t.Errorf("loadGitignorePatterns() should not error on non-existent file: %v", err)
		}
		if len(patterns) != 0 {
			t.Errorf("loadGitignorePatterns() = %v; want empty slice", patterns)
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

		_, err = loadGitignorePatterns(gitignorePath)
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

		patterns, err := loadGitignorePatterns(gitignorePath)
		if err != nil {
			t.Errorf("loadGitignorePatterns() error: %v", err)
		}
		if len(patterns) != 0 {
			t.Errorf("loadGitignorePatterns() = %v; want empty slice", patterns)
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

		patterns, err := loadGitignorePatterns(gitignorePath)
		if err != nil {
			t.Errorf("loadGitignorePatterns() error: %v", err)
		}
		if len(patterns) != 0 {
			t.Errorf("loadGitignorePatterns() = %v; want empty slice", patterns)
		}
	})
}
