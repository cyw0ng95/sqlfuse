package other

import (
	"sqlfuse/internal/stmts/stmts"
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
	"sqlfuse/internal/stmts/helper"
)

// AnalyzeGenerator is a stmts.StmtGenerator for ANALYZE statements.
type AnalyzeGenerator struct{}

// Generate implements stmts.StmtGenerator for ANALYZE statements.
func (g *AnalyzeGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	return GenAnalyzeWithFlavor(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements stmts.StmtGenerator. ANALYZE can always be generated.
func (g *AnalyzeGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	return true
}

// AnalyzeStmt represents an ANALYZE statement.
// It embeds stmts.BaseStmt to avoid boilerplate method implementations.
type AnalyzeStmt struct {
	*stmts.BaseStmt
}

// GenAnalyzeWithFlavor generates an ANALYZE statement with flavor support.
// ANALYZE gathers statistics about tables and indexes to help the query optimizer.
// According to SQLite documentation: https://sqlite.org/lang_analyze.html
// Turso COMPAT.md: Yes (full support).
func GenAnalyzeWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (stmts.Stmt, error) {
	lcg = stmts.EnsureLCG(lcg)
	if flavor == nil {
		flavor = stmts.GetDefaultFlavor()
	}

	// ANALYZE can be used in several ways:
	// 1. ANALYZE; (analyzes all tables)
	// 2. ANALYZE schema_name; (analyzes all tables in a schema)
	// 3. ANALYZE table_name; (analyzes a specific table)
	// 4. ANALYZE schema_name.table_name; (analyzes a specific table in a schema)

	choice := lcg.Intn(10)
	var sql string

	if choice < 3 || db == nil {
		// 30% chance: ANALYZE all tables, or fallback if no db
		sql = "ANALYZE;"
	} else {
		// 70% chance: ANALYZE a specific table
		tables, err := helper.GetAllTablesAndColsWithFlavor(db, flavor)
		if err != nil || len(tables) == 0 {
			// Fallback to analyzing everything
			sql = "ANALYZE;"
		} else {
			// Pick a random table
			tbl := tables[lcg.Intn(len(tables))]
			sql = fmt.Sprintf("ANALYZE \"%s\";", tbl.Name)
		}
	}

	return &AnalyzeStmt{
		BaseStmt: stmts.NewBaseStmt(sql, "analyze", flavor),
	}, nil
}
