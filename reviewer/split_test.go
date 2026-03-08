package reviewer

import (
	"strings"
	"testing"
	"unicode/utf8"
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
		runeCount := utf8.RuneCountInString(part)
		if runeCount > maxSize {
			t.Errorf("Part %d exceeds maxSize: %d > %d", i, runeCount, maxSize)
		}
	}
}

func TestFindSplitPoint_TextShorterThanMax(t *testing.T) {
	text := "Short text"
	maxSize := 100
	point := findSplitPoint(text, maxSize)

	expectedRuneCount := utf8.RuneCountInString(text)
	if point != expectedRuneCount {
		t.Errorf("Expected %d, got %d", expectedRuneCount, point)
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

	runeCount := utf8.RuneCountInString(result)
	if runeCount > maxSize {
		t.Errorf("Truncated comment exceeds maxSize: %d > %d", runeCount, maxSize)
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

	runeCount := utf8.RuneCountInString(result)
	if runeCount > maxSize {
		t.Errorf("Truncated comment exceeds maxSize: %d > %d", runeCount, maxSize)
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

	runeCount := utf8.RuneCountInString(result)
	if runeCount > maxSize {
		t.Errorf("Truncated comment exceeds maxSize: %d > %d", runeCount, maxSize)
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

// Test UTF-8 multibyte character handling
func TestSplitComment_UTF8Multibyte(t *testing.T) {
	// Japanese characters (3 bytes each in UTF-8)
	body := strings.Repeat("あ", 50) + "\n\n" + strings.Repeat("い", 50)
	maxSize := 60 // 60 characters (runes), not bytes

	parts := splitComment(body, maxSize)

	// Verify all parts are valid UTF-8
	for i, part := range parts {
		if !utf8.ValidString(part) {
			t.Errorf("Part %d contains invalid UTF-8", i)
		}
		runeCount := utf8.RuneCountInString(part)
		if runeCount > maxSize {
			t.Errorf("Part %d exceeds maxSize in runes: %d > %d", i, runeCount, maxSize)
		}
	}

	if len(parts) != 2 {
		t.Errorf("Expected 2 parts, got %d", len(parts))
	}
}

func TestTruncateComment_UTF8Multibyte(t *testing.T) {
	// Mix of ASCII and multibyte characters
	body := "Test: " + strings.Repeat("日本語", 50) // "日本語" = 3 characters
	maxSize := 50

	result := truncateComment(body, maxSize)

	// Verify result is valid UTF-8
	if !utf8.ValidString(result) {
		t.Errorf("Result contains invalid UTF-8: %q", result)
	}

	runeCount := utf8.RuneCountInString(result)
	if runeCount > maxSize {
		t.Errorf("Truncated comment exceeds maxSize: %d > %d", runeCount, maxSize)
	}

	if !strings.Contains(result, "Test:") {
		t.Errorf("Should preserve the beginning: %q", result)
	}
}

func TestFindSplitPoint_UTF8SafeBoundary(t *testing.T) {
	// Emoji (4 bytes in UTF-8) + Japanese (3 bytes each)
	text := strings.Repeat("🎉", 30) + "\n\n" + strings.Repeat("テスト", 30)
	maxSize := 40

	point := findSplitPoint(text, maxSize)

	// Verify that slicing at the split point produces valid UTF-8
	byteIndex := 0
	for i := 0; i < point; i++ {
		_, size := utf8.DecodeRuneInString(text[byteIndex:])
		byteIndex += size
	}

	firstPart := text[:byteIndex]
	if !utf8.ValidString(firstPart) {
		t.Errorf("Split produces invalid UTF-8: %q", firstPart)
	}

	if point > maxSize {
		t.Errorf("Split point %d exceeds maxSize %d", point, maxSize)
	}
}

// Test invalid maxSize (zero or negative)
func TestSplitComment_InvalidMaxSize(t *testing.T) {
	body := "Test comment body"

	// Test with zero maxSize
	parts := splitComment(body, 0)
	if len(parts) != 1 {
		t.Errorf("Expected 1 part with maxSize=0, got %d", len(parts))
	}
	if parts[0] != body {
		t.Errorf("Expected body to be returned as-is with maxSize=0, got %q", parts[0])
	}

	// Test with negative maxSize
	parts = splitComment(body, -10)
	if len(parts) != 1 {
		t.Errorf("Expected 1 part with negative maxSize, got %d", len(parts))
	}
	if parts[0] != body {
		t.Errorf("Expected body to be returned as-is with negative maxSize, got %q", parts[0])
	}
}

func TestFindSplitPoint_InvalidMaxSize(t *testing.T) {
	text := "Test text"

	// Test with zero maxSize
	point := findSplitPoint(text, 0)
	if point != 0 {
		t.Errorf("Expected 0 with maxSize=0, got %d", point)
	}

	// Test with negative maxSize
	point = findSplitPoint(text, -5)
	if point != 0 {
		t.Errorf("Expected 0 with negative maxSize, got %d", point)
	}
}

func TestTruncateComment_InvalidMaxSize(t *testing.T) {
	body := "Test comment body"

	// Test with zero maxSize
	result := truncateComment(body, 0)
	if result != "" {
		t.Errorf("Expected empty string with maxSize=0, got %q", result)
	}

	// Test with negative maxSize
	result = truncateComment(body, -10)
	if result != "" {
		t.Errorf("Expected empty string with negative maxSize, got %q", result)
	}
}
