package patternMatcher

import (
	"testing"
)

func TestPathPatternMatching(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		isDir   bool
		pattern string
		want    bool
	}{
		{"exact match", "example.js", false, "example.js", true},
		{"exact mismatch", "example.js", false, "example.txt", false},
		{"simple wildcard match", "example.txt", false, "*.txt", true},
		{"simple wildcard mismatch", "example.txt", false, "*.md", false},
		{"unanchored wildcard matches nested", "sub/dir/example.txt", false, "*.txt", true},
		{"unanchored name matches nested", "sub/build/output.js", false, "build", true},
		{"path wildcard match", "example/file.txt", false, "example/*.txt", true},
		{"path wildcard mismatch", "example/file.txt", false, "example/*.md", false},
		{"anchored wildcard not nested", "sub/example/file.txt", false, "example/*.txt", false},
		{"nested wildcard match", "example/nested/file.txt", false, "example/*/file.txt", true},
		{"double extension match", "example.tar.gz", false, "*.tar.*", true},
		{"extension match", "example.tar.gz", false, "*.gz", true},
		{"complex extension match", "file.with.multiple.dots.tar.gz", false, "*.tar.gz", true},
		{"min.css globstar match", "style.min.css", false, "**.min.css", true},
		{"nested min.css match", "nested/style.min.css", false, "**.min.css", true},
		{"deeply nested min.css", "a/b/c/d/main.min.css", false, "**.min.css", true},
		{"regular css mismatch", "style.css", false, "**.min.css", false},
		{"globstar path match", "example/nested/file.txt", false, "**/file.txt", true},
		{"globstar root match", "file.txt", false, "**/file.txt", true},
		{"trailing globstar match", "foo/a/b.js", false, "foo/**", true},
		{"trailing globstar not dir itself", "foo", true, "foo/**", false},
		{"middle globstar match", "a/x/y/b/file.go", false, "a/**/b/file.go", true},
		{"middle globstar zero dirs", "a/b/file.go", false, "a/**/b/file.go", true},
		{"middle globstar mismatch", "a/x/c/file.go", false, "a/**/b/file.go", false},
		{"anchored glob match", "foo/x.js", false, "/foo/*.js", true},
		{"anchored glob nested mismatch", "sub/foo/x.js", false, "/foo/*.js", false},
		{"directory prefix match", "ignore/file.txt", false, "ignore/", true},
		{"directory prefix nested", "ignore/nested/file.txt", false, "ignore/", true},
		{"directory prefix matches dir itself", "ignore", true, "ignore/", true},
		{"directory prefix skips plain file", "ignore", false, "ignore/", false},
		{"directory prefix mismatch", "example.txt", false, "ignore/", false},
		{"dir pattern matches nested dir", "sub/dist/x.js", false, "dist/", true},
		{"slash prefix match", "example/file.txt", false, "/example/", true},
		{"slash prefix nested match", "example/nested/file.txt", false, "/example/", true},
		{"slash prefix mismatch", "example.txt", false, "/example/", false},
		{"slash prefix other dir", "other/file.txt", false, "/example/", false},
		{"slash prefix no trailing slash file", "vendor/file.go", false, "/vendor", true},
		{"slash prefix no trailing slash nested", "vendor/pkg/lib.go", false, "/vendor", true},
		{"slash prefix no trailing slash exact", "vendor", true, "/vendor", true},
		{"slash prefix no trailing slash mismatch", "src/vendor/file.go", false, "/vendor", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns := CompilePatterns([]string{tt.pattern})
			if got := IsPathIgnored(tt.path, tt.isDir, patterns); got != tt.want {
				t.Errorf("IsPathIgnored(%q, isDir=%v, pattern=%q) = %v; want %v", tt.path, tt.isDir, tt.pattern, got, tt.want)
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
		{"md file matches no pattern", "example.md", false},
		{"temp log ignored", "temp/file.log", true},
		{"nested temp log allowed", "temp/nested/file.log", false},
		{"nested txt ignored by other pattern", "temp/file.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPathIgnored(tt.path, false, patterns); got != tt.want {
				t.Errorf("IsPathIgnored(%q, %v) = %v; want %v", tt.path, patterns, got, tt.want)
			}
		})
	}
}

func TestNegationPatterns(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		path     string
		isDir    bool
		want     MatchResult
	}{
		{"negation re-includes", []string{"*.log", "!important.log"}, "important.log", false, Reincluded},
		{"negation re-includes nested", []string{"*.log", "!important.log"}, "sub/important.log", false, Reincluded},
		{"other files stay ignored", []string{"*.log", "!important.log"}, "other.log", false, Ignored},
		{"last match wins", []string{"!important.log", "*.log"}, "important.log", false, Ignored},
		{"negation of dir does not re-include contents", []string{"*.log", "!logs"}, "logs/app.log", false, Ignored},
		{"negation alone matches nothing positive", []string{"!keep.go"}, "other.go", false, NoMatch},
		{"anchored negation", []string{"docs/", "!docs/api.md"}, "docs/api.md", false, Reincluded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns := CompilePatterns(tt.patterns)
			if got := MatchPath(tt.path, tt.isDir, patterns); got != tt.want {
				t.Errorf("MatchPath(%q, %v) = %v; want %v", tt.path, tt.patterns, got, tt.want)
			}
		})
	}
}

func TestCompiledPatterns(t *testing.T) {
	patterns := CompilePatterns([]string{"*.txt", "ignore/", "temp/*.log", "**.min.css", "/vendor/", "/build", "!keep.go"})

	if len(patterns) != 7 {
		t.Fatalf("Expected 7 compiled patterns, got %d", len(patterns))
	}

	for _, p := range patterns {
		switch p.original {
		case "ignore/", "/vendor/":
			if !p.dirOnly {
				t.Errorf("%s should be marked as directory-only", p.original)
			}
		case "!keep.go":
			if !p.negated {
				t.Error("!keep.go should be marked as negated")
			}
		default:
			if p.dirOnly || p.negated {
				t.Errorf("%s should be neither directory-only nor negated", p.original)
			}
		}
	}
}

func TestInvalidPatternsSkipped(t *testing.T) {
	patterns := CompilePatterns([]string{"[invalid", "*.txt"})

	if len(patterns) != 1 {
		t.Fatalf("Expected invalid pattern to be skipped, got %d patterns", len(patterns))
	}
	if patterns[0].original != "*.txt" {
		t.Errorf("Expected *.txt pattern, got %s", patterns[0].original)
	}
}

func TestEmptyPatterns(t *testing.T) {
	patterns := CompilePatterns([]string{"", "*.txt", "", "/", "!"})

	if len(patterns) != 1 {
		t.Errorf("Expected 1 compiled pattern (empty patterns should be skipped), got %d", len(patterns))
	}

	if patterns[0].original != "*.txt" {
		t.Errorf("Expected *.txt pattern, got %s", patterns[0].original)
	}
}
