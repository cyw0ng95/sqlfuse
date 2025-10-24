package types

import (
	"fmt"
	"sqlsmith-go/internal/common"
)

// StringLiteral returns a quoted string literal driven by the LCG.
func StringLiteral(lcg *common.LCG) string {
	v := lcg.Uint64()
	return fmt.Sprintf("'%x'", v)
}
