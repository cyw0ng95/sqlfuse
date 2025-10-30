package generators

import (
	"database/sql"
	"sqlsmith-go/internal/generators/dialects"
	"sqlsmith-go/internal/stmts/stmts"
)

// DuckDBGenerator is the DuckDB-specific generator that embeds BaseGenerator
// for common functionality and adds DuckDB-specific configuration.
// DuckDB is an in-process SQL OLAP database with extensive analytical capabilities.
type DuckDBGenerator struct {
	*BaseGenerator
	flavorConfig stmts.FlavorConfig // SQL flavor configuration for DuckDB
}

// DefaultDuckDBStmtWeights returns a weight distribution for DuckDB.
// DuckDB supports advanced analytical features, so we emphasize:
// - Window functions (OVER clause)
// - Recursive CTEs (WITH RECURSIVE)
// - EXISTS and IN subqueries
// - FILTER clauses for aggregates
// - Complex analytical queries
//
// Note: DuckDB uses SET for configuration instead of PRAGMA,
// so PRAGMA weight is set to 0.
//
// Values are token-like weights; probabilities are weight / sum(weights).
func DefaultDuckDBStmtWeights() map[stmts.StmtType]uint64 {
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
	
	// Window functions and CTEs - Higher weights for DuckDB
	// DuckDB has excellent analytical query support
	w[stmts.StmtSelectWindow] = 80          // Significantly higher than Turso
	w[stmts.StmtSelectMultipleWindows] = 60 // Higher than other flavors
	w[stmts.StmtSelectCTE] = 60             // Higher than other flavors
	w[stmts.StmtSelectMultipleCTE] = 45     // Higher than other flavors
	w[stmts.StmtSelectRecursiveCTE] = 40    // Fully supported, high weight
	
	// Extension functions - DuckDB specific
	// DuckDB has rich built-in functions but may not have all SQLite extensions
	w[stmts.StmtSelectJSON] = 60   // DuckDB has excellent JSON support
	w[stmts.StmtSelectUUID] = 30   // DuckDB has UUID support
	w[stmts.StmtSelectRegexp] = 50 // DuckDB has REGEXP support
	w[stmts.StmtSelectVector] = 0  // Vector functions may differ from Turso
	w[stmts.StmtSelectTime] = 50   // DuckDB has extensive time/date functions
	
	// DDL
	w[stmts.StmtCreateTable] = 40
	w[stmts.StmtDropTable] = 40
	w[stmts.StmtAlterTable] = 40
	w[stmts.StmtCreateView] = 35
	w[stmts.StmtDropView] = 35
	w[stmts.StmtCreateIndex] = 35
	w[stmts.StmtDropIndex] = 35
	w[stmts.StmtCreateVirtualTable] = 10 // Lower weight, different from SQLite
	w[stmts.StmtCreateTrigger] = 20
	w[stmts.StmtDropTrigger] = 20
	
	// Transaction control
	w[stmts.StmtBegin] = 25
	w[stmts.StmtCommit] = 25
	w[stmts.StmtRollback] = 25
	w[stmts.StmtSavepoint] = 15
	w[stmts.StmtRelease] = 15
	
	// Database attachment (different from SQLite)
	w[stmts.StmtAttach] = 5
	w[stmts.StmtDetach] = 5
	
	// Query analysis
	w[stmts.StmtExplain] = 25
	w[stmts.StmtExplainQueryPlan] = 25
	
	// Database maintenance
	w[stmts.StmtAnalyze] = 15
	w[stmts.StmtVacuum] = 10
	w[stmts.StmtReindex] = 15
	
	// Compound SELECT statements - DuckDB handles these well
	w[stmts.StmtSelectUnion] = 50
	w[stmts.StmtSelectIntersect] = 40
	w[stmts.StmtSelectExcept] = 40
	
	// PRAGMA - DuckDB has different pragma support than SQLite
	w[stmts.StmtPragma] = 30
	
	return w
}

// NewDuckDBGenerator creates a generator seeded with the provided seed and default weights.
func NewDuckDBGenerator(seed uint64) *DuckDBGenerator {
	g := &DuckDBGenerator{
		BaseGenerator: NewBaseGenerator(seed),
		flavorConfig:  dialects.NewDuckDBFlavorConfig(),
	}
	g.SetWeights(DefaultDuckDBStmtWeights())
	g.initGenMap()
	return g
}

// initGenMap delegates generator registration to the stmts package, which
// centralizes all statement generators and keeps database-independent logic
// inside the stmts package.
func (g *DuckDBGenerator) initGenMap() {
	built := stmts.BuildGeneratorFuncs(g.GetLCG(), g.GetMaxRecursionDepth(), g.flavorConfig)
	m := make(map[stmts.StmtType]func(db *sql.DB) (string, error), len(built))
	for k, fn := range built {
		m[stmts.StmtType(k)] = fn
	}
	g.SetGenMap(m)
}

// Name returns the generator's identifier.
func (g *DuckDBGenerator) Name() string {
	return "duckdb"
}

// SupportedStmts returns a map from stmt type name to its default weight.
func (g *DuckDBGenerator) SupportedStmts() map[string]uint64 {
	w := DefaultDuckDBStmtWeights()
	out := make(map[string]uint64, len(w))
	for k, v := range w {
		out[string(k)] = v
	}
	return out
}
