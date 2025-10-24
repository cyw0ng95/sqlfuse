package types

import (
	"sqlsmith-go/internal/common"
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
	if s == "0.0" || strings.Contains(s, "e") || strings.Contains(s, "inf") || strings.Contains(s, "nan") {
		return true
	}
	_, err := strconv.ParseFloat(strings.Trim(s, "'"), 64)
	return err == nil
}
