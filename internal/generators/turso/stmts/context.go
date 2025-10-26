package stmts

import (
	"database/sql"
	"sqlsmith-go/internal/common"
)

// GenContext provides context for recursive SQL generation.
// It tracks recursion depth to prevent infinite recursion and provides
// access to the database connection and random number generator.
type GenContext struct {
	DB       *sql.DB
	LCG      *common.LCG
	Depth    int // Current recursion depth
	MaxDepth int // Maximum allowed recursion depth
}

// NewGenContext creates a new generation context with the given database and LCG.
// maxDepth controls how deep recursive generation can go (0 = no recursion).
func NewGenContext(db *sql.DB, lcg *common.LCG, maxDepth int) *GenContext {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}
	return &GenContext{
		DB:       db,
		LCG:      lcg,
		Depth:    0,
		MaxDepth: maxDepth,
	}
}

// CanRecurse returns true if we haven't reached the maximum recursion depth.
func (ctx *GenContext) CanRecurse() bool {
	return ctx.Depth < ctx.MaxDepth
}

// Descend creates a new context with incremented depth for recursive calls.
func (ctx *GenContext) Descend() *GenContext {
	return &GenContext{
		DB:       ctx.DB,
		LCG:      ctx.LCG,
		Depth:    ctx.Depth + 1,
		MaxDepth: ctx.MaxDepth,
	}
}

// Intn is a convenience method to get a random integer from the LCG.
func (ctx *GenContext) Intn(n int) int {
	if ctx.LCG != nil {
		return ctx.LCG.Intn(n)
	}
	return 0
}
