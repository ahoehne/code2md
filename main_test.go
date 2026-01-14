package main

import (
	"bytes"
	"code2md/c2mConfig"
	"code2md/patternMatcher"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsFileAllowed(t *testing.T) {
	allowedLanguages := map[string]bool{
		".php":  true,
		".go":   true,
		".js":   true,
		".ts":   true,
		".java": true,
		".json": false,
	}

	allowedFileNames := map[string]bool{
		"go.mod":        true,
		"composer.json": true,
		"package.json":  true,
		"pom.xml":       true,
	}

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"php file", "file.php", true},
		{"go file", "file.go", true},
		{"js file", "file.js", true},
		{"ts file", "file.ts", true},
		{"java file", "file.java", true},
		{"java file uppercase", "File.java", true},
		{"special file pom.xml", "pom.xml", true},
		{"txt file not allowed", "file.txt", false},
		{"special file go.mod", "go.mod", true},
		{"special file composer.json", "composer.json", true},
		{"multi-dot js", "file.with.multiple.dots.js", true},
		{"no extension", "file_without_extension", false},
		{"multi-dot php", "file.with.multiple.dots.php", true},
		{"multi-dot go", "file.with.multiple.dots.go", true},
		{"multi-dot ts", "file.with.multiple.dots.ts", true},
		{"json file disabled", "file.json", false},
		{"hidden go file", ".hidden.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isFileAllowed(tt.filename, allowedLanguages, allowedFileNames); got != tt.want {
				t.Errorf("isFileAllowed(%q) = %v; want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestGetMdLang(t *testing.T) {
	allowedFileNames := map[string]bool{
		"go.mod": true,
	}

	specialFileLanguages := map[string]string{
		"go.mod": "go",
	}

	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{"go file", "main.go", "go"},
		{"js file", "script.js", "js"},
		{"go.mod special", "go.mod", "go"},
		{"no extension", "README", "plaintext"},
		{"md file", "README.md", "md"},
		{"php file", "index.php", "php"},
		{"ts file", "app.ts", "ts"},
		{"java file", "Main.java", "java"},
		{"multi-dot extension", "file.test.js", "js"},
		{"hidden file with extension", ".eslintrc.json", "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMdLang(tt.filename, allowedFileNames, specialFileLanguages); got != tt.want {
				t.Errorf("getMdLang(%q) = %v; want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestWriteMarkdown(t *testing.T) {
	t.Run("writes go file with code fence", func(t *testing.T) {
		tempDir := t.TempDir()
		inputFile := filepath.Join(tempDir, "test.go")
		outputFile := filepath.Join(tempDir, "output.md")

		err := os.WriteFile(inputFile, []byte(
			"package main\nfunc main() {}\n"), 0644)
		if err != nil {
			t.Fatalf("Failed to create input file: %v", err)
		}

		output, err := os.Create(outputFile)
		if err != nil {
			t.Fatalf("Failed to create output file: %v", err)
		}

		err = writeMarkdown(inputFile, output, "go", 100*1024*1024)
		output.Close()
		if err != nil {
			t.Errorf("writeMarkdown() error: %v", err)
		}

		content, _ := os.ReadFile(outputFile)
		contentStr := string(content)

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
		outputFile := filepath.Join(tempDir, "output.md")

		err := os.WriteFile(inputFile, []byte(
			"# Example Markdown Heading\nMarkdown Content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create input file: %v", err)
		}

		output, err := os.Create(outputFile)
		if err != nil {
			t.Fatalf("Failed to create output file: %v", err)
		}

		err = writeMarkdown(inputFile, output, "md", 100*1024*1024)
		output.Close()
		if err != nil {
			t.Errorf("writeMarkdown() error: %v", err)
		}

		content, _ := os.ReadFile(outputFile)
		contentStr := string(content)

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
		outputFile := filepath.Join(tempDir, "output.md")

		largeContent := bytes.Repeat([]byte("x"), 1024)
		err := os.WriteFile(inputFile, largeContent, 0644)
		if err != nil {
			t.Fatalf("Failed to create input file: %v", err)
		}

		output, err := os.Create(outputFile)
		if err != nil {
			t.Fatalf("Failed to create output file: %v", err)
		}

		err = writeMarkdown(inputFile, output, "go", 100)
		output.Close()
		if err != nil {
			t.Errorf("writeMarkdown() should not error for large files: %v", err)
		}

		content, _ := os.ReadFile(outputFile)
		if len(content) != 0 {
			t.Error("Output should be empty for skipped large files")
		}
	})

	t.Run("handles non-existent file", func(t *testing.T) {
		tempDir := t.TempDir()
		outputFile := filepath.Join(tempDir, "output.md")

		output, _ := os.Create(outputFile)
		defer output.Close()

		err := writeMarkdown("/tmp/nonexistent/file.go", output, "go", 100*1024*1024)
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

		outputFile := filepath.Join(tempDir, "output.md")
		output, err := os.Create(outputFile)
		if err != nil {
			t.Fatalf("Failed to create output file: %v", err)
		}

		allowedLanguages := map[string]bool{".go": true}
		allowedFileNames := map[string]bool{"go.mod": true}
		patterns := patternMatcher.CompilePatterns([]string{})

		err = processDirectory(tempDir, output, patterns, allowedLanguages, allowedFileNames, 100*1024*1024)
		output.Close()
		if err != nil {
			t.Errorf("processDirectory() error: %v", err)
		}

		content, _ := os.ReadFile(outputFile)
		contentStr := string(content)

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

		outputFile := filepath.Join(tempDir, "output.md")
		output, err := os.Create(outputFile)
		if err != nil {
			t.Fatalf("Failed to create output file: %v", err)
		}

		allowedLanguages := map[string]bool{".go": true, ".txt": true}
		allowedFileNames := map[string]bool{}
		patterns := patternMatcher.CompilePatterns([]string{"*.txt"})

		err = processDirectory(tempDir, output, patterns, allowedLanguages, allowedFileNames, 100*1024*1024)
		output.Close()
		if err != nil {
			t.Errorf("processDirectory() error: %v", err)
		}

		content, _ := os.ReadFile(outputFile)
		contentStr := string(content)

		if !strings.Contains(contentStr, "main.go") {
			t.Error("Output should contain main.go")
		}
		if strings.Contains(contentStr, "test.txt") {
			t.Error("Output should not contain ignored test.txt")
		}
	})

	t.Run("returns error for empty directory", func(t *testing.T) {
		tempDir := t.TempDir()
		outputFile := filepath.Join(tempDir, "output.md")
		output, _ := os.Create(outputFile)
		defer output.Close()

		allowedLanguages := map[string]bool{".go": true}
		allowedFileNames := map[string]bool{}
		patterns := patternMatcher.CompilePatterns([]string{})

		err := processDirectory(tempDir, output, patterns, allowedLanguages, allowedFileNames, 100*1024*1024)
		if err == nil {
			t.Error("processDirectory() should error for empty directory")
		}
		if !strings.Contains(err.Error(), "empty") {
			t.Errorf("Error should mention empty, got: %v", err)
		}
	})

	t.Run("returns error for non-existent directory", func(t *testing.T) {
		tempDir := t.TempDir()
		outputFile := filepath.Join(tempDir, "output.md")
		output, _ := os.Create(outputFile)
		defer output.Close()

		allowedLanguages := map[string]bool{".go": true}
		allowedFileNames := map[string]bool{}
		patterns := patternMatcher.CompilePatterns([]string{})

		err := processDirectory("/nonexistent/directory", output, patterns, allowedLanguages, allowedFileNames, 100*1024*1024)
		if err == nil {
			t.Error("processDirectory() should error for non-existent directory")
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

		outputFile := filepath.Join(tempDir, "output.md")
		output, err := os.Create(outputFile)
		if err != nil {
			t.Fatalf("Failed to create output file: %v", err)
		}

		allowedLanguages := map[string]bool{".go": true}
		allowedFileNames := map[string]bool{}
		patterns := patternMatcher.CompilePatterns([]string{"vendor/"})

		err = processDirectory(tempDir, output, patterns, allowedLanguages, allowedFileNames, 100*1024*1024)
		output.Close()
		if err != nil {
			t.Errorf("processDirectory() error: %v", err)
		}

		content, _ := os.ReadFile(outputFile)
		contentStr := string(content)

		if !strings.Contains(contentStr, "main.go") {
			t.Error("Output should contain main.go")
		}
		if strings.Contains(contentStr, "dep.go") {
			t.Error("Output should not contain vendor/dep.go")
		}
	})
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

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err = run(config)

		w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("run() error: %v", err)
		}

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

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
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	displayVersion()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "code2md") {
		t.Error("Output should contain 'code2md'")
	}
}

func TestDisplayUsageInstructions(t *testing.T) {
	t.Run("with nil config", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		displayUsageInstructions(nil, false)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

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

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		displayUsageInstructions(config, false)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		if !strings.Contains(output, "By default") {
			t.Error("Output should show language info with valid config")
		}
	})

	t.Run("shows error when requested", func(t *testing.T) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		displayUsageInstructions(nil, true)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		if !strings.Contains(output, "Error:") {
			t.Error("Output should contain error message")
		}
	})
}
