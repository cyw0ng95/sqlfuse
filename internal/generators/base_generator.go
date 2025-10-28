package generators

import (
	"database/sql"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/stmts/stmts"
)

// BaseGenerator provides common generator functionality that can be embedded
// by specific generator implementations. It handles weights, LCG state,
// recursion depth, and direction selection using the Strategy pattern.
type BaseGenerator struct {
	lcg               *common.LCG
	weights           map[stmts.StmtType]uint64
	totalWeight       uint64
	maxRecursionDepth int
	firstGeneration   bool
	genMap            map[stmts.StmtType]func(db *sql.DB) (string, error)
}

// NewBaseGenerator creates a new base generator with the given seed.
func NewBaseGenerator(seed uint64) *BaseGenerator {
	return &BaseGenerator{
		lcg:               common.NewLCG(seed),
		maxRecursionDepth: 2, // Default recursion depth
		firstGeneration:   true,
	}
}

// SetWeights replaces the current weights and recalculates totals.
func (g *BaseGenerator) SetWeights(weights map[stmts.StmtType]uint64) {
	if g.weights == nil {
		g.weights = make(map[stmts.StmtType]uint64, len(weights))
	}
	for k, v := range weights {
		g.weights[k] = v
	}
	g.recalcTotalWeight()
}

// SetWeight sets a single statement type weight and updates totals.
func (g *BaseGenerator) SetWeight(t stmts.StmtType, weight uint64) {
	if g.weights == nil {
		g.weights = make(map[stmts.StmtType]uint64)
	}
	g.weights[t] = weight
	g.recalcTotalWeight()
}

// GetWeights returns a copy of the current weights map.
func (g *BaseGenerator) GetWeights() map[stmts.StmtType]uint64 {
	out := make(map[stmts.StmtType]uint64, len(g.weights))
	for k, v := range g.weights {
		out[k] = v
	}
	return out
}

// SetMaxRecursionDepth sets the maximum recursion depth for complex SQL generation.
func (g *BaseGenerator) SetMaxRecursionDepth(depth int) {
	if depth < 0 {
		depth = 0
	}
	g.maxRecursionDepth = depth
}

// GetMaxRecursionDepth returns the current maximum recursion depth.
func (g *BaseGenerator) GetMaxRecursionDepth() int {
	return g.maxRecursionDepth
}

// SetGenMap sets the generation function map.
func (g *BaseGenerator) SetGenMap(genMap map[stmts.StmtType]func(db *sql.DB) (string, error)) {
	g.genMap = genMap
}

// GetLCG returns the underlying LCG for use by derived generators.
func (g *BaseGenerator) GetLCG() *common.LCG {
	return g.lcg
}

// Direction picks a statement type based on weights.
// On the very first call this returns StmtPragma. Afterwards it uses the LCG
// to pick proportionally based on weights.
func (g *BaseGenerator) Direction() stmts.StmtType {
	if g.firstGeneration {
		g.firstGeneration = false
		return stmts.StmtPragma
	}
	
	if g.totalWeight > 0 {
		r := g.lcg.Uint64() % g.totalWeight
		var cum uint64
		for _, t := range stmts.AllStmtTypes {
			w := g.weights[t]
			cum += w
			if r < cum {
				return t
			}
		}
		// fallback
		return stmts.StmtPragma
	}
	
	panic("totalWeight <= 0")
}

// GenerateWithDB produces a SQL statement using the registered generation functions.
// This is a template method that delegates to the genMap.
func (g *BaseGenerator) GenerateWithDB(db *sql.DB) string {
	dir := g.Direction()
	if fn, ok := g.genMap[dir]; ok {
		sqlStr, err := fn(db)
		if err != nil {
			// Error handling could be customized by derived generators
			return "SELECT 1" // fallback
		}
		return sqlStr
	}
	return "SELECT 1" // fallback for unknown directions
}

// TokensUsed returns the number of tokens used by the underlying LCG.
func (g *BaseGenerator) TokensUsed() uint64 {
	return g.lcg.TokensUsed()
}

// recalcTotalWeight recalculates the total weight from all statement types.
func (g *BaseGenerator) recalcTotalWeight() {
	var sum uint64
	for _, t := range stmts.AllStmtTypes {
		sum += g.weights[t]
	}
	g.totalWeight = sum
}
