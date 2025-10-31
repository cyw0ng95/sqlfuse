package types

import (
	"fmt"
	"sqlfuse/internal/common"
	"strings"
)

// JSONLiteral returns a JSON literal string for use in SQL statements.
// It generates various JSON structures including objects, arrays, and scalar values.
func JSONLiteral(lcg *common.LCG) string {
	choice := lcg.Intn(10)

	switch choice {
	case 0:
		// Simple object
		return fmt.Sprintf("'{\"id\":%d,\"name\":\"item%d\"}'", lcg.Intn(1000), lcg.Intn(1000))
	case 1:
		// Array of numbers
		numElements := 2 + lcg.Intn(4)
		elements := make([]string, numElements)
		for i := 0; i < numElements; i++ {
			elements[i] = fmt.Sprintf("%d", lcg.Intn(100))
		}
		return fmt.Sprintf("'[%s]'", strings.Join(elements, ","))
	case 2:
		// Array of strings
		numElements := 2 + lcg.Intn(4)
		elements := make([]string, numElements)
		for i := 0; i < numElements; i++ {
			elements[i] = fmt.Sprintf("\"value%d\"", i)
		}
		return fmt.Sprintf("'[%s]'", strings.Join(elements, ","))
	case 3:
		// Nested object
		return fmt.Sprintf("'{\"user\":{\"id\":%d,\"active\":%t},\"count\":%d}'",
			lcg.Intn(1000), lcg.Intn(2) == 0, lcg.Intn(100))
	case 4:
		// Empty object
		return "'{}'"
	case 5:
		// Empty array
		return "'[]'"
	case 6:
		// Object with array
		numElements := 1 + lcg.Intn(4)
		elements := make([]string, numElements)
		for i := 0; i < numElements; i++ {
			elements[i] = fmt.Sprintf("%d", lcg.Intn(100))
		}
		return fmt.Sprintf("'{\"items\":[%s],\"total\":%d}'",
			strings.Join(elements, ","), numElements)
	case 7:
		// Object with null value
		return fmt.Sprintf("'{\"value\":null,\"id\":%d}'", lcg.Intn(1000))
	case 8:
		// Object with boolean
		return fmt.Sprintf("'{\"enabled\":%t,\"id\":%d}'", lcg.Intn(2) == 0, lcg.Intn(1000))
	default:
		// Complex nested structure
		return fmt.Sprintf("'{\"data\":{\"values\":[%d,%d,%d],\"status\":\"active\"},\"timestamp\":%d}'",
			lcg.Intn(100), lcg.Intn(100), lcg.Intn(100), 1609459200+lcg.Intn(63072000))
	}
}
