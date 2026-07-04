package patternMatcher

import (
	"fmt"
	"os"
	"path"
	"strings"
)

type CompiledPattern struct {
	original string
	negated  bool
	dirOnly  bool
	segments []string
}

type MatchResult int

const (
	NoMatch MatchResult = iota
	Ignored
	Reincluded
)

func CompilePatterns(patterns []string) []CompiledPattern {
	compiled := make([]CompiledPattern, 0, len(patterns))

	for _, pattern := range patterns {
		if cp, ok := compilePattern(pattern); ok {
			compiled = append(compiled, cp)
		}
	}

	return compiled
}

func compilePattern(pattern string) (CompiledPattern, bool) {
	cp := CompiledPattern{original: pattern}
	rest := pattern

	if strings.HasPrefix(rest, "!") {
		cp.negated = true
		rest = rest[1:]
	}

	if strings.HasSuffix(rest, "/") {
		cp.dirOnly = true
		rest = strings.TrimSuffix(rest, "/")
	}

	anchored := strings.Contains(rest, "/")
	rest = strings.Trim(rest, "/")
	if rest == "" {
		return CompiledPattern{}, false
	}

	segments := strings.Split(rest, "/")
	for i, seg := range segments {
		if seg == "**" {
			continue
		}
		segments[i] = collapseDoubleStarsToSingle(seg)
		if _, err := path.Match(segments[i], ""); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: invalid ignore pattern %q, skipping\n", pattern)
			return CompiledPattern{}, false
		}
	}

	if !anchored && segments[0] != "**" {
		segments = append([]string{"**"}, segments...)
	}

	cp.segments = segments
	return cp, true
}

func collapseDoubleStarsToSingle(segment string) string {
	for strings.Contains(segment, "**") {
		segment = strings.ReplaceAll(segment, "**", "*")
	}
	return segment
}

func MatchPath(p string, isDir bool, patterns []CompiledPattern) MatchResult {
	lastMatch := NoMatch
	pathSegments := strings.Split(p, "/")

	for _, cp := range patterns {
		if patternMatches(cp, pathSegments, isDir) {
			if cp.negated {
				lastMatch = Reincluded
			} else {
				lastMatch = Ignored
			}
		}
	}

	return lastMatch
}

func IsPathIgnored(p string, isDir bool, patterns []CompiledPattern) bool {
	return MatchPath(p, isDir, patterns) == Ignored
}

func patternMatches(cp CompiledPattern, pathSegments []string, isDir bool) bool {
	wholePath := len(pathSegments)

	if cp.negated {
		return matchesLeadingSegments(cp, pathSegments, wholePath, isDir)
	}

	for prefixLen := wholePath; prefixLen >= 1; prefixLen-- {
		if matchesLeadingSegments(cp, pathSegments, prefixLen, isDir) {
			return true
		}
	}

	return false
}

func matchesLeadingSegments(cp CompiledPattern, pathSegments []string, prefixLen int, isDir bool) bool {
	prefixIsDir := prefixLen < len(pathSegments) || isDir
	if cp.dirOnly && !prefixIsDir {
		return false
	}
	return matchSegments(cp.segments, pathSegments[:prefixLen])
}

func matchSegments(patternSegments, pathSegments []string) bool {
	if len(patternSegments) == 0 {
		return len(pathSegments) == 0
	}

	if patternSegments[0] == "**" {
		if len(patternSegments) == 1 {
			return len(pathSegments) > 0
		}
		for i := 0; i <= len(pathSegments); i++ {
			if matchSegments(patternSegments[1:], pathSegments[i:]) {
				return true
			}
		}
		return false
	}

	if len(pathSegments) == 0 {
		return false
	}

	if ok, _ := path.Match(patternSegments[0], pathSegments[0]); !ok {
		return false
	}

	return matchSegments(patternSegments[1:], pathSegments[1:])
}
