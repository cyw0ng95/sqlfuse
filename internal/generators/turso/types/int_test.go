package types

import (
	"sqlsmith-go/internal/common"
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
