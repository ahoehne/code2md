# code2md

`code2md` is a command-line tool that converts code from a specified directory into a Markdown file. It supports multiple programming languages and allows for customization through command-line flags.

## Requirements

- Go 1.18 or later
- Make (for build and installation tasks)

## Installation

To install `code2md`, follow these steps:

1. Clone the repository:
   ```sh
   git clone https://github.com/ahoehne/code2md.git
   cd code2md
   ```

2. Build the application:
   ```sh
   make build
   ```

3. Install the application:
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