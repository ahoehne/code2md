package processor

import (
	"bytes"
	"code2md/patternMatcher"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
		err = WriteMarkdown(inputFile, &output, "go", 100*1024*1024)
		if err != nil {
			t.Errorf("WriteMarkdown() error: %v", err)
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
		err = WriteMarkdown(inputFile, &output, "md", 100*1024*1024)
		if err != nil {
			t.Errorf("WriteMarkdown() error: %v", err)
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
		err = WriteMarkdown(inputFile, &output, "go", 100)
		if err != nil {
			t.Errorf("WriteMarkdown() should not error for large files: %v", err)
		}

		if output.Len() != 0 {
			t.Error("Output should be empty for skipped large files")
		}
	})

	t.Run("handles non-existent file", func(t *testing.T) {
		var output bytes.Buffer
		err := WriteMarkdown("/tmp/nonexistent/file.go", &output, "go", 100*1024*1024)
		if err == nil {
			t.Error("WriteMarkdown() should error for non-existent file")
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
			MaxFileSize:      100 * 1024 * 1024,
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
			MaxFileSize:      100 * 1024 * 1024,
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
			MaxFileSize:      100 * 1024 * 1024,
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
			MaxFileSize:      100 * 1024 * 1024,
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
			MaxFileSize:      100 * 1024 * 1024,
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
