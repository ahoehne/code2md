package processor

import (
	"bytes"
	"code2md/patternMatcher"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testMaxFileSize = 100 * 1024 * 1024

func TestWriteMarkdown(t *testing.T) {
	t.Run("writes go file with code fence", func(t *testing.T) {
		tempDir := t.TempDir()
		inputFile := filepath.Join(tempDir, "test.go")

		err := os.WriteFile(inputFile, []byte(
			"package main\nfunc main() {}\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create input file: %v", err)
		}

		var output bytes.Buffer
		err = writeMarkdown(inputFile, inputFile, &output, "go", testMaxFileSize)
		if err != nil {
			t.Errorf("writeMarkdown() error: %v", err)
		}

		contentStr := output.String()

		if !strings.Contains(contentStr, "# "+inputFile) {
			t.Error("Output should contain file path as header")
		}
		if !strings.Contains(contentStr, "```go") {
			t.Error("Output should contain go code fence")
		}
		if !strings.Contains(contentStr, "package main") {
			t.Error("Output should contain file content")
		}
		if !strings.Contains(contentStr, "```\n") {
			t.Error("Output should contain closing code fence")
		}
	})

	t.Run("writes md file without code fence", func(t *testing.T) {
		tempDir := t.TempDir()
		inputFile := filepath.Join(tempDir, "test.md")

		err := os.WriteFile(inputFile, []byte(
			"# Example Markdown Heading\nMarkdown Content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create input file: %v", err)
		}

		var output bytes.Buffer
		err = writeMarkdown(inputFile, inputFile, &output, "md", testMaxFileSize)
		if err != nil {
			t.Errorf("writeMarkdown() error: %v", err)
		}

		contentStr := output.String()

		if strings.Contains(contentStr, "```md") {
			t.Error("Markdown files should not have code fence")
		}
		if !strings.Contains(contentStr, "# Example Markdown Heading") {
			t.Error("Output should contain file heading")
		}
		if !strings.Contains(contentStr, "Markdown Content") {
			t.Error("Output should contain file content")
		}
	})

	t.Run("skips large files", func(t *testing.T) {
		tempDir := t.TempDir()
		inputFile := filepath.Join(tempDir, "large.md")

		largeContent := bytes.Repeat([]byte("x"), 1024)
		err := os.WriteFile(inputFile, largeContent, 0644)
		if err != nil {
			t.Fatalf("Failed to create input file: %v", err)
		}

		var output bytes.Buffer
		err = writeMarkdown(inputFile, inputFile, &output, "go", 100)
		if err != nil {
			t.Errorf("writeMarkdown() should not error for large files: %v", err)
		}

		if output.Len() != 0 {
			t.Error("Output should be empty for skipped large files")
		}
	})

	t.Run("adds newline before closing fence when file lacks trailing newline", func(t *testing.T) {
		tempDir := t.TempDir()
		inputFile := filepath.Join(tempDir, "test.go")

		err := os.WriteFile(inputFile, []byte("package main"), 0644)
		if err != nil {
			t.Fatalf("Failed to create input file: %v", err)
		}

		var output bytes.Buffer
		err = writeMarkdown(inputFile, inputFile, &output, "go", testMaxFileSize)
		if err != nil {
			t.Errorf("writeMarkdown() error: %v", err)
		}

		contentStr := output.String()
		if strings.Contains(contentStr, "main```") {
			t.Error("Closing fence should not be on the same line as code")
		}
		if !strings.Contains(contentStr, "main\n```") {
			t.Error("Closing fence should be on its own line")
		}
	})

	t.Run("uses display path in header", func(t *testing.T) {
		tempDir := t.TempDir()
		inputFile := filepath.Join(tempDir, "test.go")

		err := os.WriteFile(inputFile, []byte("package main\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create input file: %v", err)
		}

		var output bytes.Buffer
		err = writeMarkdown(inputFile, "test.go", &output, "go", testMaxFileSize)
		if err != nil {
			t.Errorf("writeMarkdown() error: %v", err)
		}

		contentStr := output.String()
		if !strings.HasPrefix(contentStr, "# test.go\n") {
			t.Errorf("Header should use display path, got: %s", contentStr[:40])
		}
		if strings.Contains(contentStr, "# "+tempDir) {
			t.Error("Header should not contain absolute path")
		}
	})

	t.Run("handles non-existent file", func(t *testing.T) {
		var output bytes.Buffer
		err := writeMarkdown("/tmp/nonexistent/file.go", "file.go", &output, "go", testMaxFileSize)
		if err == nil {
			t.Error("writeMarkdown() should error for non-existent file")
		}
	})
}

func TestProcessDirectory(t *testing.T) {
	t.Run("processes go files", func(t *testing.T) {
		tempDir := t.TempDir()
		err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		err = os.WriteFile(filepath.Join(tempDir, "util.go"), []byte("package main\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		var output bytes.Buffer
		opts := Options{
			InputFolder:      tempDir,
			AllowedLanguages: map[string]bool{".go": true},
			AllowedFileNames: map[string]bool{"go.mod": true},
			IgnorePatterns:   patternMatcher.CompilePatterns([]string{}),
			MaxFileSize:      testMaxFileSize,
		}

		err = ProcessDirectory(opts, &output)
		if err != nil {
			t.Errorf("ProcessDirectory() error: %v", err)
		}

		contentStr := output.String()

		if !strings.Contains(contentStr, "main.go") {
			t.Error("Output should contain main.go")
		}
		if !strings.Contains(contentStr, "util.go") {
			t.Error("Output should contain util.go")
		}
	})

	t.Run("ignores patterns", func(t *testing.T) {
		tempDir := t.TempDir()
		err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		err = os.WriteFile(filepath.Join(tempDir, "test.txt"), []byte("test\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		var output bytes.Buffer
		opts := Options{
			InputFolder:      tempDir,
			AllowedLanguages: map[string]bool{".go": true, ".txt": true},
			AllowedFileNames: map[string]bool{},
			IgnorePatterns:   patternMatcher.CompilePatterns([]string{"*.txt"}),
			MaxFileSize:      testMaxFileSize,
		}

		err = ProcessDirectory(opts, &output)
		if err != nil {
			t.Errorf("ProcessDirectory() error: %v", err)
		}

		contentStr := output.String()

		if !strings.Contains(contentStr, "main.go") {
			t.Error("Output should contain main.go")
		}
		if strings.Contains(contentStr, "test.txt") {
			t.Error("Output should not contain ignored test.txt")
		}
	})

	t.Run("returns error for empty directory", func(t *testing.T) {
		tempDir := t.TempDir()

		var output bytes.Buffer
		opts := Options{
			InputFolder:      tempDir,
			AllowedLanguages: map[string]bool{".go": true},
			AllowedFileNames: map[string]bool{},
			IgnorePatterns:   patternMatcher.CompilePatterns([]string{}),
			MaxFileSize:      testMaxFileSize,
		}

		err := ProcessDirectory(opts, &output)
		if err == nil {
			t.Error("ProcessDirectory() should error for empty directory")
		}
		if !strings.Contains(err.Error(), "empty") {
			t.Errorf("Error should mention empty, got: %v", err)
		}
	})

	t.Run("returns error for non-existent directory", func(t *testing.T) {
		var output bytes.Buffer
		opts := Options{
			InputFolder:      "/nonexistent/directory",
			AllowedLanguages: map[string]bool{".go": true},
			AllowedFileNames: map[string]bool{},
			IgnorePatterns:   patternMatcher.CompilePatterns([]string{}),
			MaxFileSize:      testMaxFileSize,
		}

		err := ProcessDirectory(opts, &output)
		if err == nil {
			t.Error("ProcessDirectory() should error for non-existent directory")
		}
	})

	t.Run("skips ignored directories", func(t *testing.T) {
		tempDir := t.TempDir()
		vendorDir := filepath.Join(tempDir, "vendor")
		os.Mkdir(vendorDir, 0755)
		err := os.WriteFile(filepath.Join(tempDir, "main.go"), []byte("package main\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		err = os.WriteFile(filepath.Join(vendorDir, "dep.go"), []byte("package dep\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		var output bytes.Buffer
		opts := Options{
			InputFolder:      tempDir,
			AllowedLanguages: map[string]bool{".go": true},
			AllowedFileNames: map[string]bool{},
			IgnorePatterns:   patternMatcher.CompilePatterns([]string{"vendor/"}),
			MaxFileSize:      testMaxFileSize,
		}

		err = ProcessDirectory(opts, &output)
		if err != nil {
			t.Errorf("ProcessDirectory() error: %v", err)
		}

		contentStr := output.String()

		if !strings.Contains(contentStr, "main.go") {
			t.Error("Output should contain main.go")
		}
		if strings.Contains(contentStr, "dep.go") {
			t.Error("Output should not contain vendor/dep.go")
		}
	})
}
