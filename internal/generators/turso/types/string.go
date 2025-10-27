package types

import (
	"fmt"
	"sqlsmith-go/internal/common"
	"strings"
)

// StringLiteral returns a quoted string literal driven by the LCG with more complex patterns.
// The hint parameter is used to guide value generation for columns that likely have UNIQUE constraints.
func StringLiteral(lcg *common.LCG, hint string) string {
	// Check if this column likely has a UNIQUE constraint based on its name
	hintLower := strings.ToLower(hint)
	isLikelyUnique := strings.Contains(hintLower, "email") ||
		strings.Contains(hintLower, "key") ||
		strings.Contains(hintLower, "username") ||
		strings.Contains(hintLower, "user_name")

	choice := lcg.Intn(15)
	
	// If this is likely a UNIQUE column, avoid cases that can easily produce duplicates:
	// - case 1: empty string
	// - case 2: single character (only 62 possible values)
	if isLikelyUnique && (choice == 1 || choice == 2) {
		// Re-roll, avoiding cases 1 and 2
		choice = lcg.Intn(13) // 0-12
		if choice >= 1 {
			choice += 2 // Skip cases 1 and 2, map to 3-14
		}
	}
	
	switch choice {
	case 0:
		// Simple alphanumeric - includes random components for uniqueness
		return fmt.Sprintf("'%d-%d-%x'", lcg.Uint64(), lcg.Uint64(), lcg.Uint64())
	case 1:
		// Empty string - avoided for UNIQUE columns
		return "''"
	case 2:
		// Single character - avoided for UNIQUE columns
		chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		return fmt.Sprintf("'%c'", chars[lcg.Intn(len(chars))])
	case 3:
		// String with escaped single quote - use much larger range for UNIQUE columns
		if isLikelyUnique {
			return fmt.Sprintf("'test''s value %d'", lcg.Uint64()%1000000)
		}
		return fmt.Sprintf("'test''s value %d'", lcg.Intn(100))
	case 4:
		// String with newline - use much larger range for UNIQUE columns
		if isLikelyUnique {
			return fmt.Sprintf("'line1\\nline2-%d'", lcg.Uint64()%1000000)
		}
		return fmt.Sprintf("'line1\\nline2-%d'", lcg.Intn(100))
	case 5:
		// String with tab - use much larger range for UNIQUE columns
		if isLikelyUnique {
			return fmt.Sprintf("'col1\\tcol2\\t%d'", lcg.Uint64()%1000000)
		}
		return fmt.Sprintf("'col1\\tcol2\\t%d'", lcg.Intn(100))
	case 6:
		// String with backslash - use much larger range for UNIQUE columns
		if isLikelyUnique {
			return fmt.Sprintf("'path\\\\to\\\\file%d'", lcg.Uint64()%1000000)
		}
		return fmt.Sprintf("'path\\\\to\\\\file%d'", lcg.Intn(100))
	case 7:
		// Long string (limited to 200 chars to avoid issues)
		// For UNIQUE columns, add a random suffix to ensure uniqueness
		var b strings.Builder
		b.WriteString("'")
		length := 100 + lcg.Intn(101) // 100-200 chars
		for i := 0; i < length; i++ {
			b.WriteByte(byte('a' + (i % 26)))
		}
		if isLikelyUnique {
			b.WriteString(fmt.Sprintf("-%d", lcg.Uint64()))
		}
		b.WriteString("'")
		return b.String()
	case 8:
		// String with spaces - use much larger range for UNIQUE columns
		if isLikelyUnique {
			return fmt.Sprintf("'  spaces  %d  '", lcg.Uint64()%1000000)
		}
		return fmt.Sprintf("'  spaces  %d  '", lcg.Intn(100))
	case 9:
		// Numeric string
		return fmt.Sprintf("'%d'", lcg.Intn(1000000))
	case 10:
		// Email-like string - use larger range for UNIQUE columns
		if isLikelyUnique {
			return fmt.Sprintf("'user%d@example.com'", lcg.Uint64()%1000000)
		}
		return fmt.Sprintf("'user%d@example.com'", lcg.Intn(1000))
	case 11:
		// URL-like string - use much larger range for UNIQUE columns
		if isLikelyUnique {
			return fmt.Sprintf("'https://example.com/path%d'", lcg.Uint64()%1000000)
		}
		return fmt.Sprintf("'https://example.com/path%d'", lcg.Intn(1000))
	case 12:
		// JSON-like string - use larger range for UNIQUE columns
		if isLikelyUnique {
			return fmt.Sprintf("'{\"key\":%d,\"value\":\"data\"}'", lcg.Uint64()%1000000)
		}
		return fmt.Sprintf("'{\"key\":%d,\"value\":\"data\"}'", lcg.Intn(100))
	case 13:
		// String with special characters - use much larger range for UNIQUE columns
		specials := []string{"!@#$%", "&*()", "<>?", "[]{}"}
		if isLikelyUnique {
			return fmt.Sprintf("'test%s%d'", specials[lcg.Intn(len(specials))], lcg.Uint64()%1000000)
		}
		return fmt.Sprintf("'test%s%d'", specials[lcg.Intn(len(specials))], lcg.Intn(100))
	default:
		// Unicode characters (simple emoji/symbols) - use much larger range for UNIQUE columns
		unicodeStrings := []string{"'Hello 🌍 %d'", "'Test ✓ %d'", "'Star ★ %d'", "'Arrow → %d'"}
		if isLikelyUnique {
			return fmt.Sprintf(unicodeStrings[lcg.Intn(len(unicodeStrings))], lcg.Uint64()%1000000)
		}
		return fmt.Sprintf(unicodeStrings[lcg.Intn(len(unicodeStrings))], lcg.Intn(100))
	}
}
