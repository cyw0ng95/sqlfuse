package turso

import "sqlsmith-go/internal/generators"

// Info implements generators.GeneratorInfo for the turso generator.
var Info generators.GeneratorInfo = &tursoInfo{}

type tursoInfo struct{}

func (t *tursoInfo) Name() string { return "turso" }

func (t *tursoInfo) SupportedStmts() map[string]uint64 {
	w := DefaultStmtWeights()
	out := make(map[string]uint64, len(w))
	for k, v := range w {
		out[string(k)] = v
	}
	return out
}
