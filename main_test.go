package main

import (
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMdLang(tt.filename, allowedFileNames, specialFileLanguages); got != tt.want {
				t.Errorf("getMdLang(%q) = %v; want %v", tt.filename, got, tt.want)
			}
		})
	}
}
