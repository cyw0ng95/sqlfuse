package keywords

import (
	"testing"
)

// TestKeywordAsIdentifier tests that SQL keywords can be safely used as identifiers
// when properly quoted.
func TestKeywordAsIdentifier(t *testing.T) {
	tests := []struct {
		keyword  string
		expected string
	}{
		{"SELECT", "\"SELECT\""},
		{"FROM", "\"FROM\""},
		{"WHERE", "\"WHERE\""},
		{"table", "\"table\""},
		{"Table", "\"Table\""},
		{"INDEX", "\"INDEX\""},
		{"create", "\"create\""},
	}

	for _, tt := range tests {
		t.Run(tt.keyword, func(t *testing.T) {
			result := QuoteIdentifier(tt.keyword)
			if result != tt.expected {
				t.Errorf("QuoteIdentifier(%q) = %q, want %q", tt.keyword, result, tt.expected)
			}
		})
	}
}

// TestTursoUnsupportedKeywordDetection verifies that Turso-unsupported keywords
// are correctly identified.
func TestTursoUnsupportedKeywordDetection(t *testing.T) {
	tursoUnsupported := []string{
		"FILTER", "MATCH", "MATERIALIZED", "OVER", "PARTITION",
		"RAISE", "RECURSIVE", "REGEXP", "WINDOW",
	}

	for _, kw := range tursoUnsupported {
		t.Run(kw, func(t *testing.T) {
			if IsTursoSupported(kw) {
				t.Errorf("IsTursoSupported(%q) = true, expected false (Turso doesn't support this keyword)", kw)
			}

			// Verify case insensitivity - test lowercase version
			lowerKw := ""
			for _, c := range kw {
				if c >= 'A' && c <= 'Z' {
					lowerKw += string(c + 32) // Convert to lowercase
				} else {
					lowerKw += string(c)
				}
			}
			if IsTursoSupported(lowerKw) {
				t.Errorf("IsTursoSupported(%q) = true, expected false (case-insensitive check)", lowerKw)
			}
		})
	}
}

// TestAllKeywordsAreReserved verifies that all keywords in the list are marked as Reserved.
// This is important because we want to be conservative and quote all keywords.
func TestAllKeywordsAreReserved(t *testing.T) {
	keywords := AllKeywords()
	for _, kw := range keywords {
		if kw.Type != Reserved {
			t.Errorf("Keyword %s has type %v, expected Reserved", kw.Name, kw.Type)
		}
	}
}

// TestKeywordCount verifies we have the expected number of keywords
func TestKeywordCountIntegration(t *testing.T) {
	keywords := AllKeywords()
	// SQLite has 147 keywords as documented in https://sqlite.org/lang_keywords.html
	if len(keywords) != 147 {
		t.Errorf("Expected 147 keywords, got %d", len(keywords))
	}
}

// TestCommonKeywordQuoting tests quoting of common SQL keywords that might appear
// as table or column names in real-world scenarios.
func TestCommonKeywordQuoting(t *testing.T) {
	commonKeywords := []string{
		"group", "GROUP", "Group",
		"order", "ORDER", "Order",
		"table", "TABLE", "Table",
		"index", "INDEX", "Index",
		"key", "KEY", "Key",
		"values", "VALUES", "Values",
		"default", "DEFAULT", "Default",
	}

	for _, kw := range commonKeywords {
		t.Run(kw, func(t *testing.T) {
			// All of these should be quoted
			result := QuoteIdentifierIfNeeded(kw)
			if len(result) < 2 || result[0] != '"' || result[len(result)-1] != '"' {
				t.Errorf("QuoteIdentifierIfNeeded(%q) = %q, expected quoted result", kw, result)
			}
		})
	}
}

// TestNonKeywordIdentifiers verifies that non-keyword identifiers can be used unquoted
// when using QuoteIdentifierIfNeeded.
func TestNonKeywordIdentifiers(t *testing.T) {
	nonKeywords := []string{
		"mycolumn", "my_column", "column1", "col_123",
		"firstName", "last_name", "user_id", "created_at",
	}

	for _, id := range nonKeywords {
		t.Run(id, func(t *testing.T) {
			result := QuoteIdentifierIfNeeded(id)
			// These should not be quoted (no quotes at start/end)
			if result != id {
				t.Errorf("QuoteIdentifierIfNeeded(%q) = %q, expected unquoted identifier", id, result)
			}
		})
	}
}

// TestEdgeCaseIdentifiers tests edge cases that should always be quoted
func TestEdgeCaseIdentifiers(t *testing.T) {
	edgeCases := []struct {
		name       string
		identifier string
		shouldQuote bool
	}{
		{"empty string", "", true},
		{"starts with number", "123col", true},
		{"contains space", "my col", true},
		{"contains hyphen", "my-col", true},
		{"contains dot", "my.col", true},
		{"contains special char", "my@col", true},
		{"just underscore", "_", false},
		{"starts with underscore", "_col", false},
		{"contains dollar", "my$col", false},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			result := QuoteIdentifierIfNeeded(tc.identifier)
			isQuoted := len(result) >= 2 && result[0] == '"' && result[len(result)-1] == '"'
			
			if tc.shouldQuote && !isQuoted {
				t.Errorf("QuoteIdentifierIfNeeded(%q) = %q, expected quoted result", tc.identifier, result)
			}
			if !tc.shouldQuote && isQuoted {
				t.Errorf("QuoteIdentifierIfNeeded(%q) = %q, expected unquoted result", tc.identifier, result)
			}
		})
	}
}

// TestDoubleQuoteEscaping verifies that double quotes in identifiers are properly escaped
func TestDoubleQuoteEscaping(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		expected   string
	}{
		{"single quote", "my\"col", "\"my\"\"col\""},
		{"double quote", "my\"\"col", "\"my\"\"\"\"col\""},
		{"quote at start", "\"col", "\"\"\"col\""},
		{"quote at end", "col\"", "\"col\"\"\""},
		{"only quotes", "\"\"", "\"\"\"\"\"\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := QuoteIdentifier(tt.identifier)
			if result != tt.expected {
				t.Errorf("QuoteIdentifier(%q) = %q, want %q", tt.identifier, result, tt.expected)
			}
		})
	}
}
