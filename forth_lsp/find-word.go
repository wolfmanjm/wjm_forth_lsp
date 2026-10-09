package forth_lsp

import (
	"unicode"
)

// WordAtPoint extracts the word at the given byte index in a string line
func WordAtPoint(line string, point int) string {
	if point < 0 || point >= len(line) {
		return ""
	}

	// If the point itself is on a space/delimiter, look at your specific 
	// requirement (e.g., return empty or check previous character)
	if unicode.IsSpace(rune(line[point])) {
		return ""
	}

	// Find the start boundary of the word
	start := point
	for start > 0 {
		// Read previous character byte (assumes ASCII/UTF-8 single byte logic for simplicity)
		// For true Unicode support, iterate using a rune scanner or strings.LastIndexFunc
		if unicode.IsSpace(rune(line[start-1])) /* || strings.ContainsRune(".,;:!?()[]{}", rune(line[start-1]))*/ {
			break
		}
		start--
	}

	// Find the end boundary of the word
	end := point
	for end < len(line) {
		if unicode.IsSpace(rune(line[end])) /*|| strings.ContainsRune(".,;:!?()[]{}", rune(line[end]))*/ {
			break
		}
		end++
	}

	return line[start:end]
}

