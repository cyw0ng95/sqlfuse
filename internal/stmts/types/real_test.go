package types

import (
	"math"
	"sqlfuse/internal/common"
	"strconv"
	"strings"
	"testing"
)

func TestRealLiteral(t *testing.T) {
	lcg := common.NewLCG(2)
	for i := 0; i < 10; i++ {
		v := RealLiteral(lcg)
		if !isValidReal(v) {
			t.Errorf("RealLiteral returned invalid value: %q", v)
		}
	}
}

func isValidReal(s string) bool {
	// SQLite/LibSQL doesn't support inf or nan, so we only generate finite values
	if s == "0.0" || strings.Contains(s, "e") {
		return true
	}
	_, err := strconv.ParseFloat(strings.Trim(s, "'"), 64)
	return err == nil
}

func TestRealLiteral_ZeroValues(t *testing.T) {
	// Test that zero values can be generated
	found := false
	for seed := uint64(0); seed < 100; seed++ {
		lcg := common.NewLCG(seed)
		v := RealLiteral(lcg)
		if v == "0.0" || v == "-0.0" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Expected to find zero values in 100 samples")
	}
}

func TestRealLiteral_PositiveAndNegative(t *testing.T) {
	lcg := common.NewLCG(42)
	
	hasPositive := false
	hasNegative := false
	
	for i := 0; i < 50; i++ {
		v := RealLiteral(lcg)
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			t.Errorf("Failed to parse real literal %q: %v", v, err)
			continue
		}
		
		if val > 0 {
			hasPositive = true
		} else if val < 0 {
			hasNegative = true
		}
		
		if hasPositive && hasNegative {
			break
		}
	}
	
	if !hasPositive {
		t.Error("Expected to find at least one positive value")
	}
	
	if !hasNegative {
		t.Error("Expected to find at least one negative value")
	}
}

func TestRealLiteral_ScientificNotation(t *testing.T) {
	// Test that scientific notation can be generated
	found := false
	for seed := uint64(0); seed < 100; seed++ {
		lcg := common.NewLCG(seed)
		for i := 0; i < 5; i++ {
			v := RealLiteral(lcg)
			if strings.Contains(v, "e") || strings.Contains(v, "E") {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	
	if !found {
		t.Error("Expected to find scientific notation in samples")
	}
}

func TestRealLiteral_NoInfOrNaN(t *testing.T) {
	lcg := common.NewLCG(999)
	
	for i := 0; i < 200; i++ {
		v := RealLiteral(lcg)
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			t.Errorf("Failed to parse real literal %q: %v", v, err)
			continue
		}
		
		if math.IsInf(val, 0) {
			t.Errorf("RealLiteral should not generate Inf, got %q", v)
		}
		
		if math.IsNaN(val) {
			t.Errorf("RealLiteral should not generate NaN, got %q", v)
		}
	}
}

func TestRealLiteral_Variety(t *testing.T) {
	lcg := common.NewLCG(12345)
	
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		v := RealLiteral(lcg)
		seen[v] = true
	}
	
	// Should have good variety
	if len(seen) < 30 {
		t.Errorf("Expected variety in real generation, got only %d unique values", len(seen))
	}
}

func TestRealLiteral_ValidFormat(t *testing.T) {
	lcg := common.NewLCG(777)
	
	for i := 0; i < 100; i++ {
		v := RealLiteral(lcg)
		
		// Should be parseable as float64
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			t.Errorf("RealLiteral should produce valid float string, got %q: %v", v, err)
		}
		
		// Should be finite (not Inf or NaN)
		if math.IsInf(val, 0) || math.IsNaN(val) {
			t.Errorf("RealLiteral should produce finite values, got %q", v)
		}
	}
}

func TestRealLiteral_Deterministic(t *testing.T) {
	lcg1 := common.NewLCG(8888)
	lcg2 := common.NewLCG(8888)
	
	v1 := RealLiteral(lcg1)
	v2 := RealLiteral(lcg2)
	
	if v1 != v2 {
		t.Errorf("Same seed should produce same value, got %q and %q", v1, v2)
	}
}

