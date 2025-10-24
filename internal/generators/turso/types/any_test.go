package types

import (
	"sqlsmith-go/internal/common"
	"testing"
)

func TestValueForType(t *testing.T) {
	lcg := common.NewLCG(5)
	_ = ValueForType("int", lcg, "")
	_ = ValueForType("real", lcg, "")
	_ = ValueForType("blob", lcg, "")
	_ = ValueForType("text", lcg, "")
	_ = ValueForType("", lcg, "")
	// Just check that it returns a non-empty string
	if s := ValueForType("int", lcg, ""); len(s) == 0 {
		t.Error("ValueForType should return a non-empty string")
	}
}
