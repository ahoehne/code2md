package main

import (
	"code2md/c2mConfig"
	"testing"
)

func TestPathPatternMatching(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    bool
	}{
		{"example.js", "example.txt", false},
		{"example.js", "example.js", true},
		{"example.txt", "/example/", false},
		{"example/file.txt", "/example/", true},
		{"example.txt", "*.txt", true},
		{"example.txt", "*.md", false},
		{"example/file.txt", "example/*.txt", true},
		{"example/file.txt", "example/*.md", false},
		{"example/nested/file.txt", "example/*/file.txt", true},
		{"example/nested/file.txt", "example/**/file.txt", true},
		{"example.tar.gz", "*.tar.*", true},
		{"example.tar.gz", "*.gz", true},
		{"file.with.multiple.dots.tar.gz", "*.tar.gz", true},
		{"file.with.multiple.dots.tar.gz", "*.gz", true},
		{"file.with.multiple.dots.tar.gz", "*.tar.*", true},
		{"style.min.css", "**.min.css", true},
		{"nested/style.min.css", "**.min.css", true},
		{"a/b/c/d/main.min.css", "**.min.css", true},
		{"style.css", "**.min.css", false},
	}

	for _, tt := range tests {
		if got := doesPathMatchPattern(tt.path, tt.pattern); got != tt.want {
			t.Errorf("doesPathMatchPattern(%q, %q) = %v; want %v", tt.path, tt.pattern, got, tt.want)
		}
	}
}

func TestPathIgnoring(t *testing.T) {
	patterns := []string{"*.txt", "ignore/", "temp/*.log"}
	tests := []struct {
		path string
		want bool
	}{
		{"example.txt", true},
		{"ignore/file.txt", true},
		{"example.md", false},
		{"ignore/nested/file.txt", true},
		{"notignored.json", false},
		{"temp/file.log", true},
		{"temp/nested/file.log", false},
		{"temp/file.txt", false},
	}

	for _, tt := range tests {
		if got := isPathIgnored(tt.path, patterns); got != tt.want {
			t.Errorf("isPathIgnored(%q, %v) = %v; want %v", tt.path, patterns, got, tt.want)
		}
	}
}

func TestIsFileAllowed(t *testing.T) {
	allowedLanguages := map[string]bool{
		".php":  true,
		".go":   true,
		".js":   true,
		".ts":   true,
		".java": true,
		".json": false,
	}

	allowedFileNames := c2mConfig.GetAllowedFileNames(allowedLanguages)

	tests := []struct {
		filename string
		want     bool
	}{
		{"file.php", true},
		{"file.go", true},
		{"file.js", true},
		{"file.ts", true},
		{"file.java", true},
		{"File.java", true},
		{"pom.xml", true},
		{"file.txt", false},
		{"go.mod", true},
		{"composer.json", true},
		{"file.with.multiple.dots.js", true},
		{"file_without_extension", false},
		{"file.with.multiple.dots.php", true},
		{"file.with.multiple.dots.go", true},
		{"file.with.multiple.dots.ts", true},
	}

	for _, tt := range tests {
		if got := isFileAllowed(tt.filename, allowedLanguages, allowedFileNames); got != tt.want {
			t.Errorf("isFileAllowed(%q, %v, %v) = %v; want %v", tt.filename, allowedLanguages, allowedFileNames, got, tt.want)
		}
	}
}
