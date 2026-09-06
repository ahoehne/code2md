package c2mConfig

import (
	"bytes"
	"flag"
	"os"
	"strings"
	"testing"
)

func setupFlagTest(t *testing.T, args ...string) {
	t.Helper()
	origFlags := flag.CommandLine
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	origArgs := os.Args
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tempDir)
	os.Args = append([]string{"cmd"}, args...)
	t.Cleanup(func() {
		flag.CommandLine = origFlags
		os.Args = origArgs
		os.Chdir(origDir)
	})
}

func parseConfig(t *testing.T, args ...string) *Config {
	t.Helper()
	setupFlagTest(t, args...)
	config, err := InitializeConfigFromFlags()
	if err != nil {
		t.Fatalf("InitializeConfigFromFlags() error: %v", err)
	}
	return config
}

func sliceContains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

func TestInitializeConfigFromFlags(t *testing.T) {
	t.Run("default ignore patterns present when no flags passed", func(t *testing.T) {
		config := parseConfig(t, "-i", t.TempDir())

		for _, d := range strings.Split(defaultIgnoredPatterns, ",") {
			if !sliceContains(config.DefaultIgnorePatterns, d) {
				t.Errorf("expected default ignore pattern %q in %v", d, config.DefaultIgnorePatterns)
			}
		}
		if len(config.UserIgnorePatterns) != 0 {
			t.Errorf("expected no user ignore patterns, got %v", config.UserIgnorePatterns)
		}
	})

	t.Run("explicit ignore overrides defaults", func(t *testing.T) {
		config := parseConfig(t, "-i", t.TempDir(), "--ignore", "custom.txt,other.log")

		for _, expected := range []string{"custom.txt", "other.log"} {
			if !sliceContains(config.UserIgnorePatterns, expected) {
				t.Errorf("expected ignore pattern %q in %v", expected, config.UserIgnorePatterns)
			}
		}

		for _, d := range strings.Split(defaultIgnoredPatterns, ",") {
			if sliceContains(config.UserIgnorePatterns, d) || sliceContains(config.DefaultIgnorePatterns, d) {
				t.Errorf("default pattern %q should not be present when --ignore is explicit", d)
			}
		}
	})

	t.Run("output file is not turned into an ignore pattern", func(t *testing.T) {
		config := parseConfig(t, "-i", t.TempDir(), "-o", "output.md")

		if sliceContains(config.UserIgnorePatterns, "output.md") || sliceContains(config.DefaultIgnorePatterns, "output.md") {
			t.Errorf("output file should be excluded by path, not by pattern, got %v / %v", config.UserIgnorePatterns, config.DefaultIgnorePatterns)
		}
		if config.OutputMarkdown != "output.md" {
			t.Errorf("expected OutputMarkdown 'output.md', got %q", config.OutputMarkdown)
		}
	})

	t.Run("default yml ignore dropped when yml language enabled", func(t *testing.T) {
		config := parseConfig(t, "-i", t.TempDir(), "-l", "yml,go")

		if sliceContains(config.DefaultIgnorePatterns, "*.yml") {
			t.Errorf("*.yml should be removed from defaults when yml is enabled, got %v", config.DefaultIgnorePatterns)
		}
		if !sliceContains(config.DefaultIgnorePatterns, "*.xml") {
			t.Errorf("*.xml should remain when only yml is enabled, got %v", config.DefaultIgnorePatterns)
		}
		if !sliceContains(config.DefaultIgnorePatterns, "*.yaml") {
			t.Errorf("*.yaml should remain when only yml is enabled, got %v", config.DefaultIgnorePatterns)
		}
	})

	t.Run("explicit ignore preserved even when matching language is enabled", func(t *testing.T) {
		config := parseConfig(t, "-i", t.TempDir(), "-l", "yml", "--ignore", "*.yml")

		if !sliceContains(config.UserIgnorePatterns, "*.yml") {
			t.Errorf("explicit --ignore *.yml should be preserved, got %v", config.UserIgnorePatterns)
		}
	})

	t.Run("min.css ignored by default when css enabled", func(t *testing.T) {
		config := parseConfig(t, "-i", t.TempDir(), "-l", "css")

		if !sliceContains(config.DefaultIgnorePatterns, "**.min.css") {
			t.Errorf("expected **.min.css in default ignore patterns %v", config.DefaultIgnorePatterns)
		}
	})

	t.Run("rejects positional arguments", func(t *testing.T) {
		setupFlagTest(t, "-i", t.TempDir(), "stray-argument")

		if _, err := InitializeConfigFromFlags(); err == nil {
			t.Error("expected error for unexpected positional arguments")
		}
	})
}

func TestPrintUsage(t *testing.T) {
	parseConfig(t)
	var output bytes.Buffer
	PrintUsage(&output)
	for _, want := range []string{
		"Usage: code2md -i <directory> [options]",
		"-i, --input directory", "Input directory to scan (required)",
		"-o, --output file", "Output Markdown file (default: stdout)",
		"-l, --languages names", "-I, --ignore patterns",
		"(replaces defaults: *.yaml,*.yml,*.xml)",
		"-m, --max-file-size size", "Maximum size of each file, e.g. 512KB or 10MB (default: 100MB)",
		"-h, --help", "-v, --version",
		"Default languages: c, cc, cjs, cpp, cs, cts, cxx, dockerfile, go, h, hh, hpp, java, js, jsx, mjs, mts, php, py, rs, sh, ts, tsx\n",
		"Opt-in languages (-l): css, html, json, md, scss, sql, toml, xml, yaml, yml\n",
	} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("help missing %q:\n%s", want, output.String())
		}
	}
	if strings.ContainsAny(output.String(), "|`") {
		t.Errorf("help contains Markdown formatting:\n%s", output.String())
	}
}

func TestFlagAliases(t *testing.T) {
	for _, args := range [][]string{
		{"-i", "src", "-o", "code.md", "-l", "go", "-I", "*.log", "-m", "42", "-h", "-v"},
		{"--input", "src", "--output", "code.md", "--languages", "go", "--ignore", "*.log", "--max-file-size", "42", "--help", "--version"},
	} {
		t.Run(args[0], func(t *testing.T) {
			config := parseConfig(t, args...)
			if config.InputFolder != "src" || config.OutputMarkdown != "code.md" ||
				config.MaxFileSize != 42 || !config.Help || !config.Version ||
				!config.AllowedLanguages[".go"] || config.AllowedLanguages[".js"] ||
				strings.Join(config.UserIgnorePatterns, ",") != "*.log" {
				t.Errorf("unexpected config: %+v", config)
			}
		})
	}
}

func TestMaxFileSizeParsing(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  int64
	}{
		{"42", 42},
		{"512B", 512},
		{"512kb", 512 * 1024},
		{"10MB", 10 * 1024 * 1024},
		{"1GB", 1024 * 1024 * 1024},
	} {
		t.Run(tt.input, func(t *testing.T) {
			config := parseConfig(t, "-i", "src", "-m", tt.input)
			if config.MaxFileSize != tt.want {
				t.Errorf("MaxFileSize = %d; want %d", config.MaxFileSize, tt.want)
			}
		})
	}

	for _, invalid := range []string{"0", "1.5MB", "99999999999GB"} {
		t.Run("invalid "+invalid, func(t *testing.T) {
			setupFlagTest(t, "-i", "src", "-m", invalid)
			flag.CommandLine.SetOutput(&bytes.Buffer{})
			if _, err := InitializeConfigFromFlags(); err == nil {
				t.Errorf("expected an error for size %q", invalid)
			}
		})
	}
}

func TestInvalidFlagValue(t *testing.T) {
	setupFlagTest(t, "-m", "invalid")
	var output bytes.Buffer
	flag.CommandLine.SetOutput(&output)
	if _, err := InitializeConfigFromFlags(); err == nil {
		t.Fatal("expected an error for an invalid file size")
	}
	if !strings.Contains(output.String(), "Usage: code2md -i <directory> [options]") {
		t.Errorf("parse error should show the same help:\n%s", output.String())
	}
}
