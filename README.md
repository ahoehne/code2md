[![GitHub License](https://img.shields.io/github/license/ahoehne/code2md)](https://github.com/ahoehne/code2md/blob/main/LICENSE)
[![GitHub Release](https://img.shields.io/github/v/release/ahoehne/code2md?include_prereleases)](https://github.com/ahoehne/code2md/releases)

[![Go CLI Test](https://github.com/ahoehne/code2md/actions/workflows/go-tests.yaml/badge.svg)](https://github.com/ahoehne/code2md/actions/workflows/go-tests.yaml)
[![CI Shellcheck](https://github.com/ahoehne/code2md/actions/workflows/shellcheck.yml/badge.svg)](https://github.com/ahoehne/code2md/actions/workflows/shellcheck.yml)

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

- `-i, --input`: Input folder to scan (required).
- `-o, --output`: Output Markdown file (required).
- `-l, --languages`: Comma-separated list of allowed languages (default: `.php,.go,.js,.ts,.py,.sh`).
- `-I, --ignore`: Comma-separated list of files and/or search patterns to ignore (default: `*.yaml,*.yml`).

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
