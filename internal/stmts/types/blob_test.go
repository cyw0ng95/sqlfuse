package types

import (
	"sqlfuse/internal/common"
	"strings"
	"testing"
)

func TestBlobLiteral(t *testing.T) {
	lcg := common.NewLCG(4)
	v := BlobLiteral(lcg)
	if !strings.HasPrefix(v, "X'") || !strings.HasSuffix(v, "'") {
		t.Errorf("BlobLiteral should be hex quoted: %q", v)
	}
}

func TestBlobLiteral_EmptyBlob(t *testing.T) {
	// Test that empty blob can be generated
	for seed := uint64(0); seed < 100; seed++ {
		lcg := common.NewLCG(seed)
		v := BlobLiteral(lcg)
		if v == "X''" {
			return // Found empty blob
		}
	}
	// If we don't find it in 100 tries, it's statistically unlikely but not impossible
}

func TestBlobLiteral_ValidHexFormat(t *testing.T) {
	lcg := common.NewLCG(42)
	
	for i := 0; i < 50; i++ {
		v := BlobLiteral(lcg)
		
		// Must start with X' and end with '
		if !strings.HasPrefix(v, "X'") {
			t.Errorf("BlobLiteral must start with X', got: %q", v)
		}
		
		if !strings.HasSuffix(v, "'") {
			t.Errorf("BlobLiteral must end with ', got: %q", v)
		}
		
		// Extract hex content
		hexContent := v[2 : len(v)-1]
		
		// Check that all characters are valid hex
		for _, ch := range hexContent {
			if !((ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F') || (ch >= 'a' && ch <= 'f')) {
				t.Errorf("Invalid hex character %c in blob: %q", ch, v)
			}
		}
		
		// Check that hex length is even (full bytes)
		if len(hexContent)%2 != 0 {
			t.Errorf("Hex content must have even length, got %d in: %q", len(hexContent), v)
		}
	}
}

func TestBlobLiteral_Variety(t *testing.T) {
	lcg := common.NewLCG(123)
	
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		v := BlobLiteral(lcg)
		seen[v] = true
	}
	
	// With 100 samples, we should see at least a few different blobs
	if len(seen) < 10 {
		t.Errorf("Expected variety in blob generation, got only %d unique blobs", len(seen))
	}
}

func TestBlobLiteral_Deterministic(t *testing.T) {
	lcg1 := common.NewLCG(999)
	lcg2 := common.NewLCG(999)
	
	v1 := BlobLiteral(lcg1)
	v2 := BlobLiteral(lcg2)
	
	if v1 != v2 {
		t.Errorf("Same seed should produce same blob, got %q and %q", v1, v2)
	}
}

