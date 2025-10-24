package types

import (
	"fmt"
	"sqlsmith-go/internal/common"
)

// StringLiteral returns a quoted string literal driven by the LCG.
func StringLiteral(lcg *common.LCG) string {
	return fmt.Sprintf("'%d-%d-%x'", lcg.Uint64(), lcg.Uint64(), lcg.Uint64())
}
