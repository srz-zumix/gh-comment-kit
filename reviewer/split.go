package reviewer

import (
	"regexp"
	"strings"
)

// splitComment splits a comment body into multiple parts if it exceeds maxSize.
// Priority:
// 1. Split at double newlines (paragraphs)
// 2. Split at markdown headers
// 3. Split at single newlines
func splitComment(body string, maxSize int) []string {
	if len(body) <= maxSize {
		return []string{body}
	}

	var parts []string
	remaining := body

	for len(remaining) > 0 {
		if len(remaining) <= maxSize {
			parts = append(parts, remaining)
			break
		}

		splitPoint := findSplitPoint(remaining, maxSize)
		parts = append(parts, remaining[:splitPoint])
		remaining = strings.TrimLeft(remaining[splitPoint:], "\n")
	}

	return parts
}

// findSplitPoint finds the best position to split the text within maxSize
// Returns a value that is always <= maxSize
func findSplitPoint(text string, maxSize int) int {
	textLen := len(text)
	if textLen <= maxSize {
		return textLen
	}

	// Search within maxSize range
	searchText := text[:maxSize]

	// Try to split at double newline (paragraph boundary)
	if idx := strings.LastIndex(searchText, "\n\n"); idx > 0 {
		return idx
	}

	// Try to split at markdown header
	headerRegex := regexp.MustCompile(`\n#{1,6} `)
	if matches := headerRegex.FindAllStringIndex(searchText, -1); len(matches) > 0 {
		return matches[len(matches)-1][0]
	}

	// Try to split at single newline
	if idx := strings.LastIndex(searchText, "\n"); idx > 0 {
		return idx
	}

	// Last resort: split at maxSize (guaranteed to be within limit)
	return maxSize
}

// truncateComment truncates a comment body to fit within maxSize.
// It tries to truncate at a good boundary and adds "..." to indicate truncation.
func truncateComment(body string, maxSize int) string {
	const suffix = "\n\n... (truncated)"
	suffixLen := len(suffix)

	if len(body) <= maxSize {
		return body
	}

	// Need to reserve space for the suffix
	if maxSize < suffixLen {
		// If maxSize is too small, just truncate
		return body[:maxSize]
	}

	targetSize := maxSize - suffixLen
	truncatePoint := findSplitPoint(body, targetSize)

	return body[:truncatePoint] + suffix
}
