package common

import (
	"testing"
)

func TestLCG_Uint64_Sequence(t *testing.T) {
	lcg := NewLCG(1)
	v1 := lcg.Uint64()
	v2 := lcg.Uint64()
	v3 := lcg.Uint64()
	if v1 == v2 || v2 == v3 || v1 == v3 {
		t.Error("Uint64 should produce different values on each call")
	}
}

func TestLCG_Intn_Range(t *testing.T) {
	lcg := NewLCG(42)
	for n := 1; n <= 100; n++ {
		v := lcg.Intn(n)
		if v < 0 || v >= n {
			t.Errorf("Intn(%d) out of range: got %d", n, v)
		}
	}
}

func TestLCG_Float64_Range(t *testing.T) {
	lcg := NewLCG(99)
	for i := 0; i < 100; i++ {
		f := lcg.Float64()
		if f < 0.0 || f >= 1.0 {
			t.Errorf("Float64 out of range: got %f", f)
		}
	}
}

func TestLCG_TokensUsed(t *testing.T) {
	lcg := NewLCG(123)
	for i := 0; i < 10; i++ {
		_ = lcg.Uint64()
	}
	if lcg.TokensUsed() != 10 {
		t.Errorf("TokensUsed should be 10, got %d", lcg.TokensUsed())
	}
}

func TestLCG_Seed(t *testing.T) {
	lcg := NewLCG(0)
	lcg.Seed(100)
	v := lcg.Uint64()
	lcg.Seed(100)
	v2 := lcg.Uint64()
	if v != v2 {
		t.Error("Seed should reset the sequence")
	}
}

func TestLCG_SetSeed(t *testing.T) {
	lcg := NewLCG(0)
	lcg.SetSeed(12345)
	v1 := lcg.Uint64()
	lcg.SetSeed(12345)
	v2 := lcg.Uint64()
	if v1 != v2 {
		t.Error("SetSeed should reset the sequence")
	}
}

func TestLCG_Seed_Negative(t *testing.T) {
	lcg := NewLCG(0)
	lcg.Seed(-1)
	_ = lcg.Uint64() // should not panic
}

func TestLCG_Uint32(t *testing.T) {
	lcg := NewLCG(42)
	v := lcg.Uint32()
	if v == 0 {
		t.Error("Uint32 should produce a non-zero value for non-zero seed")
	}
}

func TestLCG_Int63_TokensUsed(t *testing.T) {
	lcg := NewLCG(1)
	start := lcg.TokensUsed()
	_ = lcg.Int63()
	if lcg.TokensUsed() != start+1 {
		t.Error("Int63 should increment tokens counter")
	}
}

func TestLCG_Float64_TokensUsed(t *testing.T) {
	lcg := NewLCG(1)
	start := lcg.TokensUsed()
	_ = lcg.Float64()
	if lcg.TokensUsed() != start+1 {
		t.Error("Float64 should increment tokens counter")
	}
}

func TestLCG_Intn_ZeroOrNegative(t *testing.T) {
	lcg := NewLCG(1)
	if lcg.Intn(0) != 0 {
		t.Error("Intn(0) should return 0")
	}
	if lcg.Intn(-5) != 0 {
		t.Error("Intn(-5) should return 0")
	}
}
