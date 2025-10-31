package generators

import (
	"database/sql"
	"sqlfuse/internal/generators/dialects"
	"sqlfuse/internal/stmts/stmts"
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
	w[stmts.StmtSelectWhereComplex] = 80 // Increased from 60
	w[stmts.StmtSelectWhereIn] = 70      // Increased from 50
	w[stmts.StmtSelectSubquery] = 60     // Increased from 40
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

	// Advanced SELECT with recursion/nesting - increased for more complexity
	w[stmts.StmtSelectRecursive] = 50    // Increased from 30
	w[stmts.StmtSelectNestedCase] = 40   // Increased from 25
	w[stmts.StmtSelectComplexJoin] = 40  // Increased from 25
	w[stmts.StmtSelectDeeplyNested] = 50 // New: deeply nested subqueries

	// Window functions and CTEs - Higher weights for DuckDB
	// DuckDB has excellent analytical query support
	w[stmts.StmtSelectWindow] = 100         // Increased from 80
	w[stmts.StmtSelectMultipleWindows] = 80 // Increased from 60
	w[stmts.StmtSelectCTE] = 80             // Increased from 60
	w[stmts.StmtSelectMultipleCTE] = 60     // Increased from 45
	w[stmts.StmtSelectRecursiveCTE] = 55    // Increased from 40 - fully supported, high weight

	// Extension functions - DuckDB specific
	// DuckDB has rich built-in functions but may not have all SQLite extensions
	w[stmts.StmtSelectJSON] = 60   // DuckDB has excellent JSON support
	w[stmts.StmtSelectUUID] = 30   // DuckDB has UUID support
	w[stmts.StmtSelectRegexp] = 50 // DuckDB has REGEXP support
	w[stmts.StmtSelectVector] = 0  // Vector functions may differ from Turso
	w[stmts.StmtSelectTime] = 50   // DuckDB has extensive time/date functions

	// DDL
	w[stmts.StmtCreateTable] = 80
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

	// PRAGMA - DuckDB uses SET instead of PRAGMA, so weight is 0
	w[stmts.StmtPragma] = 0

	// DuckDB-specific statements
	w[stmts.StmtCopy] = 60           // Very important for DuckDB data import/export
	w[stmts.StmtSet] = 40            // DuckDB's configuration mechanism
	w[stmts.StmtReset] = 20          // Reset configuration
	w[stmts.StmtCreateSchema] = 30   // Schema management
	w[stmts.StmtDropSchema] = 25     // Schema management
	w[stmts.StmtCreateSequence] = 25 // Sequence support
	w[stmts.StmtDropSequence] = 20   // Sequence support
	w[stmts.StmtCreateMacro] = 35    // DuckDB macros
	w[stmts.StmtDropMacro] = 30      // DuckDB macros
	w[stmts.StmtCreateType] = 20     // Custom types (ENUM)
	w[stmts.StmtDropType] = 15       // Custom types
	w[stmts.StmtDescribe] = 45       // Metadata queries - very useful
	w[stmts.StmtShow] = 45           // Metadata queries - very useful
	w[stmts.StmtSummarize] = 40      // DuckDB's data profiling feature
	w[stmts.StmtUse] = 15            // Schema switching
	w[stmts.StmtCall] = 25           // Call macros/procedures
	w[stmts.StmtCheckpoint] = 15     // Persistence control
	w[stmts.StmtExportDatabase] = 10 // Database export
	w[stmts.StmtImportDatabase] = 10 // Database import
	w[stmts.StmtPrepare] = 20        // Prepared statements
	w[stmts.StmtExecute] = 15        // Execute prepared statements

	// Additional DuckDB-specific statements
	w[stmts.StmtPivot] = 50          // Table reshaping - very useful for analytics
	w[stmts.StmtUnpivot] = 45        // Table reshaping - very useful for analytics
	w[stmts.StmtMergeInto] = 55      // UPSERT operations - important for data merging
	w[stmts.StmtQualify] = 65        // Window function filtering - unique to DuckDB
	w[stmts.StmtAlterDatabase] = 10  // Database operations
	w[stmts.StmtAlterView] = 15      // View operations
	w[stmts.StmtCreateSecret] = 20   // Credential management
	w[stmts.StmtDropSecret] = 15     // Credential management
	w[stmts.StmtLoadInstall] = 25    // Extension management
	w[stmts.StmtCommentOn] = 20      // Documentation
	w[stmts.StmtProfiling] = 30      // Query profiling
	w[stmts.StmtSetVariable] = 35    // User-defined variables

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
