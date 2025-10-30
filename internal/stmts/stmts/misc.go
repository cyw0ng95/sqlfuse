package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
)

// UseGenerator is a StmtGenerator for USE statements.
type UseGenerator struct{}

// Generate implements StmtGenerator for USE statements.
func (g *UseGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenUse(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. USE can always be generated.
func (g *UseGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// UseStmt represents a USE statement.
type UseStmt struct {
	*BaseStmt
}

// GenUse generates a USE statement.
// USE changes the active schema.
// Reference: https://duckdb.org/docs/stable/sql/statements/use
func GenUse(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	// Common schema names
	schemas := []string{"main", "temp", fmt.Sprintf("schema_%d", lcg.Uint64()%100)}
	schema := schemas[lcg.Intn(len(schemas))]
	
	sql := fmt.Sprintf("USE %s;", schema)

	return &UseStmt{
		BaseStmt: NewBaseStmt(sql, "use", flavor),
	}, nil
}

// CheckpointGenerator is a StmtGenerator for CHECKPOINT statements.
type CheckpointGenerator struct{}

// Generate implements StmtGenerator for CHECKPOINT statements.
func (g *CheckpointGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenCheckpoint(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. CHECKPOINT can always be generated.
func (g *CheckpointGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// CheckpointStmt represents a CHECKPOINT statement.
type CheckpointStmt struct {
	*BaseStmt
}

// GenCheckpoint generates a CHECKPOINT statement.
// CHECKPOINT forces a write of the WAL to the database file.
// Reference: https://duckdb.org/docs/stable/sql/statements/checkpoint
func GenCheckpoint(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	// CHECKPOINT can optionally specify a database name
	choice := lcg.Intn(3)
	var sql string
	
	if choice == 0 {
		// Simple CHECKPOINT
		sql = "CHECKPOINT;"
	} else if choice == 1 {
		// FORCE CHECKPOINT
		sql = "FORCE CHECKPOINT;"
	} else {
		// CHECKPOINT for specific database
		dbName := []string{"main", "temp"}[lcg.Intn(2)]
		sql = fmt.Sprintf("CHECKPOINT %s;", dbName)
	}

	return &CheckpointStmt{
		BaseStmt: NewBaseStmt(sql, "checkpoint", flavor),
	}, nil
}

// ExportDatabaseGenerator is a StmtGenerator for EXPORT DATABASE statements.
type ExportDatabaseGenerator struct{}

// Generate implements StmtGenerator for EXPORT DATABASE statements.
func (g *ExportDatabaseGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenExportDatabase(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. EXPORT DATABASE can always be generated.
func (g *ExportDatabaseGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// ExportDatabaseStmt represents an EXPORT DATABASE statement.
type ExportDatabaseStmt struct {
	*BaseStmt
}

// GenExportDatabase generates an EXPORT DATABASE statement.
// Reference: https://duckdb.org/docs/stable/sql/statements/export
func GenExportDatabase(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	exportPath := fmt.Sprintf("/tmp/export_%d", lcg.Uint64()%10000)
	
	var sql string
	choice := lcg.Intn(2)
	
	if choice == 0 {
		// Export to directory
		sql = fmt.Sprintf("EXPORT DATABASE '%s';", exportPath)
	} else {
		// Export with format specification
		sql = fmt.Sprintf("EXPORT DATABASE '%s' (FORMAT PARQUET);", exportPath)
	}

	return &ExportDatabaseStmt{
		BaseStmt: NewBaseStmt(sql, "export_database", flavor),
	}, nil
}

// ImportDatabaseGenerator is a StmtGenerator for IMPORT DATABASE statements.
type ImportDatabaseGenerator struct{}

// Generate implements StmtGenerator for IMPORT DATABASE statements.
func (g *ImportDatabaseGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenImportDatabase(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. IMPORT DATABASE can always be generated.
func (g *ImportDatabaseGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// ImportDatabaseStmt represents an IMPORT DATABASE statement.
type ImportDatabaseStmt struct {
	*BaseStmt
}

// GenImportDatabase generates an IMPORT DATABASE statement.
// Reference: https://duckdb.org/docs/stable/sql/statements/import
func GenImportDatabase(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	importPath := fmt.Sprintf("/tmp/import_%d", lcg.Uint64()%10000)
	sql := fmt.Sprintf("IMPORT DATABASE '%s';", importPath)

	return &ImportDatabaseStmt{
		BaseStmt: NewBaseStmt(sql, "import_database", flavor),
	}, nil
}

// PrepareGenerator is a StmtGenerator for PREPARE statements.
type PrepareGenerator struct{}

// Generate implements StmtGenerator for PREPARE statements.
func (g *PrepareGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenPrepare(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. PREPARE can always be generated.
func (g *PrepareGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// PrepareStmt represents a PREPARE statement.
type PrepareStmt struct {
	*BaseStmt
}

// GenPrepare generates a PREPARE statement.
// Reference: https://duckdb.org/docs/stable/sql/statements/prepare
func GenPrepare(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	stmtName := fmt.Sprintf("stmt_%d", lcg.Uint64()%10000)
	
	var sql string
	choice := lcg.Intn(3)
	
	if choice == 0 {
		// Simple SELECT
		sql = fmt.Sprintf("PREPARE %s AS SELECT $1;", stmtName)
	} else if choice == 1 {
		// INSERT with parameters
		sql = fmt.Sprintf("PREPARE %s AS INSERT INTO temp_table VALUES ($1, $2);", stmtName)
	} else {
		// SELECT with WHERE clause
		sql = fmt.Sprintf("PREPARE %s AS SELECT * FROM temp_table WHERE id = $1;", stmtName)
	}

	return &PrepareStmt{
		BaseStmt: NewBaseStmt(sql, "prepare", flavor),
	}, nil
}

// ExecuteGenerator is a StmtGenerator for EXECUTE statements.
type ExecuteGenerator struct{}

// Generate implements StmtGenerator for EXECUTE statements.
func (g *ExecuteGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenExecute(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. EXECUTE can always be generated.
func (g *ExecuteGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// ExecuteStmt represents an EXECUTE statement.
type ExecuteStmt struct {
	*BaseStmt
}

// GenExecute generates an EXECUTE statement.
// Reference: https://duckdb.org/docs/stable/sql/statements/execute
func GenExecute(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	stmtName := fmt.Sprintf("stmt_%d", lcg.Uint64()%100)
	
	var sql string
	choice := lcg.Intn(3)
	
	if choice == 0 {
		// Execute with one parameter
		sql = fmt.Sprintf("EXECUTE %s(%d);", stmtName, lcg.Intn(1000))
	} else if choice == 1 {
		// Execute with two parameters
		sql = fmt.Sprintf("EXECUTE %s(%d, 'test');", stmtName, lcg.Intn(1000))
	} else {
		// Execute without parameters
		sql = fmt.Sprintf("EXECUTE %s;", stmtName)
	}

	return &ExecuteStmt{
		BaseStmt: NewBaseStmt(sql, "execute", flavor),
	}, nil
}
