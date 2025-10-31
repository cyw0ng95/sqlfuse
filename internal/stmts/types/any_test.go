package types

import (
	"sqlfuse/internal/common"
	"testing"
)

func TestValueForType(t *testing.T) {
	lcg := common.NewLCG(5)
	_ = ValueForType("int", lcg, "")
	_ = ValueForType("real", lcg, "")
	_ = ValueForType("blob", lcg, "")
	_ = ValueForType("text", lcg, "")
	_ = ValueForType("json", lcg, "")
	_ = ValueForType("", lcg, "")
	// Just check that it returns a non-empty string
	if s := ValueForType("int", lcg, ""); len(s) == 0 {
		t.Error("ValueForType should return a non-empty string")
	}
	// Check that JSON type returns JSON literal
	jsonVal := ValueForType("JSON", lcg, "")
	if len(jsonVal) == 0 {
		t.Error("ValueForType for JSON should return a non-empty string")
	}
	// JSON values should be quoted
	if jsonVal[0] != '\'' {
		t.Errorf("ValueForType for JSON should return a quoted string, got: %s", jsonVal)
	}
}
