// Package mq provides Go bindings for mq - a jq-like tool for Markdown processing.
//
// This package wraps the mq-ffi C library using cgo, providing a safe, idiomatic
// Go API for querying and transforming Markdown, MDX, HTML, and text content.
//
// Example usage:
//
//	engine, err := mq.New()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer engine.Close()
//
//	result, err := engine.Run(".h1", "# Hello World\n\nText")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result.Text()) // "# Hello World"
package mq

/*
#cgo LDFLAGS: -lmq_ffi
#include <stdlib.h>
#include <stdbool.h>

typedef struct mq_result_t {
    char **values;
    uintptr_t values_len;
    char *error_msg;
} mq_result_t;

typedef struct MqConversionOptions {
    bool extract_scripts_as_code_blocks;
    bool generate_front_matter;
    bool use_title_as_h1;
} MqConversionOptions;

void *mq_create(void);
void mq_destroy(void *engine_ptr);
mq_result_t mq_eval(void *engine_ptr, const char *code_c, const char *input_c, const char *input_format_c);
void mq_free_result(mq_result_t result);
char *mq_html_to_markdown(const char *html_input_c, MqConversionOptions options, char **error_msg);
void mq_free_string(char *s);
*/
import "C"

import (
	"errors"
	"strings"
	"unsafe"
)

// InputFormat specifies the format of the input content.
type InputFormat string

const (
	// FormatMarkdown is standard Markdown (CommonMark/GFM).
	FormatMarkdown InputFormat = "markdown"
	// FormatMDX is Markdown with JSX support.
	FormatMDX InputFormat = "mdx"
	// FormatText is plain text (split by lines).
	FormatText InputFormat = "text"
	// FormatHTML is HTML content (auto-converted to Markdown).
	FormatHTML InputFormat = "html"
)

// ConversionOptions controls the behavior of HTML to Markdown conversion.
type ConversionOptions struct {
	// ExtractScriptsAsCodeBlocks extracts script tags as code blocks.
	ExtractScriptsAsCodeBlocks bool
	// GenerateFrontMatter generates front matter from HTML head metadata.
	GenerateFrontMatter bool
	// UseTitleAsH1 uses the HTML title tag as an H1 heading.
	UseTitleAsH1 bool
}

// Result holds the output of an mq query evaluation.
type Result struct {
	values []string
}

// Text returns all non-empty result values joined by newlines.
func (r *Result) Text() string {
	var nonEmpty []string
	for _, v := range r.values {
		if v != "" {
			nonEmpty = append(nonEmpty, v)
		}
	}
	return strings.Join(nonEmpty, "\n")
}

// Values returns all non-empty result values as a string slice.
func (r *Result) Values() []string {
	var nonEmpty []string
	for _, v := range r.values {
		if v != "" {
			nonEmpty = append(nonEmpty, v)
		}
	}
	return nonEmpty
}

// Len returns the total number of result values.
func (r *Result) Len() int {
	return len(r.values)
}

// Get returns the value at the specified index.
func (r *Result) Get(index int) (string, error) {
	if index < 0 || index >= len(r.values) {
		return "", errors.New("index out of range")
	}
	return r.values[index], nil
}

// Engine is an mq query engine instance.
type Engine struct {
	ptr unsafe.Pointer
}

// New creates a new mq engine instance.
func New() (*Engine, error) {
	ptr := C.mq_create()
	if ptr == nil {
		return nil, errors.New("failed to create mq engine")
	}
	return &Engine{ptr: ptr}, nil
}

// Close releases the native resources associated with the engine.
func (e *Engine) Close() {
	if e.ptr != nil {
		C.mq_destroy(e.ptr)
		e.ptr = nil
	}
}

// Run executes an mq query on the provided content using Markdown format.
func (e *Engine) Run(code, content string) (*Result, error) {
	return e.RunWithFormat(code, content, FormatMarkdown)
}

// RunWithFormat executes an mq query on the provided content with the specified format.
func (e *Engine) RunWithFormat(code, content string, format InputFormat) (*Result, error) {
	if e.ptr == nil {
		return nil, errors.New("engine has been closed")
	}

	cCode := C.CString(code)
	defer C.free(unsafe.Pointer(cCode))
	cContent := C.CString(content)
	defer C.free(unsafe.Pointer(cContent))
	cFormat := C.CString(string(format))
	defer C.free(unsafe.Pointer(cFormat))

	result := C.mq_eval(e.ptr, cCode, cContent, cFormat)
	defer C.mq_free_result(result)

	if result.error_msg != nil {
		errMsg := C.GoString(result.error_msg)
		return nil, errors.New(errMsg)
	}

	length := int(result.values_len)
	values := make([]string, length)
	if length > 0 && result.values != nil {
		cValues := unsafe.Slice((**C.char)(unsafe.Pointer(result.values)), length)
		for i, v := range cValues {
			if v != nil {
				values[i] = C.GoString(v)
			}
		}
	}

	return &Result{values: values}, nil
}

// HTMLToMarkdown converts HTML content to Markdown.
func HTMLToMarkdown(htmlContent string) (string, error) {
	return HTMLToMarkdownWithOptions(htmlContent, ConversionOptions{})
}

// HTMLToMarkdownWithOptions converts HTML content to Markdown with the specified options.
func HTMLToMarkdownWithOptions(htmlContent string, opts ConversionOptions) (string, error) {
	cHTML := C.CString(htmlContent)
	defer C.free(unsafe.Pointer(cHTML))

	cOpts := C.MqConversionOptions{
		extract_scripts_as_code_blocks: C.bool(opts.ExtractScriptsAsCodeBlocks),
		generate_front_matter:          C.bool(opts.GenerateFrontMatter),
		use_title_as_h1:               C.bool(opts.UseTitleAsH1),
	}

	var cErrorMsg *C.char
	resultPtr := C.mq_html_to_markdown(cHTML, cOpts, &cErrorMsg)

	if resultPtr == nil {
		var errMsg string
		if cErrorMsg != nil {
			errMsg = C.GoString(cErrorMsg)
			C.mq_free_string(cErrorMsg)
		} else {
			errMsg = "unknown error"
		}
		return "", errors.New(errMsg)
	}
	defer C.mq_free_string(resultPtr)

	return C.GoString(resultPtr), nil
}
