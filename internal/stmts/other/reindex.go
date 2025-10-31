package other

import (
	"sqlfuse/internal/stmts/stmts"
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
	"sqlfuse/internal/stmts/helper"
)

// ReindexGenerator is a stmts.StmtGenerator for REINDEX statements.
type ReindexGenerator struct{}

// Generate implements stmts.StmtGenerator for REINDEX statements.
func (g *ReindexGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	return GenReindex(ctx.DB, ctx.LCG)
}

// CanGenerate implements stmts.StmtGenerator. REINDEX can always be generated.
func (g *ReindexGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	return true
}

// ReindexStmt represents a REINDEX statement.
// It embeds stmts.BaseStmt to avoid boilerplate method implementations.
type ReindexStmt struct {
	*stmts.BaseStmt
}

// GenReindex generates a REINDEX statement.
// REINDEX rebuilds indexes to correct corruption or improve performance.
// According to SQLite documentation: https://sqlite.org/lang_reindex.html
// Turso COMPAT.md: Yes (full support).
func GenReindex(db *sql.DB, lcg *common.LCG) (stmts.Stmt, error) {
	lcg = stmts.EnsureLCG(lcg)

	// REINDEX can be used in several ways:
	// 1. REINDEX; (reindex all indexes)
	// 2. REINDEX collation_name; (reindex all indexes using a collation)
	// 3. REINDEX table_or_index_name; (reindex a specific table or index)
	// 4. REINDEX schema_name.table_or_index_name;

	choice := lcg.Intn(10)
	var sql string

	if choice < 2 || db == nil {
		// 20% chance: REINDEX all, or fallback if no db
		sql = "REINDEX;"
	} else if choice < 4 {
		// 20% chance: REINDEX with collation
		collations := []string{"BINARY", "NOCASE", "RTRIM"}
		collation := collations[lcg.Intn(len(collations))]
		sql = fmt.Sprintf("REINDEX %s;", collation)
	} else {
		// 60% chance: REINDEX a specific table
		tables, err := helper.GetAllTablesAndCols(db, "sqlite")
		if err != nil || len(tables) == 0 {
			// Fallback to reindexing everything
			sql = "REINDEX;"
		} else {
			// Pick a random table
			tbl := tables[lcg.Intn(len(tables))]
			sql = fmt.Sprintf("REINDEX \"%s\";", tbl.Name)
		}
	}

	return &ReindexStmt{
		BaseStmt: stmts.NewBaseStmt(sql, "reindex", stmts.GetDefaultFlavor()),
	}, nil
}
