package turso

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/stmts/stmts"
)

// Generator uses an LCG to drive generation directions and produce SQL snippets.
type Generator struct {
	lcg               *common.LCG
	first             bool // first generation is forced into pragma
	weights           map[stmts.StmtType]uint64
	totalWeight       uint64
	maxRecursionDepth int                // Maximum depth for recursive generation (default: 2)
	flavorConfig      stmts.FlavorConfig // SQL flavor configuration for Turso LibSQL
	genMap            map[stmts.StmtType]func(db *sql.DB) (string, error)
}

// DefaultStmtWeights returns a sensible default weight distribution.
// Values are token-like weights; probabilities are weight / sum(weights).
func DefaultStmtWeights() map[stmts.StmtType]uint64 {
	w := map[stmts.StmtType]uint64{}
	// INSERT variants
	w[stmts.StmtInsert] = 250
	w[stmts.StmtInsertMultiple] = 70
	w[stmts.StmtInsertBulk] = 15
	w[stmts.StmtInsertOrReplace] = 25
	w[stmts.StmtInsertOrIgnore] = 25
	w[stmts.StmtInsertOrAbort] = 15
	w[stmts.StmtInsertOrRollback] = 10
	w[stmts.StmtInsertOrFail] = 10
	// UPDATE and DELETE
	w[stmts.StmtUpdate] = 80
	w[stmts.StmtDelete] = 60
	// Basic SELECT variants
	w[stmts.StmtSelectBasic] = 120
	w[stmts.StmtSelectWhere] = 100
	w[stmts.StmtSelectWhereComplex] = 60
	w[stmts.StmtSelectWhereIn] = 50
	w[stmts.StmtSelectSubquery] = 40
	w[stmts.StmtSelectCase] = 40
	w[stmts.StmtSelectAggregateComplex] = 30
	w[stmts.StmtSelectLike] = 80
	w[stmts.StmtSelectLimit] = 60
	w[stmts.StmtSelectOrder] = 40
	w[stmts.StmtSelectGroup] = 40
	w[stmts.StmtSelectHaving] = 40
	// JOIN variants
	w[stmts.StmtSelectJoin] = 20
	w[stmts.StmtSelectCross] = 20
	w[stmts.StmtSelectInner] = 20
	w[stmts.StmtSelectOuter] = 20
	w[stmts.StmtSelectJoinUsing] = 20
	w[stmts.StmtSelectNatural] = 20
	// Advanced SELECT with recursion/nesting
	w[stmts.StmtSelectRecursive] = 30
	w[stmts.StmtSelectNestedCase] = 25
	w[stmts.StmtSelectComplexJoin] = 25
	// Window functions and CTEs (new)
	w[stmts.StmtSelectWindow] = 35
	w[stmts.StmtSelectMultipleWindows] = 20
	w[stmts.StmtSelectCTE] = 30
	w[stmts.StmtSelectMultipleCTE] = 20
	w[stmts.StmtSelectRecursiveCTE] = 15
	// new: Turso extension functions
	w[stmts.StmtSelectUUID] = 30
	w[stmts.StmtSelectRegexp] = 30
	w[stmts.StmtSelectVector] = 20
	w[stmts.StmtSelectTime] = 35
	// DDL
	w[stmts.StmtCreateTable] = 40
	w[stmts.StmtDropTable] = 40
	w[stmts.StmtAlterTable] = 40
	return w
}

// NewGenerator creates a generator seeded with the provided seed and default weights.
func NewGenerator(seed uint64) *Generator {
	g := &Generator{
		lcg:               common.NewLCG(seed),
		first:             true,
		maxRecursionDepth: 2, // Default recursion depth
		flavorConfig:      nil,
	}
	g.SetWeights(DefaultStmtWeights())
	g.initGenMap()
	return g
}

// SetWeights replaces the current weights and recalculates totals.
func (g *Generator) SetWeights(weights map[stmts.StmtType]uint64) {
	if g.weights == nil {
		g.weights = make(map[stmts.StmtType]uint64, len(weights))
	}
	for k, v := range weights {
		g.weights[k] = v
	}
	g.recalcTotalWeight()
}

// SetWeight sets a single statement type weight and updates totals.
func (g *Generator) SetWeight(t stmts.StmtType, weight uint64) {
	if g.weights == nil {
		g.weights = DefaultStmtWeights()
	}
	g.weights[t] = weight
	g.recalcTotalWeight()
}

// GetWeights returns a copy of the current weights map.
func (g *Generator) GetWeights() map[stmts.StmtType]uint64 {
	out := make(map[stmts.StmtType]uint64, len(g.weights))
	for k, v := range g.weights {
		out[k] = v
	}
	return out
}

// SetMaxRecursionDepth sets the maximum recursion depth for complex SQL generation.
// A depth of 0 means no recursion (simple queries only).
// A depth of 1 allows one level of nesting (e.g., subquery in WHERE).
// A depth of 2 or more allows deeper nesting.
func (g *Generator) SetMaxRecursionDepth(depth int) {
	if depth < 0 {
		depth = 0
	}
	g.maxRecursionDepth = depth
}

// GetMaxRecursionDepth returns the current maximum recursion depth.
func (g *Generator) GetMaxRecursionDepth() int {
	return g.maxRecursionDepth
}

// createGenContext creates a GenContext with Turso-specific flavor configuration.
func (g *Generator) createGenContext(db *sql.DB) *stmts.GenContext {
	return stmts.NewGenContextWithFlavor(db, g.lcg, g.maxRecursionDepth, g.flavorConfig)
}

func (g *Generator) recalcTotalWeight() {
	var sum uint64
	for _, t := range stmts.AllStmtTypes {
		sum += g.weights[t]
	}
	g.totalWeight = sum
}

// initGenMap delegates generator registration to the stmts package, which
// centralizes all statement generators and keeps database-independent logic
// inside the stmts package.
func (g *Generator) initGenMap() {
	built := stmts.BuildGeneratorFuncs(g.lcg, g.maxRecursionDepth, g.flavorConfig)
	m := make(map[stmts.StmtType]func(db *sql.DB) (string, error), len(built))
	for k, fn := range built {
		m[stmts.StmtType(k)] = fn
	}
	g.genMap = m
}

// Direction picks a direction to drive generation.
// On the very first call this is 100% StmtPragma. Afterwards it uses the LCG
// to pick between pragma, ddl, and dml. If weights are set (totalWeight>0)
// selection is proportional to weights.
func (g *Generator) Direction() stmts.StmtType {
	if g.first {
		g.first = false
		return stmts.StmtPragma
	}
	// If weights are configured, pick proportionally.
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
	} else {
		panic("totalWeight < 0")
	}
}

// GenerateWithDB produces a single SQL statement according to the chosen direction.
// If db is provided, can generate SELECTs using schema.
func (g *Generator) GenerateWithDB(db *sql.DB) string {
	dir := g.Direction()
	if fn, ok := g.genMap[dir]; ok {
		sqlStr, err := fn(db)
		if err != nil {
			fmt.Println("Error generating", string(dir)+":", err)
		}
		return sqlStr
	}
	return "SELECT 1" // fallback for unknown directions
}

// TokensUsed returns the number of tokens used by the underlying LCG.
func (g *Generator) TokensUsed() uint64 {
	return g.lcg.TokensUsed()
}
