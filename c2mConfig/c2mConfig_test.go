package c2mConfig

import (
	"flag"
	"os"
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
			if !sliceContains(config.DefaultIgnorePatterns, d) {
				t.Errorf("expected default ignore pattern %q in %v", d, config.DefaultIgnorePatterns)
			}
		}
		if len(config.UserIgnorePatterns) != 0 {
			t.Errorf("expected no user ignore patterns, got %v", config.UserIgnorePatterns)
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
			if !sliceContains(config.UserIgnorePatterns, expected) {
				t.Errorf("expected ignore pattern %q in %v", expected, config.UserIgnorePatterns)
			}
		}

		for _, d := range strings.Split(defaultIgnoredPatterns, ",") {
			if sliceContains(config.UserIgnorePatterns, d) || sliceContains(config.DefaultIgnorePatterns, d) {
				t.Errorf("default pattern %q should not be present when --ignore is explicit", d)
			}
		}
	})

	t.Run("output file is not turned into an ignore pattern", func(t *testing.T) {
		tempDir := t.TempDir()
		cleanup := setupFlagTest(t, "-i", tempDir, "-o", "output.md")
		defer cleanup()

		config, err := InitializeConfigFromFlags()
		if err != nil {
			t.Fatalf("InitializeConfigFromFlags() error: %v", err)
		}

		if sliceContains(config.UserIgnorePatterns, "output.md") || sliceContains(config.DefaultIgnorePatterns, "output.md") {
			t.Errorf("output file should be excluded by path, not by pattern, got %v / %v", config.UserIgnorePatterns, config.DefaultIgnorePatterns)
		}
		if config.OutputMarkdown != "output.md" {
			t.Errorf("expected OutputMarkdown 'output.md', got %q", config.OutputMarkdown)
		}
	})

	t.Run("default yml ignore dropped when yml language enabled", func(t *testing.T) {
		tempDir := t.TempDir()
		cleanup := setupFlagTest(t, "-i", tempDir, "-l", "yml,go")
		defer cleanup()

		config, err := InitializeConfigFromFlags()
		if err != nil {
			t.Fatalf("InitializeConfigFromFlags() error: %v", err)
		}

		if sliceContains(config.DefaultIgnorePatterns, "*.yml") {
			t.Errorf("*.yml should be removed from defaults when yml is enabled, got %v", config.DefaultIgnorePatterns)
		}
		if !sliceContains(config.DefaultIgnorePatterns, "*.xml") {
			t.Errorf("*.xml should remain when only yml is enabled, got %v", config.DefaultIgnorePatterns)
		}
		if !sliceContains(config.DefaultIgnorePatterns, "*.yaml") {
			t.Errorf("*.yaml should remain when only yml is enabled, got %v", config.DefaultIgnorePatterns)
		}
	})

	t.Run("explicit ignore preserved even when matching language is enabled", func(t *testing.T) {
		tempDir := t.TempDir()
		cleanup := setupFlagTest(t, "-i", tempDir, "-l", "yml", "--ignore", "*.yml")
		defer cleanup()

		config, err := InitializeConfigFromFlags()
		if err != nil {
			t.Fatalf("InitializeConfigFromFlags() error: %v", err)
		}

		if !sliceContains(config.UserIgnorePatterns, "*.yml") {
			t.Errorf("explicit --ignore *.yml should be preserved, got %v", config.UserIgnorePatterns)
		}
	})

	t.Run("min.css ignored by default when css enabled", func(t *testing.T) {
		tempDir := t.TempDir()
		cleanup := setupFlagTest(t, "-i", tempDir, "-l", "css")
		defer cleanup()

		config, err := InitializeConfigFromFlags()
		if err != nil {
			t.Fatalf("InitializeConfigFromFlags() error: %v", err)
		}

		if !sliceContains(config.DefaultIgnorePatterns, "**.min.css") {
			t.Errorf("expected **.min.css in default ignore patterns %v", config.DefaultIgnorePatterns)
		}
	})

	t.Run("rejects positional arguments", func(t *testing.T) {
		tempDir := t.TempDir()
		cleanup := setupFlagTest(t, "-i", tempDir, "stray-argument")
		defer cleanup()

		if _, err := InitializeConfigFromFlags(); err == nil {
			t.Error("expected error for unexpected positional arguments")
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
