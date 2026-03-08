package reviewer

import (
	"strings"
	"testing"
)

func TestSplitComment_NoSplitNeeded(t *testing.T) {
	body := "Short comment"
	maxSize := 100
	parts := splitComment(body, maxSize)

	if len(parts) != 1 {
		t.Errorf("Expected 1 part, got %d", len(parts))
	}
	if parts[0] != body {
		t.Errorf("Expected %q, got %q", body, parts[0])
	}
}

func TestSplitComment_SplitAtDoubleNewline(t *testing.T) {
	body := strings.Repeat("a", 50) + "\n\n" + strings.Repeat("b", 50)
	maxSize := 60
	parts := splitComment(body, maxSize)

	if len(parts) != 2 {
		t.Errorf("Expected 2 parts, got %d", len(parts))
	}
	if !strings.Contains(parts[0], "aaa") {
		t.Errorf("First part should contain 'a's: %q", parts[0])
	}
	if !strings.Contains(parts[1], "bbb") {
		t.Errorf("Second part should contain 'b's: %q", parts[1])
	}
}

func TestSplitComment_SplitAtMarkdownHeader(t *testing.T) {
	body := strings.Repeat("a", 50) + "\n## Header\n" + strings.Repeat("b", 50)
	maxSize := 60
	parts := splitComment(body, maxSize)

	if len(parts) < 2 {
		t.Errorf("Expected at least 2 parts, got %d", len(parts))
	}
	// First part should not contain the header
	if strings.Contains(parts[0], "## Header") {
		t.Errorf("First part should not contain header: %q", parts[0])
	}
}

func TestSplitComment_SplitAtSingleNewline(t *testing.T) {
	body := strings.Repeat("a", 50) + "\n" + strings.Repeat("b", 50)
	maxSize := 60
	parts := splitComment(body, maxSize)

	if len(parts) != 2 {
		t.Errorf("Expected 2 parts, got %d", len(parts))
	}
}

func TestSplitComment_MultipleSplits(t *testing.T) {
	// Create a body that needs multiple splits
	paragraph1 := strings.Repeat("a", 40)
	paragraph2 := strings.Repeat("b", 40)
	paragraph3 := strings.Repeat("c", 40)
	body := paragraph1 + "\n\n" + paragraph2 + "\n\n" + paragraph3
	maxSize := 50

	parts := splitComment(body, maxSize)

	if len(parts) != 3 {
		t.Errorf("Expected 3 parts, got %d", len(parts))
	}
}

func TestSplitComment_AllPartsWithinLimit(t *testing.T) {
	// Test that all parts are within the size limit
	body := strings.Repeat("x", 300)
	maxSize := 100

	parts := splitComment(body, maxSize)

	for i, part := range parts {
		if len(part) > maxSize {
			t.Errorf("Part %d exceeds maxSize: %d > %d", i, len(part), maxSize)
		}
	}
}

func TestFindSplitPoint_TextShorterThanMax(t *testing.T) {
	text := "Short text"
	maxSize := 100
	point := findSplitPoint(text, maxSize)

	if point != len(text) {
		t.Errorf("Expected %d, got %d", len(text), point)
	}
}

func TestFindSplitPoint_AtDoubleNewline(t *testing.T) {
	text := strings.Repeat("a", 30) + "\n\n" + strings.Repeat("b", 30)
	maxSize := 50
	point := findSplitPoint(text, maxSize)

	// Should split at the double newline (position 30)
	if point != 30 {
		t.Errorf("Expected split at 30, got %d", point)
	}
	if point > maxSize {
		t.Errorf("Split point %d exceeds maxSize %d", point, maxSize)
	}
}

func TestFindSplitPoint_AtMarkdownHeader(t *testing.T) {
	text := strings.Repeat("a", 30) + "\n## Header\n" + strings.Repeat("b", 30)
	maxSize := 50
	point := findSplitPoint(text, maxSize)

	// Should split before the header
	if point != 30 {
		t.Errorf("Expected split at 30, got %d", point)
	}
	if point > maxSize {
		t.Errorf("Split point %d exceeds maxSize %d", point, maxSize)
	}
}

func TestFindSplitPoint_AtSingleNewline(t *testing.T) {
	text := strings.Repeat("a", 30) + "\n" + strings.Repeat("b", 30)
	maxSize := 50
	point := findSplitPoint(text, maxSize)

	// Should split at the newline (position 30)
	if point != 30 {
		t.Errorf("Expected split at 30, got %d", point)
	}
	if point > maxSize {
		t.Errorf("Split point %d exceeds maxSize %d", point, maxSize)
	}
}

func TestFindSplitPoint_AtMaxSize(t *testing.T) {
	// No good split points - should split at maxSize
	text := strings.Repeat("x", 100)
	maxSize := 50
	point := findSplitPoint(text, maxSize)

	if point != maxSize {
		t.Errorf("Expected split at maxSize %d, got %d", maxSize, point)
	}
}

func TestFindSplitPoint_NeverExceedsMaxSize(t *testing.T) {
	testCases := []struct {
		text    string
		maxSize int
	}{
		{strings.Repeat("a", 100), 50},
		{strings.Repeat("a", 30) + "\n\n" + strings.Repeat("b", 30), 40},
		{strings.Repeat("a", 20) + "\n## Header\n" + strings.Repeat("b", 20), 30},
		{"a\nb\nc\nd\ne\nf\ng\nh\ni\nj\nk", 5},
	}

	for i, tc := range testCases {
		point := findSplitPoint(tc.text, tc.maxSize)
		if point > tc.maxSize {
			t.Errorf("Test case %d: split point %d exceeds maxSize %d", i, point, tc.maxSize)
		}
	}
}

func TestSplitComment_PreservesContent(t *testing.T) {
	// Test that splitting and joining preserves the original content
	body := strings.Repeat("a", 30) + "\n\n" + strings.Repeat("b", 30) + "\n\n" + strings.Repeat("c", 30)
	maxSize := 40

	parts := splitComment(body, maxSize)

	// Join parts back (with newlines that were trimmed)
	joined := strings.Join(parts, "\n")

	// Content should be mostly preserved (allowing for trimmed newlines)
	originalWords := strings.Fields(body)
	joinedWords := strings.Fields(joined)

	if len(originalWords) != len(joinedWords) {
		t.Errorf("Content not preserved: original had %d words, joined has %d words",
			len(originalWords), len(joinedWords))
	}
}

func TestTruncateComment_NoTruncateNeeded(t *testing.T) {
	body := "Short comment"
	maxSize := 100
	result := truncateComment(body, maxSize)

	if result != body {
		t.Errorf("Expected %q, got %q", body, result)
	}
}

func TestTruncateComment_WithTruncation(t *testing.T) {
	body := strings.Repeat("a", 100)
	maxSize := 50
	result := truncateComment(body, maxSize)

	if len(result) > maxSize {
		t.Errorf("Truncated comment exceeds maxSize: %d > %d", len(result), maxSize)
	}

	if !strings.HasSuffix(result, "... (truncated)") {
		t.Errorf("Expected truncation suffix, got: %q", result)
	}
}

func TestTruncateComment_AtParagraphBoundary(t *testing.T) {
	paragraph1 := strings.Repeat("a", 40)
	paragraph2 := strings.Repeat("b", 40)
	body := paragraph1 + "\n\n" + paragraph2
	maxSize := 60

	result := truncateComment(body, maxSize)

	if len(result) > maxSize {
		t.Errorf("Truncated comment exceeds maxSize: %d > %d", len(result), maxSize)
	}

	// Should truncate at paragraph boundary
	if strings.Contains(result, "bbb") {
		t.Errorf("Should have truncated before second paragraph: %q", result)
	}

	if !strings.Contains(result, "aaa") {
		t.Errorf("Should contain first paragraph: %q", result)
	}
}

func TestTruncateComment_VerySmallMaxSize(t *testing.T) {
	body := strings.Repeat("a", 100)
	maxSize := 5
	result := truncateComment(body, maxSize)

	if len(result) > maxSize {
		t.Errorf("Truncated comment exceeds maxSize: %d > %d", len(result), maxSize)
	}
}

func TestTruncateComment_PreservesBeginning(t *testing.T) {
	body := "Important start" + strings.Repeat("x", 100)
	maxSize := 50
	result := truncateComment(body, maxSize)

	if !strings.Contains(result, "Important start") {
		t.Errorf("Should preserve the beginning: %q", result)
	}

	if strings.Contains(result, strings.Repeat("x", 50)) {
		t.Errorf("Should have truncated the repeated content: %q", result)
	}
}
