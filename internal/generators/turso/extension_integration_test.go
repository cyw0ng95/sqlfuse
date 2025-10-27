package turso

import (
	"database/sql"
	"strings"
	"testing"

	_ "github.com/tursodatabase/turso-go"
)

// TestExtensionFunctionGeneration verifies that extension functions are generated
func TestExtensionFunctionGeneration(t *testing.T) {
	db, err := sql.Open("turso", "file::memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	gen := NewGenerator(12345)

	// Generate many statements to ensure we hit extension functions
	extensionCounts := map[string]int{}
	totalStatements := 500

	for i := 0; i < totalStatements; i++ {
		sql := gen.GenerateWithDB(db)

		// Check for extension functions
		if containsUUID(sql) {
			extensionCounts["uuid"]++
		}
		if containsRegexp(sql) {
			extensionCounts["regexp"]++
		}
		if containsVector(sql) {
			extensionCounts["vector"]++
		}
		if containsTime(sql) {
			extensionCounts["time"]++
		}
	}

	t.Logf("Extension function generation counts (out of %d queries):", totalStatements)
	for ext, count := range extensionCounts {
		t.Logf("  %s: %d (%.1f%%)", ext, count, float64(count)/float64(totalStatements)*100)
	}

	// We expect at least some extension functions to be generated
	if len(extensionCounts) == 0 {
		t.Errorf("No extension functions were generated in %d attempts", totalStatements)
	}

	// Check that each type of extension was generated at least once
	for _, ext := range []string{"uuid", "regexp", "vector", "time"} {
		if extensionCounts[ext] > 0 {
			t.Logf("✓ %s extension functions were generated", ext)
		} else {
			t.Logf("⚠ %s extension functions were NOT generated (may need weight adjustment)", ext)
		}
	}
}

func containsUUID(sql string) bool {
	keywords := []string{"uuid4()", "uuid7()", "uuid_str", "uuid_blob"}
	for _, kw := range keywords {
		if strings.Contains(sql, kw) {
			return true
		}
	}
	return false
}

func containsRegexp(sql string) bool {
	keywords := []string{"regexp(", "regexp_like", "regexp_substr", "regexp_capture", "regexp_replace"}
	for _, kw := range keywords {
		if strings.Contains(sql, kw) {
			return true
		}
	}
	return false
}

func containsVector(sql string) bool {
	keywords := []string{"vector(", "vector32", "vector64", "vector_distance", "vector_concat", "vector_slice"}
	for _, kw := range keywords {
		if strings.Contains(sql, kw) {
			return true
		}
	}
	return false
}

func containsTime(sql string) bool {
	keywords := []string{"time_now", "time_date", "time_unix", "time_get", "time_to_", "time_add", "time_fmt", "dur_"}
	for _, kw := range keywords {
		if strings.Contains(sql, kw) {
			return true
		}
	}
	return false
}
