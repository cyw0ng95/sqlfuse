package types

import (
	"fmt"
	"sqlfuse/internal/common"
)

// BlobLiteral returns a hex blob literal with various patterns.
func BlobLiteral(lcg *common.LCG) string {
	choice := lcg.Intn(10)
	switch choice {
	case 0:
		// Empty blob
		return "X''"
	case 1:
		// Single byte
		return fmt.Sprintf("X'%02X'", lcg.Intn(256))
	case 2:
		// Two bytes
		return fmt.Sprintf("X'%02X%02X'", lcg.Intn(256), lcg.Intn(256))
	case 3:
		// Four bytes (int32)
		return fmt.Sprintf("X'%02X%02X%02X%02X'", lcg.Intn(256), lcg.Intn(256), lcg.Intn(256), lcg.Intn(256))
	case 4:
		// Eight bytes (int64) - original format
		v := lcg.Uint64()
		return fmt.Sprintf("X'%016x'", v)
	case 5:
		// All zeros
		length := 1 + lcg.Intn(8)
		result := "X'"
		for i := 0; i < length; i++ {
			result += "00"
		}
		result += "'"
		return result
	case 6:
		// All ones (0xFF)
		length := 1 + lcg.Intn(8)
		result := "X'"
		for i := 0; i < length; i++ {
			result += "FF"
		}
		result += "'"
		return result
	case 7:
		// Pattern (0xAA or 0x55)
		length := 1 + lcg.Intn(8)
		pattern := "AA"
		if lcg.Intn(2) == 0 {
			pattern = "55"
		}
		result := "X'"
		for i := 0; i < length; i++ {
			result += pattern
		}
		result += "'"
		return result
	case 8:
		// Random medium length (16-32 bytes)
		length := 16 + lcg.Intn(17)
		result := "X'"
		for i := 0; i < length; i++ {
			result += fmt.Sprintf("%02X", lcg.Intn(256))
		}
		result += "'"
		return result
	default:
		// Small random blob (4-16 bytes)
		length := 4 + lcg.Intn(13)
		result := "X'"
		for i := 0; i < length; i++ {
			result += fmt.Sprintf("%02X", lcg.Intn(256))
		}
		result += "'"
		return result
	}
}
