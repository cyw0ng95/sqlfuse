package types

import (
	"encoding/json"
	"sqlfuse/internal/common"
	"strings"
	"testing"
)

func TestJSONLiteral(t *testing.T) {
	lcg := common.NewLCG(42)

	for i := 0; i < 100; i++ {
		result := JSONLiteral(lcg)

		// Check that result is non-empty
		if result == "" {
			t.Error("JSONLiteral returned empty string")
		}

		// Check that result is quoted (SQL string literal)
		if !strings.HasPrefix(result, "'") || !strings.HasSuffix(result, "'") {
			t.Errorf("JSONLiteral result should be quoted: %s", result)
		}

		// Extract JSON content (remove outer quotes)
		jsonContent := result[1 : len(result)-1]

		// Validate that it's valid JSON
		var parsed interface{}
		if err := json.Unmarshal([]byte(jsonContent), &parsed); err != nil {
			t.Errorf("JSONLiteral produced invalid JSON on iteration %d: %s\nError: %v", i, result, err)
		}
	}
}

func TestJSONLiteralDeterminism(t *testing.T) {
	// Same seed should produce same results
	lcg1 := common.NewLCG(12345)
	lcg2 := common.NewLCG(12345)

	for i := 0; i < 10; i++ {
		result1 := JSONLiteral(lcg1)
		result2 := JSONLiteral(lcg2)

		if result1 != result2 {
			t.Errorf("Same seed produced different results on iteration %d:\n  First:  %s\n  Second: %s",
				i, result1, result2)
		}
	}
}

func TestJSONLiteralVariety(t *testing.T) {
	lcg := common.NewLCG(9999)

	// Track different patterns we've seen
	seenPatterns := make(map[string]bool)

	for i := 0; i < 100; i++ {
		result := JSONLiteral(lcg)

		// Extract JSON content
		jsonContent := result[1 : len(result)-1]

		// Classify pattern
		var parsed interface{}
		if err := json.Unmarshal([]byte(jsonContent), &parsed); err == nil {
			switch v := parsed.(type) {
			case map[string]interface{}:
				seenPatterns["object"] = true
			case []interface{}:
				seenPatterns["array"] = true
			default:
				t.Errorf("Unexpected JSON type: %T, value: %v", v, v)
			}
		}
	}

	// We should see both objects and arrays
	if !seenPatterns["object"] {
		t.Error("Did not generate any JSON objects")
	}
	if !seenPatterns["array"] {
		t.Error("Did not generate any JSON arrays")
	}
}

func TestJSONLiteralEdgeCases(t *testing.T) {
	lcg := common.NewLCG(111)

	// Run many iterations to increase likelihood of hitting edge cases
	seenEmpty := false
	seenNested := false

	for i := 0; i < 200; i++ {
		result := JSONLiteral(lcg)
		jsonContent := result[1 : len(result)-1]

		// Check for empty object/array
		if jsonContent == "{}" || jsonContent == "[]" {
			seenEmpty = true
		}

		// Check for nested structures (simple heuristic: contains nested braces)
		if strings.Count(jsonContent, "{") > 1 || strings.Count(jsonContent, "[") > 1 {
			seenNested = true
		}
	}

	if !seenEmpty {
		t.Log("Did not generate empty objects/arrays (may be rare but acceptable)")
	}
	if !seenNested {
		t.Log("Did not generate nested structures (may be rare but acceptable)")
	}
}
