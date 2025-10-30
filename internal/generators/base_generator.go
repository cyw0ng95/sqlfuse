package generators

import (
	"database/sql"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/stmts/stmts"
)

// BaseGenerator provides common generator functionality that can be embedded
// by specific generator implementations. It handles weights, LCG state,
// recursion depth, and direction selection using the Strategy pattern.
// Enhanced with impedance matching and statistics tracking inspired by
// the original SQLsmith (https://github.com/anse1/sqlsmith).
type BaseGenerator struct {
	lcg               *common.LCG
	weights           map[stmts.StmtType]uint64
	totalWeight       uint64
	maxRecursionDepth int
	genMap            map[stmts.StmtType]func(db *sql.DB) (string, error)

	// Impedance matching and statistics
	impedanceMatcher *common.ImpedanceMatcher
	stats            *common.GenerationStats
	enableImpedance  bool
	enableStats      bool
}

// NewBaseGenerator creates a new base generator with the given seed.
func NewBaseGenerator(seed uint64) *BaseGenerator {
	return &BaseGenerator{
		lcg:               common.NewLCG(seed),
		maxRecursionDepth: 4, // Default recursion depth - increased for more complex SQL
		impedanceMatcher:  common.NewImpedanceMatcher(),
		stats:             common.NewGenerationStats(),
		enableImpedance:   false, // Disabled by default for backward compatibility
		enableStats:       false, // Disabled by default for backward compatibility
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
// When impedance matching is enabled, blacklisted statement types are skipped.
func (g *BaseGenerator) Direction() stmts.StmtType {
	if g.totalWeight > 0 {
		// Try up to 100 times to find a non-blacklisted statement type
		const maxAttempts = 100
		for attempt := 0; attempt < maxAttempts; attempt++ {
			r := g.lcg.Uint64() % g.totalWeight
			var cum uint64
			for _, t := range stmts.AllStmtTypes {
				w := g.weights[t]
				cum += w
				if r < cum {
					// Check if this type is blacklisted
					if g.IsBlacklisted(t) {
						break // Try again with a new random number
					}
					return t
				}
			}
		}

		// If we couldn't find a non-blacklisted type after 100 attempts,
		// fall back to selecting the first non-blacklisted type
		for _, t := range stmts.AllStmtTypes {
			if g.weights[t] > 0 && !g.IsBlacklisted(t) {
				return t
			}
		}

		// Ultimate fallback if everything is blacklisted
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

// EnableImpedanceMatching enables or disables impedance matching.
// When enabled, statement types with high error rates are blacklisted.
func (g *BaseGenerator) EnableImpedanceMatching(enabled bool) {
	g.enableImpedance = enabled
	if g.impedanceMatcher != nil {
		g.impedanceMatcher.SetEnabled(enabled)
	}
}

// EnableStatistics enables or disables statistics tracking.
func (g *BaseGenerator) EnableStatistics(enabled bool) {
	g.enableStats = enabled
}

// GetImpedanceMatcher returns the impedance matcher for inspection.
func (g *BaseGenerator) GetImpedanceMatcher() *common.ImpedanceMatcher {
	return g.impedanceMatcher
}

// GetStatistics returns the statistics tracker for inspection.
func (g *BaseGenerator) GetStatistics() *common.GenerationStats {
	return g.stats
}

// RecordSuccess records a successful statement generation/execution.
func (g *BaseGenerator) RecordSuccess(stmtType stmts.StmtType) {
	if g.enableImpedance && g.impedanceMatcher != nil {
		g.impedanceMatcher.RecordSuccess(string(stmtType))
	}
}

// RecordFailure records a failed statement generation/execution.
func (g *BaseGenerator) RecordFailure(stmtType stmts.StmtType) {
	if g.enableImpedance && g.impedanceMatcher != nil {
		g.impedanceMatcher.RecordFailure(string(stmtType))
	}
}

// IsBlacklisted checks if a statement type is blacklisted due to high error rate.
func (g *BaseGenerator) IsBlacklisted(stmtType stmts.StmtType) bool {
	if !g.enableImpedance || g.impedanceMatcher == nil {
		return false
	}
	return g.impedanceMatcher.IsBlacklisted(string(stmtType))
}
