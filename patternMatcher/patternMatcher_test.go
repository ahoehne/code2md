package patternMatcher

import (
	"testing"
)

func TestPathPatternMatching(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		pattern string
		want    bool
	}{
		{"exact match", "example.js", "example.js", true},
		{"exact mismatch", "example.js", "example.txt", false},
		{"simple wildcard match", "example.txt", "*.txt", true},
		{"simple wildcard mismatch", "example.txt", "*.md", false},
		{"path wildcard match", "example/file.txt", "example/*.txt", true},
		{"path wildcard mismatch", "example/file.txt", "example/*.md", false},
		{"nested wildcard match", "example/nested/file.txt", "example/*/file.txt", true},
		{"double extension match", "example.tar.gz", "*.tar.*", true},
		{"extension match", "example.tar.gz", "*.gz", true},
		{"complex extension match", "file.with.multiple.dots.tar.gz", "*.tar.gz", true},
		{"gz extension only", "file.with.multiple.dots.tar.gz", "*.gz", true},
		{"tar wildcard match", "file.with.multiple.dots.tar.gz", "*.tar.*", true},
		{"min.css globstar match", "style.min.css", "**.min.css", true},
		{"nested min.css match", "nested/style.min.css", "**.min.css", true},
		{"deeply nested min.css", "a/b/c/d/main.min.css", "**.min.css", true},
		{"regular css mismatch", "style.css", "**.min.css", false},
		{"globstar path match", "example/nested/file.txt", "**/file.txt", true},
		{"directory prefix match", "ignore/file.txt", "ignore/", true},
		{"directory prefix nested", "ignore/nested/file.txt", "ignore/", true},
		{"directory prefix mismatch", "example.txt", "ignore/", false},
		{"slash prefix match", "example/file.txt", "/example/", true},
		{"slash prefix nested match", "example/nested/file.txt", "/example/", true},
		{"slash prefix mismatch", "example.txt", "/example/", false},
		{"slash prefix other dir", "other/file.txt", "/example/", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns := CompilePatterns([]string{tt.pattern})
			if got := IsPathIgnored(tt.path, patterns); got != tt.want {
				t.Errorf("IsPathIgnored(%q, pattern=%q) = %v; want %v", tt.path, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestPathIgnoring(t *testing.T) {
	patterns := CompilePatterns([]string{"*.txt", "ignore/", "temp/*.log"})
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"txt file ignored", "example.txt", true},
		{"ignore directory", "ignore/file.txt", true},
		{"md file allowed", "example.md", false},
		{"nested ignore", "ignore/nested/file.txt", true},
		{"json allowed", "notignored.json", false},
		{"temp log ignored", "temp/file.log", true},
		{"nested temp log allowed", "temp/nested/file.log", false},
		{"temp txt allowed", "temp/file.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPathIgnored(tt.path, patterns); got != tt.want {
				t.Errorf("IsPathIgnored(%q, %v) = %v; want %v", tt.path, patterns, got, tt.want)
			}
		})
	}
}

func TestSlashPrefixPatterns(t *testing.T) {
	patterns := CompilePatterns([]string{"/vendor/", "/node_modules/"})
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"vendor root", "vendor/package/file.go", true},
		{"node_modules root", "node_modules/lib/index.js", true},
		{"nested vendor allowed", "src/vendor/file.go", false},
		{"other directory", "src/main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPathIgnored(tt.path, patterns); got != tt.want {
				t.Errorf("IsPathIgnored(%q) = %v; want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestCompiledPatterns(t *testing.T) {
	patterns := CompilePatterns([]string{"*.txt", "ignore/", "temp/*.log", "**.min.css", "/vendor/"})

	if len(patterns) != 5 {
		t.Errorf("Expected 5 compiled patterns, got %d", len(patterns))
	}

	for _, p := range patterns {
		switch p.original {
		case "*.txt":
			if !p.isSimpleGlob {
				t.Error("*.txt should be marked as simple glob")
			}
		case "ignore/":
			if !p.isDirPrefix {
				t.Error("ignore/ should be marked as directory prefix")
			}
		case "temp/*.log":
			if !p.isSimpleGlob {
				t.Error("temp/*.log should be marked as simple glob")
			}
		case "**.min.css":
			if !p.isGlobstar {
				t.Error("**.min.css should be marked as globstar")
			}
		case "/vendor/":
			if !p.isSlashPrefix {
				t.Error("/vendor/ should be marked as slash prefix")
			}
			if p.prefix != "vendor/" {
				t.Errorf("/vendor/ prefix should be 'vendor/', got %q", p.prefix)
			}
		}
	}
}

func TestMatchGlobstar(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		pattern string
		want    bool
	}{
		{"min.css globstar", "style.min.css", "**.min.css", true},
		{"nested min.css", "nested/style.min.css", "**.min.css", true},
		{"deeply nested", "a/b/c/d/main.min.css", "**.min.css", true},
		{"regular css", "style.css", "**.min.css", false},
		{"globstar with path", "src/node_modules/file.js", "**/node_modules/file.js", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchGlobstar(tt.path, tt.pattern); got != tt.want {
				t.Errorf("matchGlobstar(%q, %q) = %v; want %v", tt.path, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestEmptyPatterns(t *testing.T) {
	patterns := CompilePatterns([]string{"", "*.txt", ""})

	if len(patterns) != 1 {
		t.Errorf("Expected 1 compiled pattern (empty patterns should be skipped), got %d", len(patterns))
	}

	if patterns[0].original != "*.txt" {
		t.Errorf("Expected *.txt pattern, got %s", patterns[0].original)
	}
}
