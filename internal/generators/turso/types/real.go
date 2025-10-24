package types

import (
	"fmt"
	"math"
	"sqlsmith-go/internal/common"
)

// RealLiteral returns a more complex floating point literal.
func RealLiteral(lcg *common.LCG) string {
	// Randomly choose between normal, negative, zero, inf, nan, sci notation
	switch lcg.Intn(6) {
	case 0:
		return "0.0"
	case 1:
		return fmt.Sprintf("%f", -float64(lcg.Intn(100000))/100.0)
	case 2:
		return fmt.Sprintf("%e", float64(lcg.Intn(100000))/100.0)
	case 3:
		return fmt.Sprintf("%e", -float64(lcg.Intn(100000))/100.0)
	case 4:
		return fmt.Sprintf("%f", math.Inf(1))
	case 5:
		return fmt.Sprintf("%f", math.Inf(-1))
	default:
		return fmt.Sprintf("%f", float64(lcg.Intn(1000000)-500000)/100.0)
	}
}
