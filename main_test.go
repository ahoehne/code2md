package main

import (
	"bytes"
	"code2md/c2mConfig"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestRun(t *testing.T) {
	t.Run("writes to stdout when no output specified", func(t *testing.T) {
		tempDir := t.TempDir()
		err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		config := &c2mConfig.Config{
			InputFolder:      tempDir,
			OutputMarkdown:   "",
			AllowedLanguages: map[string]bool{".go": true},
			AllowedFileNames: map[string]bool{},
			IgnorePatterns:   []string{},
			MaxFileSize:      100 * 1024 * 1024,
		}

		var runErr error
		output := captureStdout(t, func() {
			runErr = run(config)
		})

		if runErr != nil {
			t.Errorf("run() error: %v", runErr)
		}
		if !strings.Contains(output, "main.go") {
			t.Error("Stdout should contain main.go")
		}
	})

	t.Run("writes to file when output specified", func(t *testing.T) {
		tempDir := t.TempDir()
		err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		outputFile := filepath.Join(tempDir, "output.md")
		config := &c2mConfig.Config{
			InputFolder:      tempDir,
			OutputMarkdown:   outputFile,
			AllowedLanguages: map[string]bool{".go": true},
			AllowedFileNames: map[string]bool{},
			IgnorePatterns:   []string{},
			MaxFileSize:      100 * 1024 * 1024,
		}

		err = run(config)
		if err != nil {
			t.Errorf("run() error: %v", err)
		}

		content, _ := os.ReadFile(outputFile)
		if !strings.Contains(string(content), "main.go") {
			t.Error("Output file should contain main.go")
		}
	})

	t.Run("creates output directory if needed", func(t *testing.T) {
		tempDir := t.TempDir()
		err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		outputFile := filepath.Join(tempDir, "subdir", "nested", "output.md")
		config := &c2mConfig.Config{
			InputFolder:      tempDir,
			OutputMarkdown:   outputFile,
			AllowedLanguages: map[string]bool{".go": true},
			AllowedFileNames: map[string]bool{},
			IgnorePatterns:   []string{},
			MaxFileSize:      100 * 1024 * 1024,
		}

		err = run(config)
		if err != nil {
			t.Errorf("run() error: %v", err)
		}

		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Error("Output file should be created in nested directory")
		}
	})
}

func TestDisplayVersion(t *testing.T) {
	output := captureStdout(t, func() {
		displayVersion()
	})

	if !strings.Contains(output, "code2md") {
		t.Error("Output should contain 'code2md'")
	}
}

func TestDisplayUsageInstructions(t *testing.T) {
	t.Run("with nil config", func(t *testing.T) {
		output := captureStdout(t, func() {
			displayUsageInstructions(nil, false)
		})

		if !strings.Contains(output, "Usage:") {
			t.Error("Output should contain usage instructions")
		}
		if strings.Contains(output, "By default") {
			t.Error("Should not show language info with nil config")
		}
	})

	t.Run("with valid config shows languages", func(t *testing.T) {
		config := &c2mConfig.Config{
			AllowedLanguages: map[string]bool{".go": true, ".js": false},
		}

		output := captureStdout(t, func() {
			displayUsageInstructions(config, false)
		})

		if !strings.Contains(output, "By default") {
			t.Error("Output should show language info with valid config")
		}
	})

	t.Run("shows error when requested", func(t *testing.T) {
		output := captureStdout(t, func() {
			displayUsageInstructions(nil, true)
		})

		if !strings.Contains(output, "Error:") {
			t.Error("Output should contain error message")
		}
	})
}
