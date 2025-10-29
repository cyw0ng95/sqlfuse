package keywords

import (
	"testing"
)

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple name", "mycolumn", "\"mycolumn\""},
		{"with underscore", "my_column", "\"my_column\""},
		{"with number", "column1", "\"column1\""},
		{"reserved keyword SELECT", "SELECT", "\"SELECT\""},
		{"reserved keyword select", "select", "\"select\""},
		{"with double quote", "my\"column", "\"my\"\"column\""},
		{"multiple double quotes", "\"\"test\"\"", "\"\"\"\"\"test\"\"\"\"\""},
		{"empty string", "", "\"\""},
		{"space in name", "my column", "\"my column\""},
		{"special chars", "my-column", "\"my-column\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := QuoteIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("QuoteIdentifier(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestQuoteIdentifierIfNeeded(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple safe name", "mycolumn", "mycolumn"},
		{"with underscore", "my_column", "my_column"},
		{"with number", "column1", "column1"},
		{"with dollar", "my$column", "my$column"},
		{"reserved keyword SELECT", "SELECT", "\"SELECT\""},
		{"reserved keyword select", "select", "\"select\""},
		{"reserved keyword INSERT", "INSERT", "\"INSERT\""},
		{"starts with number", "1column", "\"1column\""},
		{"contains space", "my column", "\"my column\""},
		{"contains hyphen", "my-column", "\"my-column\""},
		{"contains dot", "my.column", "\"my.column\""},
		{"empty string", "", "\"\""},
		{"just underscore", "_", "_"},
		{"starts with dollar", "$column", "\"$column\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := QuoteIdentifierIfNeeded(tt.input)
			if result != tt.expected {
				t.Errorf("QuoteIdentifierIfNeeded(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNeedsQuoting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"simple name", "mycolumn", false},
		{"with underscore", "my_column", false},
		{"with number", "column1", false},
		{"with dollar", "my$column", false},
		{"starts with letter", "abc", false},
		{"starts with underscore", "_column", false},
		{"starts with number", "1column", true},
		{"starts with dollar", "$column", true},
		{"contains space", "my column", true},
		{"contains hyphen", "my-column", true},
		{"contains dot", "my.column", true},
		{"empty string", "", true},
		{"uppercase letters", "MYCOLUMN", false},
		{"mixed case", "MyColumn", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := needsQuoting(tt.input)
			if result != tt.expected {
				t.Errorf("needsQuoting(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsLetter(t *testing.T) {
	tests := []struct {
		name     string
		input    byte
		expected bool
	}{
		{"lowercase a", 'a', true},
		{"lowercase z", 'z', true},
		{"uppercase A", 'A', true},
		{"uppercase Z", 'Z', true},
		{"digit 0", '0', false},
		{"digit 9", '9', false},
		{"underscore", '_', false},
		{"space", ' ', false},
		{"hyphen", '-', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isLetter(tt.input)
			if result != tt.expected {
				t.Errorf("isLetter(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsDigit(t *testing.T) {
	tests := []struct {
		name     string
		input    byte
		expected bool
	}{
		{"digit 0", '0', true},
		{"digit 5", '5', true},
		{"digit 9", '9', true},
		{"letter a", 'a', false},
		{"letter Z", 'Z', false},
		{"underscore", '_', false},
		{"space", ' ', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDigit(tt.input)
			if result != tt.expected {
				t.Errorf("isDigit(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestQuoteIdentifierWithKeywords(t *testing.T) {
	// Test that all reserved keywords are properly quoted
	keywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "DROP",
		"TABLE", "INDEX", "VIEW", "WHERE", "FROM", "JOIN",
		"GROUP", "ORDER", "BY", "HAVING", "LIMIT", "OFFSET",
	}

	for _, keyword := range keywords {
		t.Run(keyword, func(t *testing.T) {
			result := QuoteIdentifierIfNeeded(keyword)
			if result[0] != '"' || result[len(result)-1] != '"' {
				t.Errorf("QuoteIdentifierIfNeeded(%q) = %q, expected quoted result", keyword, result)
			}
		})
	}
}
