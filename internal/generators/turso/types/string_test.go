package types

import (
	"sqlsmith-go/internal/common"
	"testing"
)

func TestStringLiteral_Lengths(t *testing.T) {
	lcg := common.NewLCG(99)
	lengths := make(map[int]struct{})
	for i := 0; i < 1000; i++ {
		s := StringLiteral(lcg)
		lengths[len(s)] = struct{}{}
	}
	if len(lengths) < 5 {
		t.Errorf("Expected at least 5 different string lengths, got %d", len(lengths))
	}
}
