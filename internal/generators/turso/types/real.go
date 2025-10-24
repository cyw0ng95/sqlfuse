package types

import (
	"fmt"
	"sqlsmith-go/internal/common"
)

// RealLiteral returns a floating point literal.
func RealLiteral(lcg *common.LCG) string {
	v := float64(lcg.Intn(100000)) / 100.0
	return fmt.Sprintf("%f", v)
}
