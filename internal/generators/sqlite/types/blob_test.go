package types

import (
	"sqlsmith-go/internal/common"
	"strings"
	"testing"
)

func TestBlobLiteral(t *testing.T) {
	lcg := common.NewLCG(4)
	v := BlobLiteral(lcg)
	if !strings.HasPrefix(v, "X'") || !strings.HasSuffix(v, "'") {
		t.Errorf("BlobLiteral should be hex quoted: %q", v)
	}
}
