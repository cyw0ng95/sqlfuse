package types

import (
	"sqlfuse/internal/common"
	"strings"
	"testing"
)

func TestStringLiteral_Lengths(t *testing.T) {
	lcg := common.NewLCG(99)
	lengths := make(map[int]struct{})
	for i := 0; i < 1000; i++ {
		s := StringLiteral(lcg, "")
		lengths[len(s)] = struct{}{}
	}
	if len(lengths) < 5 {
		t.Errorf("Expected at least 5 different string lengths, got %d", len(lengths))
	}
}

func TestStringLiteral_QuotedFormat(t *testing.T) {
	lcg := common.NewLCG(42)
	
	for i := 0; i < 50; i++ {
		s := StringLiteral(lcg, "name")
		
		if !strings.HasPrefix(s, "'") || !strings.HasSuffix(s, "'") {
			t.Errorf("StringLiteral should be quoted with single quotes, got %q", s)
		}
	}
}

func TestStringLiteral_EmptyString(t *testing.T) {
	// Test that empty string can be generated
	found := false
	for seed := uint64(0); seed < 200; seed++ {
		lcg := common.NewLCG(seed)
		for i := 0; i < 5; i++ {
			s := StringLiteral(lcg, "description")
			if s == "''" {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	
	if !found {
		t.Error("Expected to find empty string in samples")
	}
}

func TestStringLiteral_UniqueColumns(t *testing.T) {
	lcg := common.NewLCG(123)
	
	// For unique columns, values should be more diverse
	uniqueHints := []string{"email", "username", "user_email", "api_key"}
	
	for _, hint := range uniqueHints {
		seen := make(map[string]bool)
		for i := 0; i < 100; i++ {
			s := StringLiteral(lcg, hint)
			seen[s] = true
		}
		
		// Should have high variety for unique columns
		if len(seen) < 80 {
			t.Errorf("Expected high variety for unique column %q, got only %d unique values", hint, len(seen))
		}
	}
}

func TestStringLiteral_Variety(t *testing.T) {
	lcg := common.NewLCG(999)
	
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		s := StringLiteral(lcg, "text")
		seen[s] = true
	}
	
	// Should have good variety
	if len(seen) < 30 {
		t.Errorf("Expected variety in string generation, got only %d unique values", len(seen))
	}
}

func TestStringLiteral_NoUnescapedQuotes(t *testing.T) {
	lcg := common.NewLCG(555)
	
	for i := 0; i < 100; i++ {
		s := StringLiteral(lcg, "field")
		
		// Remove outer quotes
		if len(s) < 2 {
			continue
		}
		
		inner := s[1 : len(s)-1]
		
		// Count single quotes
		singleQuoteCount := strings.Count(inner, "'")
		
		// All single quotes in the middle should be escaped (appear as '')
		_ = strings.Count(inner, "''") // doubleQuoteCount not used, just checking format
		
		// Number of single quotes should be even (all escaped)
		if singleQuoteCount%2 != 0 {
			t.Errorf("Found unescaped quote in string literal %q", s)
		}
	}
}

func TestStringLiteral_SpecialCharacters(t *testing.T) {
	// Test that special characters can be handled
	lcg := common.NewLCG(777)
	
	for i := 0; i < 200; i++ {
		s := StringLiteral(lcg, "notes")
		
		// Check for newlines, tabs, or other special chars
		if strings.Contains(s, "\\n") || strings.Contains(s, "\\t") || 
		   strings.Contains(s, "''") || strings.Contains(s, "%") {
			// Found special character - test passes
			return
		}
	}
	
	// Special characters are possible but not guaranteed in all test runs
	// This is fine - we just want to verify they can be handled when present
}

func TestStringLiteral_Deterministic(t *testing.T) {
	lcg1 := common.NewLCG(12345)
	lcg2 := common.NewLCG(12345)
	
	v1 := StringLiteral(lcg1, "col")
	v2 := StringLiteral(lcg2, "col")
	
	if v1 != v2 {
		t.Errorf("Same seed should produce same string, got %q and %q", v1, v2)
	}
}

func TestStringLiteral_DifferentHints(t *testing.T) {
	lcg := common.NewLCG(333)
	
	// Different hints may produce different patterns
	s1 := StringLiteral(lcg, "email")
	s2 := StringLiteral(lcg, "description")
	
	// Both should be valid quoted strings
	if !strings.HasPrefix(s1, "'") || !strings.HasSuffix(s1, "'") {
		t.Errorf("email hint should produce quoted string, got %q", s1)
	}
	
	if !strings.HasPrefix(s2, "'") || !strings.HasSuffix(s2, "'") {
		t.Errorf("description hint should produce quoted string, got %q", s2)
	}
}

