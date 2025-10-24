package types

import (
	"fmt"
	"sqlsmith-go/internal/common"
	"strings"
)

// IntLiteral returns an integer literal appropriate for a column.
// `hint` may contain substrings like "id" to bias the generated value.
func IntLiteral(lcg *common.LCG, hint string) string {
	h := strings.ToLower(hint)
	if strings.Contains(h, "id") {
		// small, likely-sequential id-like values
		return fmt.Sprintf("%d", 1+lcg.Intn(10000))
	}
	// general integer magnitude
	return fmt.Sprintf("%d", int64(lcg.Intn(1000000))-500000)
}
