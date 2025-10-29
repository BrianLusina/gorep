# gorep

[![progress-banner](https://backend.codecrafters.io/progress/grep/4318fdb4-1f39-4457-a350-a88d03157495)](https://app.codecrafters.io/users/codecrafters-bot?r=2qF)

This is a [Go](https://go.dev/) implementation of the [grep](https://www.gnu.org/software/grep/) command line tool that utilizes regular expressions for searching text.

## Pre-requisites

### [Go](https://go.dev/)

Ensure you have Go installed locally. You can download it from the [official Go website](https://go.dev/dl/) the minimum required version is `1.24`.

## Getting started

This program uses very few external dependencies. To install them, run:

```bash
go mod download
```

Now run the tests to ensure everything is set up correctly:

```bash
go test ./...
```

## Usage

The program is designed to be run from the command line. It takes a regular expression and one or more file paths as arguments.

```bash
./your_program.sh -E <pattern> <file1> <file2> ...
```

- `-E`: The regular expression to search for.
- `<file1> <file2> ...`: The file(s) to search in.

## Running the program

1. Run `./your_program.sh` to run the program, which is implemented in `app/main.go`.

## Resources

- [Regular expressions](https://en.wikipedia.org/wiki/Regular_expression) (Regexes, for short) are patterns used to match character combinations in strings.
- [`grep`](https://en.wikipedia.org/wiki/Grep) is a CLI tool for searching using Regexes.
