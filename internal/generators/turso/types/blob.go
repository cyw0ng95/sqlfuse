package types

import (
	"fmt"
	"sqlsmith-go/internal/common"
)

// BlobLiteral returns a hex blob literal
func BlobLiteral(lcg *common.LCG) string {
	v := lcg.Uint64()
	return fmt.Sprintf("X'%016x'", v)
}
