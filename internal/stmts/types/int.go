package types

import (
	"fmt"
	"math"
	"sqlfuse/internal/common"
	"strings"
)

// IntLiteral returns an integer literal appropriate for a column with more edge cases.
// `hint` may contain substrings like "id" to bias the generated value.
func IntLiteral(lcg *common.LCG, hint string) string {
	h := strings.ToLower(hint)
	if strings.Contains(h, "id") {
		// small, likely-sequential id-like values with occasional edge cases
		choice := lcg.Intn(10)
		switch choice {
		case 0:
			return "0" // minimum id
		case 1:
			return "1" // first id
		case 2:
			return fmt.Sprintf("%d", math.MaxInt32) // large id
		default:
			return fmt.Sprintf("%d", 1+lcg.Intn(10000))
		}
	}

	// Generate various integer patterns including edge cases
	choice := lcg.Intn(15)
	switch choice {
	case 0:
		return "0"
	case 1:
		return "1"
	case 2:
		return "-1"
	case 3:
		return fmt.Sprintf("%d", math.MaxInt32) // 2147483647
	case 4:
		return fmt.Sprintf("%d", math.MinInt32) // -2147483648
	case 5:
		return fmt.Sprintf("%d", math.MaxInt64) // 9223372036854775807
	case 6:
		return fmt.Sprintf("%d", math.MinInt64) // -9223372036854775808
	case 7:
		// Small positive
		return fmt.Sprintf("%d", lcg.Intn(100))
	case 8:
		// Small negative
		return fmt.Sprintf("%d", -lcg.Intn(100))
	case 9:
		// Medium positive
		return fmt.Sprintf("%d", lcg.Intn(10000))
	case 10:
		// Medium negative
		return fmt.Sprintf("%d", -lcg.Intn(10000))
	case 11:
		// Large positive
		return fmt.Sprintf("%d", lcg.Intn(1000000000))
	case 12:
		// Large negative
		return fmt.Sprintf("%d", -lcg.Intn(1000000000))
	case 13:
		// Power of 2
		power := lcg.Intn(30)
		return fmt.Sprintf("%d", 1<<power)
	default:
		// General range
		return fmt.Sprintf("%d", int64(lcg.Intn(1000000))-500000)
	}
}
