package keywords

import (
	"testing"
)

func TestIsKeyword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"SELECT uppercase", "SELECT", true},
		{"select lowercase", "select", true},
		{"SeLeCt mixed case", "SeLeCt", true},
		{"not a keyword", "mycol", false},
		{"empty string", "", false},
		{"ABORT", "ABORT", true},
		{"WINDOW", "WINDOW", true},
		{"WITH", "WITH", true},
		{"WITHOUT", "WITHOUT", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsKeyword(tt.input)
			if result != tt.expected {
				t.Errorf("IsKeyword(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsReserved(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"SELECT is reserved", "SELECT", true},
		{"INSERT is reserved", "INSERT", true},
		{"not a keyword", "mycol", false},
		{"case insensitive", "select", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsReserved(tt.input)
			if result != tt.expected {
				t.Errorf("IsReserved(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsTursoSupported(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"SELECT is Turso supported", "SELECT", true},
		{"FILTER is not Turso supported", "FILTER", false},
		{"MATCH is not Turso supported", "MATCH", false},
		{"MATERIALIZED is not Turso supported", "MATERIALIZED", false},
		{"OVER is not Turso supported", "OVER", false},
		{"PARTITION is not Turso supported", "PARTITION", false},
		{"RAISE is not Turso supported", "RAISE", false},
		{"RECURSIVE is not Turso supported", "RECURSIVE", false},
		{"REGEXP is not Turso supported", "REGEXP", false},
		{"WINDOW is not Turso supported", "WINDOW", false},
		{"not a keyword returns true", "mycol", true},
		{"case insensitive", "filter", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTursoSupported(tt.input)
			if result != tt.expected {
				t.Errorf("IsTursoSupported(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetKeyword(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectNil     bool
		expectedName  string
		expectedTurso bool
	}{
		{"SELECT keyword", "SELECT", false, "SELECT", true},
		{"FILTER keyword", "FILTER", false, "FILTER", false},
		{"not a keyword", "mycol", true, "", false},
		{"case insensitive", "select", false, "SELECT", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetKeyword(tt.input)
			if tt.expectNil {
				if result != nil {
					t.Errorf("GetKeyword(%q) = %v, want nil", tt.input, result)
				}
			} else {
				if result == nil {
					t.Errorf("GetKeyword(%q) = nil, want non-nil", tt.input)
					return
				}
				if result.Name != tt.expectedName {
					t.Errorf("GetKeyword(%q).Name = %q, want %q", tt.input, result.Name, tt.expectedName)
				}
				if result.TursoSupport != tt.expectedTurso {
					t.Errorf("GetKeyword(%q).TursoSupport = %v, want %v", tt.input, result.TursoSupport, tt.expectedTurso)
				}
			}
		})
	}
}

func TestAllKeywords(t *testing.T) {
	keywords := AllKeywords()
	
	// Verify we have all expected keywords (147 as of SQLite 3.x)
	if len(keywords) < 140 {
		t.Errorf("AllKeywords() returned %d keywords, expected at least 140", len(keywords))
	}
	
	// Verify some known keywords exist
	keywordNames := make(map[string]bool)
	for _, kw := range keywords {
		keywordNames[kw.Name] = true
	}
	
	expectedKeywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "DROP",
		"ALTER", "TABLE", "INDEX", "VIEW", "WHERE", "FROM",
		"JOIN", "LEFT", "RIGHT", "INNER", "OUTER", "ON",
		"GROUP", "ORDER", "BY", "HAVING", "LIMIT", "OFFSET",
		"UNION", "INTERSECT", "EXCEPT", "WITH", "AS",
		"FILTER", "WINDOW", "PARTITION", "OVER", "RECURSIVE",
		"MATERIALIZED", "REGEXP", "MATCH", "RAISE",
	}
	
	for _, expected := range expectedKeywords {
		if !keywordNames[expected] {
			t.Errorf("AllKeywords() missing expected keyword: %s", expected)
		}
	}
}

func TestToUpper(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "select", "SELECT"},
		{"uppercase", "SELECT", "SELECT"},
		{"mixed case", "SeLeCt", "SELECT"},
		{"with underscore", "current_date", "CURRENT_DATE"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toUpper(tt.input)
			if result != tt.expected {
				t.Errorf("toUpper(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestKeywordCount(t *testing.T) {
	// Verify the keyword map is initialized correctly
	if len(keywordMap) != len(allKeywords) {
		t.Errorf("keywordMap has %d entries, allKeywords has %d entries", len(keywordMap), len(allKeywords))
	}
}

func TestTursoUnsupportedKeywords(t *testing.T) {
	// List of keywords that should be marked as Turso unsupported
	unsupportedKeywords := []string{
		"FILTER", "MATCH", "MATERIALIZED", "OVER", "PARTITION",
		"RAISE", "RECURSIVE", "REGEXP", "WINDOW",
	}
	
	for _, keyword := range unsupportedKeywords {
		kw := GetKeyword(keyword)
		if kw == nil {
			t.Errorf("Keyword %s not found in keyword map", keyword)
			continue
		}
		if kw.TursoSupport {
			t.Errorf("Keyword %s should be marked as Turso unsupported", keyword)
		}
	}
}

func TestTursoSupportedKeywords(t *testing.T) {
	// List of common keywords that should be marked as Turso supported
	supportedKeywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "DROP",
		"ALTER", "TABLE", "WHERE", "FROM", "JOIN", "ON",
		"GROUP", "ORDER", "BY", "HAVING", "LIMIT", "OFFSET",
		"UNION", "WITH", "AS", "BEGIN", "COMMIT", "ROLLBACK",
	}
	
	for _, keyword := range supportedKeywords {
		kw := GetKeyword(keyword)
		if kw == nil {
			t.Errorf("Keyword %s not found in keyword map", keyword)
			continue
		}
		if !kw.TursoSupport {
			t.Errorf("Keyword %s should be marked as Turso supported", keyword)
		}
	}
}
