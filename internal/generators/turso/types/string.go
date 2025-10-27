package types

import (
	"fmt"
	"sqlsmith-go/internal/common"
	"strings"
)

// StringLiteral returns a quoted string literal driven by the LCG with more complex patterns.
func StringLiteral(lcg *common.LCG) string {
	choice := lcg.Intn(15)
	switch choice {
	case 0:
		// Simple alphanumeric
		return fmt.Sprintf("'%d-%d-%x'", lcg.Uint64(), lcg.Uint64(), lcg.Uint64())
	case 1:
		// Empty string
		return "''"
	case 2:
		// Single character
		chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		return fmt.Sprintf("'%c'", chars[lcg.Intn(len(chars))])
	case 3:
		// String with escaped single quote
		return fmt.Sprintf("'test''s value %d'", lcg.Intn(100))
	case 4:
		// String with newline
		return fmt.Sprintf("'line1\\nline2-%d'", lcg.Intn(100))
	case 5:
		// String with tab
		return fmt.Sprintf("'col1\\tcol2\\t%d'", lcg.Intn(100))
	case 6:
		// String with backslash
		return fmt.Sprintf("'path\\\\to\\\\file%d'", lcg.Intn(100))
	case 7:
		// Long string (limited to 200 chars to avoid issues)
		var b strings.Builder
		b.WriteString("'")
		length := 100 + lcg.Intn(101) // 100-200 chars
		for i := 0; i < length; i++ {
			b.WriteByte(byte('a' + (i % 26)))
		}
		b.WriteString("'")
		return b.String()
	case 8:
		// String with spaces
		return fmt.Sprintf("'  spaces  %d  '", lcg.Intn(100))
	case 9:
		// Numeric string
		return fmt.Sprintf("'%d'", lcg.Intn(1000000))
	case 10:
		// Email-like string
		return fmt.Sprintf("'user%d@example.com'", lcg.Intn(1000))
	case 11:
		// URL-like string
		return fmt.Sprintf("'https://example.com/path%d'", lcg.Intn(1000))
	case 12:
		// JSON-like string (escaped)
		return fmt.Sprintf("'{\"key\":%d,\"value\":\"data\"}'", lcg.Intn(100))
	case 13:
		// String with special characters
		specials := []string{"!@#$%", "&*()", "<>?", "[]{}"}
		return fmt.Sprintf("'test%s%d'", specials[lcg.Intn(len(specials))], lcg.Intn(100))
	default:
		// Unicode characters (simple emoji/symbols)
		unicodeStrings := []string{"'Hello 🌍 %d'", "'Test ✓ %d'", "'Star ★ %d'", "'Arrow → %d'"}
		return fmt.Sprintf(unicodeStrings[lcg.Intn(len(unicodeStrings))], lcg.Intn(100))
	}
}
