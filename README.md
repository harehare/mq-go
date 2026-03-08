<h1 align="center">mq-go</h1>

Go bindings for [mq](https://github.com/harehare/mq) — a jq-like query language for Markdown processing.

This package wraps the `mq-ffi` C library using cgo, providing a safe, idiomatic Go API for querying and transforming Markdown, MDX, HTML, and plain text content.

## Requirements

- Go 1.21+
- `mq-ffi` C library installed on your system

## Installation

```bash
go get github.com/harehare/mq-go
```

## Usage

### Basic query

```go
engine, err := mq.New()
if err != nil {
    log.Fatal(err)
}
defer engine.Close()

result, err := engine.Run(".h1", "# Hello World\n\nSome text")
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Text()) // "# Hello World"
```

### Query with specific input format

```go
// Plain text (split by lines)
result, err := engine.RunWithFormat(`select(contains("foo"))`, "foo\nbar\nbaz", mq.FormatText)

// MDX content
result, err := engine.RunWithFormat("select(is_mdx())", "# Title\n\n<Component />", mq.FormatMDX)

// HTML content (auto-converted to Markdown)
result, err := engine.RunWithFormat(`select(contains("Hello"))`, "<h1>Hello</h1>", mq.FormatHTML)
```

### HTML to Markdown conversion

```go
markdown, err := mq.HTMLToMarkdown("<h1>Hello</h1><p>World</p>")

// With options
markdown, err := mq.HTMLToMarkdownWithOptions(html, mq.ConversionOptions{
    UseTitleAsH1:               true,
    GenerateFrontMatter:        true,
    ExtractScriptsAsCodeBlocks: true,
})
```

### Working with results

```go
result, err := engine.Run(".h2", content)

result.Text()         // all values joined by newline
result.Values()       // []string of non-empty values
result.Len()          // total number of values
result.Get(0)         // value at index
```

## Input Formats

| Constant | Description |
|---|---|
| `FormatMarkdown` | Standard Markdown (CommonMark/GFM) — default |
| `FormatMDX` | Markdown with JSX support |
| `FormatText` | Plain text, split by lines |
| `FormatHTML` | HTML, auto-converted to Markdown |

## Query Examples

```
.h1                          # extract H1 headings
.h2                          # extract H2 headings
.code                        # extract code blocks
.[]                          # iterate over all nodes
.h2 | select(contains("x"))  # filter by content
```

## License

[MIT](LICENSE)
