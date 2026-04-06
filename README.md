# watcher

A Go program that:

1. Fetches a list of URLs from a configurable remote source (default: the
   [`urls.txt`](urls.txt) file in this repository).
2. Randomly selects 3 of those URLs.
3. Visits each selected URL with a headless Chrome browser via
   [chromedp](https://github.com/chromedp/chromedp) and prints the page title.

## Requirements

- Go 1.26+
- Google Chrome / Chromium installed and on `$PATH`

## Usage

```sh
# Build
go build -o watcher .

# Run with the default source URL (this repo's urls.txt on main)
./watcher

# Run with a custom source URL
./watcher -source https://example.com/my-urls.txt
```

The source file must be plain text with **one URL per line**.
Lines starting with `#` and blank lines are ignored.

## Running tests

```sh
go test ./...
```