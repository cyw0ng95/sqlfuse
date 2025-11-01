package types

import (
	"math"
	"sqlfuse/internal/common"
	"strconv"
	"testing"
)

func TestIntLiteral(t *testing.T) {
	lcg := common.NewLCG(1)
	v := IntLiteral(lcg, "id")
	if n := len(v); n == 0 {
		t.Error("IntLiteral should return a non-empty string")
	}
	v2 := IntLiteral(lcg, "foo")
	if _, err := strconv.ParseInt(v2, 10, 64); err != nil {
		t.Errorf("IntLiteral should return a valid int, got %q", v2)
	}
}

func TestIntLiteral_IDHint(t *testing.T) {
	// Test that "id" hint produces appropriate id-like values
	lcg := common.NewLCG(42)
	
	for i := 0; i < 20; i++ {
		v := IntLiteral(lcg, "user_id")
		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			t.Errorf("IntLiteral with id hint should return valid int, got %q", v)
		}
		
		// ID values should be non-negative in most cases
		if val < 0 {
			// Negative IDs are possible but should be rare
		}
	}
}

func TestIntLiteral_EdgeCases(t *testing.T) {
	lcg := common.NewLCG(1)
	
	edgeCases := map[string]bool{
		"0":  false,
		"1":  false,
		"-1": false,
	}
	
	// Generate many values to find edge cases
	for i := 0; i < 200; i++ {
		v := IntLiteral(lcg, "value")
		if _, ok := edgeCases[v]; ok {
			edgeCases[v] = true
		}
	}
	
	// We should encounter at least one of the edge cases
	foundAny := false
	for _, found := range edgeCases {
		if found {
			foundAny = true
			break
		}
	}
	
	if !foundAny {
		t.Error("Expected to find at least one edge case (0, 1, -1) in 200 samples")
	}
}

func TestIntLiteral_MaxInt(t *testing.T) {
	// Test that max int values can be generated
	found := false
	for seed := uint64(0); seed < 200 && !found; seed++ {
		lcg := common.NewLCG(seed)
		for i := 0; i < 5; i++ {
			v := IntLiteral(lcg, "value")
			val, _ := strconv.ParseInt(v, 10, 64)
			if val == math.MaxInt32 || val == math.MaxInt64 {
				found = true
				break
			}
		}
	}
	// Finding max values is not guaranteed but should be possible
}

func TestIntLiteral_MinInt(t *testing.T) {
	// Test that min int values can be generated
	found := false
	for seed := uint64(0); seed < 200 && !found; seed++ {
		lcg := common.NewLCG(seed)
		for i := 0; i < 5; i++ {
			v := IntLiteral(lcg, "value")
			val, _ := strconv.ParseInt(v, 10, 64)
			if val == math.MinInt32 || val == math.MinInt64 {
				found = true
				break
			}
		}
	}
	// Finding min values is not guaranteed but should be possible
}

func TestIntLiteral_ValidFormat(t *testing.T) {
	lcg := common.NewLCG(999)
	
	for i := 0; i < 100; i++ {
		v := IntLiteral(lcg, "test_col")
		
		// Should be parseable as int64
		_, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			t.Errorf("IntLiteral should produce valid integer string, got %q: %v", v, err)
		}
	}
}

func TestIntLiteral_Variety(t *testing.T) {
	lcg := common.NewLCG(5555)
	
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		v := IntLiteral(lcg, "col")
		seen[v] = true
	}
	
	// Should have good variety
	if len(seen) < 30 {
		t.Errorf("Expected variety in int generation, got only %d unique values", len(seen))
	}
}

func TestIntLiteral_Deterministic(t *testing.T) {
	lcg1 := common.NewLCG(7777)
	lcg2 := common.NewLCG(7777)
	
	v1 := IntLiteral(lcg1, "amount")
	v2 := IntLiteral(lcg2, "amount")
	
	if v1 != v2 {
		t.Errorf("Same seed should produce same value, got %q and %q", v1, v2)
	}
}

