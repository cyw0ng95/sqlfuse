package generators

import "testing"

func TestDepthHelper_CanRecurse(t *testing.T) {
	dh := NewDepthHelper(0, 5)
	if !dh.CanRecurse() {
		t.Error("Should be able to recurse at depth 0 with max 5")
	}

	dh2 := NewDepthHelper(5, 5)
	if dh2.CanRecurse() {
		t.Error("Should not be able to recurse at max depth")
	}

	dh3 := NewDepthHelper(6, 5)
	if dh3.CanRecurse() {
		t.Error("Should not be able to recurse beyond max depth")
	}
}

func TestDepthHelper_ShouldRecurse(t *testing.T) {
	dh := NewDepthHelper(0, 10)
	if !dh.ShouldRecurse(3) {
		t.Error("Should recurse at depth 0 with threshold 3")
	}

	dh2 := NewDepthHelper(3, 10)
	if dh2.ShouldRecurse(3) {
		t.Error("Should not recurse at depth 3 with threshold 3")
	}

	dh3 := NewDepthHelper(10, 10)
	if dh3.ShouldRecurse(20) {
		t.Error("Should not recurse at max depth regardless of threshold")
	}
}

func TestDepthHelper_Descend(t *testing.T) {
	dh := NewDepthHelper(2, 10)
	descended := dh.Descend()

	if descended.GetDepth() != 3 {
		t.Errorf("Expected depth 3 after descend, got %d", descended.GetDepth())
	}

	if descended.GetMaxDepth() != 10 {
		t.Errorf("Expected max depth 10, got %d", descended.GetMaxDepth())
	}

	// Original should be unchanged
	if dh.GetDepth() != 2 {
		t.Error("Original depth helper should not be modified")
	}
}

func TestDepthHelper_AtMaxDepth(t *testing.T) {
	dh := NewDepthHelper(5, 5)
	if !dh.AtMaxDepth() {
		t.Error("Should be at max depth")
	}

	dh2 := NewDepthHelper(4, 5)
	if dh2.AtMaxDepth() {
		t.Error("Should not be at max depth yet")
	}
}

func TestDepthHelper_ShouldGenerateComplex(t *testing.T) {
	// At depth 0 with max 10, should prefer complex
	dh := NewDepthHelper(0, 10)
	complexCount := 0
	for i := 0; i < 100; i++ {
		if dh.ShouldGenerateComplex(3, 8) {
			complexCount++
		}
	}
	// At depth 0, threshold should be close to maxProbability (8)
	// So depth 0 < 8 should always be true
	if complexCount != 100 {
		t.Errorf("At depth 0, should always generate complex (expected 100, got %d)", complexCount)
	}

	// At max depth, should never generate complex
	dh2 := NewDepthHelper(10, 10)
	if dh2.ShouldGenerateComplex(3, 8) {
		t.Error("At max depth, should not generate complex")
	}
}
