package keywords

import (
	"fmt"
	"strings"
)

// QuoteIdentifier quotes a SQL identifier if necessary to avoid keyword conflicts.
// It uses double-quote syntax as per SQLite standards.
//
// Quoting rules:
// 1. Always quote if the identifier is a reserved keyword
// 2. Always quote if the identifier contains special characters
// 3. Always quote to be safe (current implementation - conservative approach)
//
// The identifier is escaped by replacing " with "".
func QuoteIdentifier(name string) string {
	return fmt.Sprintf("\"%s\"", strings.ReplaceAll(name, "\"", "\"\""))
}

// QuoteIdentifierIfNeeded quotes a SQL identifier only if it's a reserved keyword
// or contains special characters. Otherwise returns the identifier unquoted.
//
// This is a more permissive approach that only quotes when necessary.
// However, for SQL generation, QuoteIdentifier is generally safer.
func QuoteIdentifierIfNeeded(name string) string {
	// Check if it's a reserved keyword (case-insensitive)
	if IsReserved(name) {
		return QuoteIdentifier(name)
	}
	
	// Check if it contains special characters that require quoting
	if needsQuoting(name) {
		return QuoteIdentifier(name)
	}
	
	// Return unquoted if safe
	return name
}

// needsQuoting checks if an identifier contains characters that require quoting.
// SQLite identifiers can contain: letters, digits, underscore (_), and dollar sign ($)
// They must start with a letter or underscore.
func needsQuoting(name string) bool {
	if len(name) == 0 {
		return true // Empty names must be quoted
	}
	
	// First character must be letter or underscore
	first := name[0]
	if !isLetter(first) && first != '_' {
		return true
	}
	
	// Remaining characters can be letters, digits, underscore, or dollar sign
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !isLetter(c) && !isDigit(c) && c != '_' && c != '$' {
			return true
		}
	}
	
	return false
}

// isLetter checks if a byte is an ASCII letter (a-z or A-Z).
func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isDigit checks if a byte is an ASCII digit (0-9).
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}
