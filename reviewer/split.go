package reviewer

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// markdownHeaderRegex matches a newline followed by a markdown header (e.g., "\n## ").
var markdownHeaderRegex = regexp.MustCompile(`\n#{1,6} `)

// splitComment splits a comment body into multiple parts if it exceeds maxSize.
// maxSize is interpreted as the number of characters (runes), not bytes, to match GitHub's limit.
// Priority:
// 1. Split at double newlines (paragraphs)
// 2. Split at markdown headers
// 3. Split at single newlines
func splitComment(body string, maxSize int) []string {
	// Prevent infinite loops and invalid operations when maxSize is invalid
	if maxSize <= 0 {
		// Return the body as-is, caller should handle this case
		return []string{body}
	}

	if utf8.RuneCountInString(body) <= maxSize {
		return []string{body}
	}

	var parts []string
	remaining := body

	for utf8.RuneCountInString(remaining) > 0 {
		runeCount := utf8.RuneCountInString(remaining)
		if runeCount <= maxSize {
			parts = append(parts, remaining)
			break
		}

		splitPoint := findSplitPoint(remaining, maxSize)
		// Convert rune index to byte index for slicing
		byteIndex := runeIndexToByteIndex(remaining, splitPoint)
		parts = append(parts, remaining[:byteIndex])
		remaining = strings.TrimLeft(remaining[byteIndex:], "\n")
	}

	return parts
}

// runeIndexToByteIndex converts a rune index to a byte index in the string.
func runeIndexToByteIndex(s string, runeIndex int) int {
	if runeIndex <= 0 {
		return 0
	}
	byteIndex := 0
	for i := 0; i < runeIndex && byteIndex < len(s); i++ {
		_, size := utf8.DecodeRuneInString(s[byteIndex:])
		byteIndex += size
	}
	return byteIndex
}

// findSplitPoint finds the best position (in rune count) to split the text within maxSize.
// Returns a value that is always <= maxSize (in runes).
func findSplitPoint(text string, maxSize int) int {
	// Guard against invalid maxSize to prevent panics and infinite loops
	if maxSize <= 0 {
		return 0
	}

	runeCount := utf8.RuneCountInString(text)
	if runeCount <= maxSize {
		return runeCount
	}

	// Get the substring up to maxSize runes
	searchByteLen := runeIndexToByteIndex(text, maxSize)
	searchText := text[:searchByteLen]

	// Try to split at double newline (paragraph boundary)
	if idx := strings.LastIndex(searchText, "\n\n"); idx > 0 {
		return utf8.RuneCountInString(text[:idx])
	}

	// Try to split at markdown header
	if matches := markdownHeaderRegex.FindAllStringIndex(searchText, -1); len(matches) > 0 {
		lastMatchByteIndex := matches[len(matches)-1][0]
		return utf8.RuneCountInString(text[:lastMatchByteIndex])
	}

	// Try to split at single newline
	if idx := strings.LastIndex(searchText, "\n"); idx > 0 {
		return utf8.RuneCountInString(text[:idx])
	}

	// Last resort: split at maxSize (guaranteed to be within limit)
	return maxSize
}

// truncateComment truncates a comment body to fit within maxSize.
// maxSize is interpreted as the number of characters (runes), not bytes.
// It tries to truncate at a good boundary and adds "..." to indicate truncation.
func truncateComment(body string, maxSize int) string {
	// Guard against invalid maxSize
	if maxSize <= 0 {
		return ""
	}

	const suffix = "\n\n... (truncated)"
	suffixRuneCount := utf8.RuneCountInString(suffix)

	bodyRuneCount := utf8.RuneCountInString(body)
	if bodyRuneCount <= maxSize {
		return body
	}

	// Need to reserve space for the suffix
	if maxSize < suffixRuneCount {
		// If maxSize is too small, just truncate
		byteIndex := runeIndexToByteIndex(body, maxSize)
		return body[:byteIndex]
	}

	targetSize := maxSize - suffixRuneCount
	truncatePoint := findSplitPoint(body, targetSize)
	byteIndex := runeIndexToByteIndex(body, truncatePoint)

	return body[:byteIndex] + suffix
}
