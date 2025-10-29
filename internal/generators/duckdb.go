package generators

import (
	"database/sql"
	"sqlsmith-go/internal/generators/dialects"
	"sqlsmith-go/internal/stmts/stmts"
)

// DuckDBGenerator is the DuckDB-specific generator that embeds BaseGenerator
// for common functionality and adds DuckDB-specific configuration.
// DuckDB is an analytical database with extensive SQL support beyond SQLite.
type DuckDBGenerator struct {
	*BaseGenerator
	flavorConfig stmts.FlavorConfig // SQL flavor configuration for DuckDB
}

// DefaultDuckDBStmtWeights returns a weight distribution for DuckDB.
// DuckDB is an analytical database, so we emphasize:
// - SELECT queries over DML (analytical workload)
// - Window functions and CTEs (analytical features)
// - Aggregations and complex queries
// - Lower weights for transactions (less common in analytical workloads)
//
// Values are token-like weights; probabilities are weight / sum(weights).
func DefaultDuckDBStmtWeights() map[stmts.StmtType]uint64 {
	w := map[stmts.StmtType]uint64{}
	
	// INSERT variants - Lower than OLTP databases since DuckDB is analytical
	w[stmts.StmtInsert] = 150
	w[stmts.StmtInsertMultiple] = 50
	w[stmts.StmtInsertBulk] = 30
	w[stmts.StmtInsertOrReplace] = 15
	w[stmts.StmtInsertOrIgnore] = 15
	w[stmts.StmtInsertOrAbort] = 10
	w[stmts.StmtInsertOrRollback] = 5
	w[stmts.StmtInsertOrFail] = 5
	
	// UPDATE and DELETE - Lower for analytical workloads
	w[stmts.StmtUpdate] = 40
	w[stmts.StmtDelete] = 30
	
	// Basic SELECT variants - Higher weights for analytical queries
	w[stmts.StmtSelectBasic] = 150
	w[stmts.StmtSelectWhere] = 120
	w[stmts.StmtSelectWhereComplex] = 80
	w[stmts.StmtSelectWhereIn] = 70
	w[stmts.StmtSelectSubquery] = 60
	w[stmts.StmtSelectCase] = 60
	w[stmts.StmtSelectAggregateComplex] = 50
	w[stmts.StmtSelectLike] = 40
	w[stmts.StmtSelectLimit] = 80
	w[stmts.StmtSelectOrder] = 60
	w[stmts.StmtSelectGroup] = 70
	w[stmts.StmtSelectHaving] = 60
	
	// JOIN variants - Important for analytical queries
	w[stmts.StmtSelectJoin] = 40
	w[stmts.StmtSelectCross] = 30
	w[stmts.StmtSelectInner] = 40
	w[stmts.StmtSelectOuter] = 40
	w[stmts.StmtSelectJoinUsing] = 30
	w[stmts.StmtSelectNatural] = 25
	
	// Advanced SELECT with recursion/nesting
	w[stmts.StmtSelectRecursive] = 40
	w[stmts.StmtSelectNestedCase] = 35
	w[stmts.StmtSelectComplexJoin] = 35
	
	// Window functions and CTEs - Much higher for DuckDB (analytical database)
	w[stmts.StmtSelectWindow] = 80            // Higher than go-sqlite3's 60
	w[stmts.StmtSelectMultipleWindows] = 50   // Higher than go-sqlite3's 40
	w[stmts.StmtSelectCTE] = 70               // Higher than go-sqlite3's 50
	w[stmts.StmtSelectMultipleCTE] = 45       // Higher than go-sqlite3's 35
	w[stmts.StmtSelectRecursiveCTE] = 40      // Higher than go-sqlite3's 30
	
	// Extension functions - DuckDB has rich built-in functions
	w[stmts.StmtSelectJSON] = 60   // JSON support is built-in
	w[stmts.StmtSelectUUID] = 40   // UUID support
	w[stmts.StmtSelectRegexp] = 50 // REGEXP support
	w[stmts.StmtSelectVector] = 30 // Some vector operations
	w[stmts.StmtSelectTime] = 50   // Rich date/time functions
	
	// DDL - Normal weights
	w[stmts.StmtCreateTable] = 40
	w[stmts.StmtDropTable] = 40
	w[stmts.StmtAlterTable] = 35
	w[stmts.StmtCreateView] = 30
	w[stmts.StmtDropView] = 30
	w[stmts.StmtCreateIndex] = 35
	w[stmts.StmtDropIndex] = 35
	w[stmts.StmtCreateVirtualTable] = 10 // DuckDB handles this differently
	w[stmts.StmtCreateTrigger] = 20
	w[stmts.StmtDropTrigger] = 20
	
	// Transaction control - Lower for analytical workloads
	w[stmts.StmtBegin] = 15
	w[stmts.StmtCommit] = 15
	w[stmts.StmtRollback] = 15
	w[stmts.StmtSavepoint] = 10
	w[stmts.StmtRelease] = 10
	
	// Database attachment - DuckDB supports this
	w[stmts.StmtAttach] = 15
	w[stmts.StmtDetach] = 15
	
	// Query analysis - Useful for analytical database
	w[stmts.StmtExplain] = 25
	w[stmts.StmtExplainQueryPlan] = 25
	
	// Database maintenance
	w[stmts.StmtAnalyze] = 20
	w[stmts.StmtVacuum] = 10
	w[stmts.StmtReindex] = 15
	
	// Compound SELECT statements - Important for analytical queries
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
