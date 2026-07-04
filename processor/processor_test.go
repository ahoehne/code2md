package processor

import (
	"bytes"
	"code2md/patternMatcher"
	"os"
	"path/filepath"
	"reflect"
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

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
}

func TestProcessDirectory(t *testing.T) {
	t.Run("processes go files", func(t *testing.T) {
		tempDir := t.TempDir()
		writeTestFile(t, filepath.Join(tempDir, "main.go"), "package main\n")
		writeTestFile(t, filepath.Join(tempDir, "util.go"), "package main\n")

		var output bytes.Buffer
		opts := Options{
			InputFolder:      tempDir,
			AllowedLanguages: map[string]bool{".go": true},
			AllowedFileNames: map[string]bool{"go.mod": true},
			MaxFileSize:      testMaxFileSize,
		}

		err := ProcessDirectory(opts, &output)
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
		writeTestFile(t, filepath.Join(tempDir, "main.go"), "package main\n")
		writeTestFile(t, filepath.Join(tempDir, "test.txt"), "test\n")
		writeTestFile(t, filepath.Join(tempDir, "sub", "nested.txt"), "test\n")

		var output bytes.Buffer
		opts := Options{
			InputFolder:        tempDir,
			AllowedLanguages:   map[string]bool{".go": true, ".txt": true},
			AllowedFileNames:   map[string]bool{},
			UserIgnorePatterns: patternMatcher.CompilePatterns([]string{"*.txt"}),
			MaxFileSize:        testMaxFileSize,
		}

		err := ProcessDirectory(opts, &output)
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
		if strings.Contains(contentStr, "nested.txt") {
			t.Error("Output should not contain nested ignored sub/nested.txt")
		}
	})

	t.Run("returns error for empty directory", func(t *testing.T) {
		tempDir := t.TempDir()

		var output bytes.Buffer
		opts := Options{
			InputFolder:      tempDir,
			AllowedLanguages: map[string]bool{".go": true},
			AllowedFileNames: map[string]bool{},
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
			MaxFileSize:      testMaxFileSize,
		}

		err := ProcessDirectory(opts, &output)
		if err == nil {
			t.Error("ProcessDirectory() should error for non-existent directory")
		}
	})

	t.Run("skips ignored directories", func(t *testing.T) {
		tempDir := t.TempDir()
		writeTestFile(t, filepath.Join(tempDir, "main.go"), "package main\n")
		writeTestFile(t, filepath.Join(tempDir, "vendor", "dep.go"), "package dep\n")

		var output bytes.Buffer
		opts := Options{
			InputFolder:        tempDir,
			AllowedLanguages:   map[string]bool{".go": true},
			AllowedFileNames:   map[string]bool{},
			UserIgnorePatterns: patternMatcher.CompilePatterns([]string{"vendor/"}),
			MaxFileSize:        testMaxFileSize,
		}

		err := ProcessDirectory(opts, &output)
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

	t.Run("skips .git directory", func(t *testing.T) {
		tempDir := t.TempDir()
		writeTestFile(t, filepath.Join(tempDir, "main.go"), "package main\n")
		writeTestFile(t, filepath.Join(tempDir, ".git", "refs", "heads", "feature.go"), "abc123\n")

		var output bytes.Buffer
		opts := Options{
			InputFolder:      tempDir,
			AllowedLanguages: map[string]bool{".go": true},
			AllowedFileNames: map[string]bool{},
			MaxFileSize:      testMaxFileSize,
		}

		err := ProcessDirectory(opts, &output)
		if err != nil {
			t.Errorf("ProcessDirectory() error: %v", err)
		}

		if strings.Contains(output.String(), "feature.go") {
			t.Error("Output should not contain files from .git")
		}
	})

	t.Run("skips output file by absolute path", func(t *testing.T) {
		tempDir := t.TempDir()
		writeTestFile(t, filepath.Join(tempDir, "main.go"), "package main\n")
		writeTestFile(t, filepath.Join(tempDir, "out.md"), "previous run output\n")

		var output bytes.Buffer
		opts := Options{
			InputFolder:       tempDir,
			AllowedLanguages:  map[string]bool{".go": true, ".md": true},
			AllowedFileNames:  map[string]bool{},
			AbsOutputFilePath: filepath.Join(tempDir, "out.md"),
			MaxFileSize:       testMaxFileSize,
		}

		err := ProcessDirectory(opts, &output)
		if err != nil {
			t.Errorf("ProcessDirectory() error: %v", err)
		}

		if strings.Contains(output.String(), "previous run output") {
			t.Error("Output should not contain the output file itself")
		}
	})

	t.Run("allowedFileNames bypass default ignore patterns", func(t *testing.T) {
		tempDir := t.TempDir()
		writeTestFile(t, filepath.Join(tempDir, "pom.xml"), "<project/>")
		writeTestFile(t, filepath.Join(tempDir, "other.xml"), "<root/>")

		var output bytes.Buffer
		opts := Options{
			InputFolder:           tempDir,
			AllowedLanguages:      map[string]bool{".java": true, ".xml": false},
			AllowedFileNames:      map[string]bool{"pom.xml": true},
			DefaultIgnorePatterns: patternMatcher.CompilePatterns([]string{"*.xml"}),
			MaxFileSize:           testMaxFileSize,
		}

		err := ProcessDirectory(opts, &output)
		if err != nil {
			t.Errorf("ProcessDirectory() error: %v", err)
		}

		contentStr := output.String()
		if !strings.Contains(contentStr, "pom.xml") {
			t.Error("pom.xml should bypass default *.xml ignore via allowedFileNames")
		}
		if strings.Contains(contentStr, "other.xml") {
			t.Error("other.xml should still be ignored")
		}
	})

	t.Run("explicit ignore patterns win over allowedFileNames", func(t *testing.T) {
		tempDir := t.TempDir()
		writeTestFile(t, filepath.Join(tempDir, "main.go"), "package main\n")
		writeTestFile(t, filepath.Join(tempDir, "pom.xml"), "<project/>")

		var output bytes.Buffer
		opts := Options{
			InputFolder:        tempDir,
			AllowedLanguages:   map[string]bool{".go": true, ".java": true},
			AllowedFileNames:   map[string]bool{"pom.xml": true},
			UserIgnorePatterns: patternMatcher.CompilePatterns([]string{"pom.xml"}),
			MaxFileSize:        testMaxFileSize,
		}

		err := ProcessDirectory(opts, &output)
		if err != nil {
			t.Errorf("ProcessDirectory() error: %v", err)
		}

		if strings.Contains(output.String(), "pom.xml") {
			t.Error("explicitly ignored pom.xml should not be in output")
		}
	})

	t.Run("respects root gitignore", func(t *testing.T) {
		tempDir := t.TempDir()
		writeTestFile(t, filepath.Join(tempDir, ".gitignore"), "generated.go\nbuild/\n")
		writeTestFile(t, filepath.Join(tempDir, "main.go"), "package main\n")
		writeTestFile(t, filepath.Join(tempDir, "generated.go"), "package main // generated\n")
		writeTestFile(t, filepath.Join(tempDir, "sub", "generated.go"), "package sub // generated\n")
		writeTestFile(t, filepath.Join(tempDir, "build", "artifact.go"), "package build\n")

		var output bytes.Buffer
		opts := Options{
			InputFolder:      tempDir,
			AllowedLanguages: map[string]bool{".go": true},
			AllowedFileNames: map[string]bool{},
			MaxFileSize:      testMaxFileSize,
		}

		err := ProcessDirectory(opts, &output)
		if err != nil {
			t.Errorf("ProcessDirectory() error: %v", err)
		}

		contentStr := output.String()
		if !strings.Contains(contentStr, "main.go") {
			t.Error("Output should contain main.go")
		}
		if strings.Contains(contentStr, "generated.go") {
			t.Error("Output should not contain gitignored generated.go (any depth)")
		}
		if strings.Contains(contentStr, "artifact.go") {
			t.Error("Output should not contain files from gitignored build/")
		}
	})

	t.Run("respects nested gitignore scoped to its directory", func(t *testing.T) {
		tempDir := t.TempDir()
		writeTestFile(t, filepath.Join(tempDir, "main.go"), "package main\n")
		writeTestFile(t, filepath.Join(tempDir, "sub", ".gitignore"), "local.go\n")
		writeTestFile(t, filepath.Join(tempDir, "sub", "local.go"), "package sub\n")
		writeTestFile(t, filepath.Join(tempDir, "local.go"), "package main // root\n")

		var output bytes.Buffer
		opts := Options{
			InputFolder:      tempDir,
			AllowedLanguages: map[string]bool{".go": true},
			AllowedFileNames: map[string]bool{},
			MaxFileSize:      testMaxFileSize,
		}

		err := ProcessDirectory(opts, &output)
		if err != nil {
			t.Errorf("ProcessDirectory() error: %v", err)
		}

		contentStr := output.String()
		if strings.Contains(contentStr, "sub/local.go") {
			t.Error("Output should not contain sub/local.go ignored by nested .gitignore")
		}
		if !strings.Contains(contentStr, "# local.go") {
			t.Error("Output should contain root local.go (nested .gitignore must not apply outside its directory)")
		}
	})

	t.Run("gitignore negation re-includes file", func(t *testing.T) {
		tempDir := t.TempDir()
		writeTestFile(t, filepath.Join(tempDir, ".gitignore"), "*.go\n!keep.go\n")
		writeTestFile(t, filepath.Join(tempDir, "skip.go"), "package main\n")
		writeTestFile(t, filepath.Join(tempDir, "keep.go"), "package main // keep\n")

		var output bytes.Buffer
		opts := Options{
			InputFolder:      tempDir,
			AllowedLanguages: map[string]bool{".go": true},
			AllowedFileNames: map[string]bool{},
			MaxFileSize:      testMaxFileSize,
		}

		err := ProcessDirectory(opts, &output)
		if err != nil {
			t.Errorf("ProcessDirectory() error: %v", err)
		}

		contentStr := output.String()
		if strings.Contains(contentStr, "skip.go") {
			t.Error("Output should not contain gitignored skip.go")
		}
		if !strings.Contains(contentStr, "keep.go") {
			t.Error("Output should contain re-included keep.go")
		}
	})
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

}
