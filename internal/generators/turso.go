package generators

import (
	"database/sql"
	"sqlsmith-go/internal/generators/dialects"
	"sqlsmith-go/internal/stmts/stmts"
)

// TursoGenerator is the Turso-specific generator that embeds BaseGenerator
// for common functionality and adds Turso-specific configuration.
type TursoGenerator struct {
	*BaseGenerator
	flavorConfig stmts.FlavorConfig // SQL flavor configuration for Turso LibSQL
}

// DefaultTursoStmtWeights returns a sensible default weight distribution.
// Values are token-like weights; probabilities are weight / sum(weights).
func DefaultTursoStmtWeights() map[stmts.StmtType]uint64 {
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
	
	// PRAGMA - Turso supports a limited set of pragmas
	w[stmts.StmtPragma] = 30
	
	return w
}

// NewTursoGenerator creates a generator seeded with the provided seed and default weights.
func NewTursoGenerator(seed uint64) *TursoGenerator {
	g := &TursoGenerator{
		BaseGenerator: NewBaseGenerator(seed),
		flavorConfig:  dialects.NewTursoFlavorConfig(),
	}
	g.SetWeights(DefaultTursoStmtWeights())
	g.initGenMap()
	return g
}

// initGenMap delegates generator registration to the stmts package, which
// centralizes all statement generators and keeps database-independent logic
// inside the stmts package.
func (g *TursoGenerator) initGenMap() {
	built := stmts.BuildGeneratorFuncs(g.GetLCG(), g.GetMaxRecursionDepth(), g.flavorConfig)
	m := make(map[stmts.StmtType]func(db *sql.DB) (string, error), len(built))
	for k, fn := range built {
		m[stmts.StmtType(k)] = fn
	}
	g.SetGenMap(m)
}

// Name returns the generator's identifier.
func (g *TursoGenerator) Name() string {
	return "turso"
}

// SupportedStmts returns a map from stmt type name to its default weight.
func (g *TursoGenerator) SupportedStmts() map[string]uint64 {
	w := DefaultTursoStmtWeights()
	out := make(map[string]uint64, len(w))
	for k, v := range w {
		out[string(k)] = v
	}
	return out
}
