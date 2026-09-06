[![GitHub License](https://img.shields.io/github/license/ahoehne/code2md)](https://github.com/ahoehne/code2md/blob/main/LICENSE)
[![GitHub Release](https://img.shields.io/github/v/release/ahoehne/code2md?include_prereleases)](https://github.com/ahoehne/code2md/releases)

[![Go CLI Test](https://github.com/ahoehne/code2md/actions/workflows/go-tests.yaml/badge.svg)](https://github.com/ahoehne/code2md/actions/workflows/go-tests.yaml)
[![CI Shellcheck](https://github.com/ahoehne/code2md/actions/workflows/shellcheck.yaml/badge.svg)](https://github.com/ahoehne/code2md/actions/workflows/shellcheck.yaml)

# code2md

`code2md` is a command-line tool that converts code from a specified directory into a Markdown file. It supports multiple programming languages and allows for customization through command-line flags.

## Installation

### From Released Binaries
For the easiest installation, download a pre-built binary.

**Download the Binary:**
- Go to the [releases page](https://github.com/ahoehne/code2md/releases).
- Download the latest binary for your operating system.
- **Windows**
   - Rename the downloaded file to code2md.exe.
    - Place the binary in a directory that is included in your system's PATH, such as C:\Windows, or add the directory containing code2md.exe to your PATH environment variable.


- **Linux/Mac:**
   1. Rename the binary based on your architecture:
      ```sh
      mv code2md-linux-amd64 code2md    # For Linux AMD64
      mv code2md-linux-arm64 code2md    # For Linux ARM64
      mv code2md-darwin-amd64 code2md   # For macOS AMD64
      mv code2md-darwin-arm64 code2md   # For macOS ARM64
      ```
   2. make the binary executable and move it to `/usr/local/bin`
      ```sh
      chmod +x code2md
      sudo mv code2md /usr/local/bin/
      sudo chown root:root /usr/local/bin/code2md
      ```

### From Source

1. Requirements:
   Install the following dependencies
   - Git to clone this repository
   - Go 1.18 or later
   - Make (for build and installation tasks)

2. Clone the repository:
   ```sh
   git clone https://github.com/ahoehne/code2md.git
   cd code2md
   ```

3. Build the application:
   ```sh
   make build
   ```

4. Install the application:
   ```sh
   sudo make install
   ```

## Usage

### Build the Application

To build the application for multiple platforms, run:
```sh
make buildall
```

### Run Tests

To run the tests, use:
```sh
make test
```

### Example Command

To convert code from the current directory into a Markdown file named `code.md`, use the following command:
```sh
code2md -i . -o code.md
```

### Command-Line Flags

| Flag              | Short | Description                                                     |
| ----------------- | ----- | --------------------------------------------------------------- |
| `--input`         | `-i`  | Input directory to scan (required)                              |
| `--output`        | `-o`  | Output Markdown file (optional, defaults to stdout)             |
| `--languages`     | `-l`  | Comma-separated list of allowed languages (extensions or names) |
| `--ignore`        | `-I`  | Comma-separated ignore patterns                                 |
| `--max-file-size` | `-m`  | Maximum size of each file, e.g. 512KB or 10MB (default: 100MB)  |
| `--help`          | `-h`  | Show help                                                       |
| `--version`       | `-v`  | Show version information                                        |

### Ignoring Files

Files are excluded in three ways:

1. **`.gitignore` files** in the input directory (including nested ones) are respected with gitignore semantics.
2. **`--ignore` patterns** use the same syntax and replace the built-in defaults (`*.yaml,*.yml,*.xml`).
3. The `.git` directory and the output file itself are always skipped.

Pattern syntax follows `.gitignore`:

- `*.log` matches at any depth; `*` and `?` never cross `/`
- `build` matches a file or directory named `build` at any depth
- `build/` matches directories only
- `/build` or `src/build` are anchored to the input directory root
- `**` matches any number of directories (`docs/**`, `a/**/b`, `**/file.txt`)
- `!pattern` re-includes a previously ignored file; the last matching pattern wins

Language manifest files (`pom.xml`, `package.json`, `go.mod`, ...) of enabled languages bypass the built-in default patterns, but never patterns you pass via `--ignore` or `.gitignore`.

### Supported Languages

- **On by default:** `php`, `go`, `js`, `mjs`, `cjs`, `ts`, `mts`, `cts`, `jsx`, `tsx`, `py`, `sh`, `java`, `c`, `h`, `cpp`, `cc`, `cxx`, `hpp`, `hh`, `cs`, `rs`, `Dockerfile`
- **Opt-in via `-l`:** `md`, `html`, `scss`, `css`, `json`, `yaml`, `yml`, `xml`, `toml`, `sql`

## Hint: getting the generated file into clipboard
These commands copy the contents of `code.md` into the clipboard.

### Linux (xclip)

```sh
xclip -sel clip < code.md
```

### macOS (pbcopy)

```sh
pbcopy < code.md
```

### Windows (clip)

```sh
clip < code.md
```

## Contributing

Contributions are welcome. Please open an issue or submit a pull request.

## License

This project is licensed under the MIT License.
