package patternMatcher

import (
	"path/filepath"
	"strings"
)

type CompiledPattern struct {
	original      string
	isExact       bool
	isDirPrefix   bool
	isSlashPrefix bool
	isSimpleGlob  bool
	isGlobstar    bool
	prefix        string
}

func CompilePatterns(patterns []string) []CompiledPattern {
	compiled := make([]CompiledPattern, 0, len(patterns))

	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}

		cp := CompiledPattern{original: pattern}

		if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
			cp.isSlashPrefix = true
			cp.prefix = strings.TrimPrefix(pattern, "/")
		} else if strings.HasSuffix(pattern, "/") {
			cp.isDirPrefix = true
			cp.prefix = pattern
		} else if strings.Contains(pattern, "**") {
			cp.isGlobstar = true
		} else if strings.Contains(pattern, "*") {
			cp.isSimpleGlob = true
		} else {
			cp.isExact = true
		}

		compiled = append(compiled, cp)
	}

	return compiled
}

func IsPathIgnored(path string, patterns []CompiledPattern) bool {
	for _, pattern := range patterns {
		if pattern.isExact && pattern.original == path {
			return true
		}

		if pattern.isSlashPrefix && strings.HasPrefix(path, pattern.prefix) {
			return true
		}

		if pattern.isDirPrefix && strings.HasPrefix(path, pattern.prefix) {
			return true
		}

		if pattern.isSimpleGlob {
			matched, _ := filepath.Match(pattern.original, path)
			if matched {
				return true
			}
		}

		if pattern.isGlobstar && matchGlobstar(path, pattern.original) {
			return true
		}
	}

	return false
}

func matchGlobstar(path, pattern string) bool {
	subPattern := strings.TrimPrefix(pattern, "**/")
	if subPattern == pattern {
		subPattern = strings.TrimPrefix(pattern, "**")
		_, fileName := filepath.Split(path)
		matched, _ := filepath.Match("*"+subPattern, fileName)
		return matched
	}

	parts := strings.Split(path, string(filepath.Separator))
	for i := 0; i < len(parts); i++ {
		remainingPath := strings.Join(parts[i:], string(filepath.Separator))
		matched, _ := filepath.Match(subPattern, remainingPath)
		if matched {
			return true
		}
	}
	return false
}
