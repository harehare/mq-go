package mq

import (
	"strings"
	"testing"
)

func newEngine(t *testing.T) *Engine {
	t.Helper()
	engine, err := New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	t.Cleanup(func() { engine.Close() })
	return engine
}

func TestRun_ExtractsH1Headings(t *testing.T) {
	engine := newEngine(t)
	result, err := engine.Run(".h1", "# Hello World\n\n## Heading2\n\nText")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	values := result.Values()
	if len(values) != 1 || values[0] != "# Hello World" {
		t.Errorf("expected [# Hello World], got %v", values)
	}
}

func TestRun_ExtractsH2Headings(t *testing.T) {
	engine := newEngine(t)
	result, err := engine.Run(".h2", "# Hello World\n\n## Heading2\n\nText")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	values := result.Values()
	if len(values) != 1 || values[0] != "## Heading2" {
		t.Errorf("expected [## Heading2], got %v", values)
	}
}

func TestRun_ExtractsMultipleH2Headings(t *testing.T) {
	engine := newEngine(t)
	result, err := engine.Run(".h2", "# Main Title\n\n## Heading2A\n\nText\n\n## Heading2B\n\nMore text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	values := result.Values()
	if len(values) != 2 || values[0] != "## Heading2A" || values[1] != "## Heading2B" {
		t.Errorf("expected [## Heading2A, ## Heading2B], got %v", values)
	}
}

func TestRun_FiltersHeadingsWithSelect(t *testing.T) {
	engine := newEngine(t)
	result, err := engine.Run(`.h2 | select(contains("Feature"))`, "# Product\n\n## Features\n\nText\n\n## Installation\n\nMore text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	values := result.Values()
	if len(values) != 1 || values[0] != "## Features" {
		t.Errorf("expected [## Features], got %v", values)
	}
}

func TestRun_ExtractsListItems(t *testing.T) {
	engine := newEngine(t)
	result, err := engine.Run(".[]", "# List\n\n- Item 1\n- Item 2\n- Item 3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	values := result.Values()
	if len(values) != 3 {
		t.Errorf("expected 3 items, got %d: %v", len(values), values)
	}
}

func TestRun_ExtractsCodeBlocks(t *testing.T) {
	engine := newEngine(t)
	result, err := engine.Run(".code", "# Code\n\n```python\nprint('Hello')\n```")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	values := result.Values()
	if len(values) != 1 || !strings.Contains(values[0], "print('Hello')") {
		t.Errorf("expected code block, got %v", values)
	}
}

func TestRunWithFormat_Text(t *testing.T) {
	engine := newEngine(t)
	result, err := engine.RunWithFormat(`select(contains("2"))`, "Line 1\nLine 2\nLine 3", FormatText)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	values := result.Values()
	if len(values) != 1 || values[0] != "Line 2" {
		t.Errorf("expected [Line 2], got %v", values)
	}
}

func TestRunWithFormat_MDX(t *testing.T) {
	engine := newEngine(t)
	result, err := engine.RunWithFormat("select(is_mdx())", "# MDX Content\n\n<Component />", FormatMDX)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	values := result.Values()
	if len(values) != 1 || values[0] != "<Component />" {
		t.Errorf("expected [<Component />], got %v", values)
	}
}

func TestRunWithFormat_HTML(t *testing.T) {
	engine := newEngine(t)
	result, err := engine.RunWithFormat(`select(contains("Hello"))`, "<h1>Hello</h1><p>World</p>", FormatHTML)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	values := result.Values()
	if len(values) != 1 || values[0] != "# Hello" {
		t.Errorf("expected [# Hello], got %v", values)
	}
}

func TestRun_InvalidSyntax(t *testing.T) {
	engine := newEngine(t)
	_, err := engine.Run(".invalid_selector!!!", "# Heading")
	if err == nil {
		t.Fatal("expected error for invalid syntax")
	}
}

func TestRun_ClosedEngine(t *testing.T) {
	engine, err := New()
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	engine.Close()
	_, err = engine.Run(".h1", "# Heading")
	if err == nil {
		t.Fatal("expected error for closed engine")
	}
}

func TestHTMLToMarkdown(t *testing.T) {
	markdown, err := HTMLToMarkdown("<h1>Hello World</h1><p>This is a <strong>test</strong>.</p>")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(markdown, "# Hello World") {
		t.Errorf("expected markdown to contain '# Hello World', got: %s", markdown)
	}
	if !strings.Contains(markdown, "**test**") {
		t.Errorf("expected markdown to contain '**test**', got: %s", markdown)
	}
}

func TestHTMLToMarkdownWithOptions(t *testing.T) {
	opts := ConversionOptions{
		UseTitleAsH1: true,
	}
	markdown, err := HTMLToMarkdownWithOptions(
		"<html><head><title>Page Title</title></head><body><h1>Content</h1></body></html>",
		opts,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(markdown, "# Page Title") {
		t.Errorf("expected markdown to contain '# Page Title', got: %s", markdown)
	}
}

func TestResult_Text(t *testing.T) {
	engine := newEngine(t)
	result, err := engine.Run(".h2", "# Title\n\n## Section 1\n\n## Section 2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := result.Text()
	if text != "## Section 1\n## Section 2" {
		t.Errorf("expected '## Section 1\\n## Section 2', got: %s", text)
	}
}

func TestResult_Values(t *testing.T) {
	engine := newEngine(t)
	result, err := engine.Run(".h2", "# Title\n\n## Section 1\n\n## Section 2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	values := result.Values()
	if len(values) != 2 || values[0] != "## Section 1" || values[1] != "## Section 2" {
		t.Errorf("expected [## Section 1, ## Section 2], got %v", values)
	}
}

func TestResult_Get(t *testing.T) {
	engine := newEngine(t)
	result, err := engine.Run(".h2", "# Title\n\n## Section 1\n\n## Section 2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = result.Get(-1)
	if err == nil {
		t.Error("expected error for negative index")
	}
	_, err = result.Get(100)
	if err == nil {
		t.Error("expected error for out of range index")
	}
}
