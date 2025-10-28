package generators

import (
	"database/sql"
	"sqlsmith-go/internal/generators/dialects"
	"sqlsmith-go/internal/stmts/stmts"
)

// GoSQLite3Generator is the go-sqlite3-specific generator that embeds BaseGenerator
// for common functionality and adds go-sqlite3-specific configuration.
// go-sqlite3 supports full SQLite3 features, unlike Turso LibSQL which has restrictions.
type GoSQLite3Generator struct {
	*BaseGenerator
	flavorConfig stmts.FlavorConfig // SQL flavor configuration for go-sqlite3
}

// DefaultGoSQLite3StmtWeights returns a weight distribution for go-sqlite3.
// This includes higher weights for advanced features that Turso doesn't support:
// - Window functions (OVER clause)
// - Recursive CTEs (WITH RECURSIVE)
// - EXISTS and IN subqueries
// - FILTER clauses for aggregates
// - REGEXP and MATCH operators
//
// Values are token-like weights; probabilities are weight / sum(weights).
func DefaultGoSQLite3StmtWeights() map[stmts.StmtType]uint64 {
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
	
	// Window functions and CTEs - Higher weights for go-sqlite3 since it supports them fully
	// Unlike Turso which doesn't support these, go-sqlite3 does, so we emphasize them
	w[stmts.StmtSelectWindow] = 60          // Increased from 35 in Turso
	w[stmts.StmtSelectMultipleWindows] = 40 // Increased from 20 in Turso
	w[stmts.StmtSelectCTE] = 50             // Increased from 30 in Turso
	w[stmts.StmtSelectMultipleCTE] = 35     // Increased from 20 in Turso
	w[stmts.StmtSelectRecursiveCTE] = 30    // Increased from 15 in Turso - fully supported!
	
	// Extension functions - go-sqlite3 specific
	// Note: UUID, Vector functions may not be available by default in go-sqlite3
	// So we use lower weights than Turso (which has these built-in)
	w[stmts.StmtSelectUUID] = 10   // Lower than Turso's 30
	w[stmts.StmtSelectRegexp] = 40 // Higher than Turso's 30 - REGEXP is standard SQLite
	w[stmts.StmtSelectVector] = 5  // Lower than Turso's 20 - not standard in go-sqlite3
	w[stmts.StmtSelectTime] = 35   // Same as Turso
	
	// DDL
	w[stmts.StmtCreateTable] = 40
	w[stmts.StmtDropTable] = 40
	w[stmts.StmtAlterTable] = 40
	
	// PRAGMA - go-sqlite3 supports all SQLite3 pragmas
	w[stmts.StmtPragma] = 50
	
	return w
}

// NewGoSQLite3Generator creates a generator seeded with the provided seed and default weights.
func NewGoSQLite3Generator(seed uint64) *GoSQLite3Generator {
	g := &GoSQLite3Generator{
		BaseGenerator: NewBaseGenerator(seed),
		flavorConfig:  dialects.NewGoSQLite3FlavorConfig(),
	}
	g.SetWeights(DefaultGoSQLite3StmtWeights())
	g.initGenMap()
	return g
}

// initGenMap delegates generator registration to the stmts package, which
// centralizes all statement generators and keeps database-independent logic
// inside the stmts package.
func (g *GoSQLite3Generator) initGenMap() {
	built := stmts.BuildGeneratorFuncs(g.GetLCG(), g.GetMaxRecursionDepth(), g.flavorConfig)
	m := make(map[stmts.StmtType]func(db *sql.DB) (string, error), len(built))
	for k, fn := range built {
		m[stmts.StmtType(k)] = fn
	}
	g.SetGenMap(m)
}

// Name returns the generator's identifier.
func (g *GoSQLite3Generator) Name() string {
	return "go-sqlite3"
}

// SupportedStmts returns a map from stmt type name to its default weight.
func (g *GoSQLite3Generator) SupportedStmts() map[string]uint64 {
	w := DefaultGoSQLite3StmtWeights()
	out := make(map[string]uint64, len(w))
	for k, v := range w {
		out[string(k)] = v
	}
	return out
}
