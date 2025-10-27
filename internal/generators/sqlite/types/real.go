package types

import (
	"fmt"
	"math"
	"sqlsmith-go/internal/common"
)

// RealLiteral returns a more complex floating point literal with edge cases.
func RealLiteral(lcg *common.LCG) string {
	// Randomly choose between normal, negative, zero, large values, sci notation, and edge cases
	// Note: Inf and NaN are not supported by SQLite/LibSQL
	choice := lcg.Intn(20)
	switch choice {
	case 0:
		return "0.0"
	case 1:
		return "-0.0"
	case 2:
		return "1.0"
	case 3:
		return "-1.0"
	case 4:
		// Very small positive value (close to zero)
		return fmt.Sprintf("%e", 1.0/float64(lcg.Intn(1000000)+1))
	case 5:
		// Very small negative value (close to zero)
		return fmt.Sprintf("%e", -1.0/float64(lcg.Intn(1000000)+1))
	case 6:
		// Pi
		return fmt.Sprintf("%f", math.Pi)
	case 7:
		// E (Euler's number)
		return fmt.Sprintf("%f", math.E)
	case 8:
		// Golden ratio
		return fmt.Sprintf("%f", (1.0+math.Sqrt(5.0))/2.0)
	case 9:
		// Fraction
		numerator := float64(lcg.Intn(100))
		denominator := float64(1 + lcg.Intn(99))
		return fmt.Sprintf("%f", numerator/denominator)
	case 10:
		// Negative fraction
		numerator := float64(lcg.Intn(100))
		denominator := float64(1 + lcg.Intn(99))
		return fmt.Sprintf("%f", -numerator/denominator)
	case 11:
		// Scientific notation - small exponent
		mantissa := float64(lcg.Intn(1000)) / 100.0
		exp := lcg.Intn(5) - 2 // -2 to 2
		return fmt.Sprintf("%.2fe%d", mantissa, exp)
	case 12:
		// Scientific notation - moderate exponent (keep within safe range for SQLite)
		mantissa := float64(lcg.Intn(1000)) / 100.0
		exp := 3 + lcg.Intn(7) // 3..9 to stay well within REAL range
		return fmt.Sprintf("%.2fe%d", mantissa, exp)
	case 13:
		// Negative scientific notation
		return fmt.Sprintf("%e", -float64(lcg.Intn(100000))/100.0)
	case 14:
		// Moderately large positive (keep well within SQLite REAL range)
		return fmt.Sprintf("%f", 1e9*float64(lcg.Intn(100)))
	case 15:
		// Moderately large negative (keep well within SQLite REAL range)
		return fmt.Sprintf("%f", -1e9*float64(lcg.Intn(100)))
	case 16:
		// Decimal with many digits
		return fmt.Sprintf("%.10f", float64(lcg.Intn(1000000))/123456.789)
	case 17:
		// Small positive value
		return fmt.Sprintf("%f", float64(lcg.Intn(100))/100.0)
	case 18:
		// Medium positive value
		return fmt.Sprintf("%f", float64(lcg.Intn(100000))/100.0)
	default:
		// General range
		return fmt.Sprintf("%f", float64(lcg.Intn(1000000)-500000)/100.0)
	}
}
