package ddl

import (
	"sqlfuse/internal/stmts/stmts"
	"fmt"
	"sqlfuse/internal/common"
)

// AlterTableGenerator is a stmts.StmtGenerator for ALTER TABLE statements.
type AlterTableGenerator struct{}

// Generate implements stmts.StmtGenerator for ALTER TABLE statements.
func (g *AlterTableGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	return genAlterTableInternal(ctx.LCG)
}

// CanGenerate implements stmts.StmtGenerator. ALTER TABLE can always be generated.
func (g *AlterTableGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	return true
}

// AlterTableStmt represents an ALTER TABLE statement.
// It embeds stmts.BaseStmt to avoid boilerplate method implementations.
type AlterTableStmt struct {
	*stmts.BaseStmt
}

// GenAlterTable generates simple ALTER TABLE statements:
// - ADD COLUMN
// - RENAME COLUMN
// - RENAME TABLE
// It targets lightweight pseudo-random table and column names produced by the LCG.
// This function is kept for backward compatibility with existing code.
func GenAlterTable(lcg *common.LCG) (stmts.Stmt, error) {
	return genAlterTableInternal(lcg)
}

// genAlterTableInternal is the internal implementation used by both old and new interfaces.
func genAlterTableInternal(lcg *common.LCG) (stmts.Stmt, error) {
	lcg = stmts.EnsureLCG(lcg)
	flavor := stmts.GetDefaultFlavor()

	op := lcg.Intn(3)
	tbl := fmt.Sprintf("tbl_%d", lcg.Uint64()%1000000)
	types := []string{"INTEGER", "TEXT", "REAL", "BLOB"}

	var sql string
	switch op {
	case 0: // ADD COLUMN
		col := fmt.Sprintf("col%d", 1+lcg.Intn(100))
		t := types[lcg.Intn(len(types))]
		sql = fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN \"%s\" %s;", tbl, col, t)
	case 1: // RENAME COLUMN
		old := fmt.Sprintf("col%d", 1+lcg.Intn(100))
		new := fmt.Sprintf("col%d", 101+lcg.Intn(100))
		sql = fmt.Sprintf("ALTER TABLE \"%s\" RENAME COLUMN \"%s\" TO \"%s\";", tbl, old, new)
	default: // RENAME TABLE
		newTbl := fmt.Sprintf("tbl_%d", lcg.Uint64()%1000000)
		sql = fmt.Sprintf("ALTER TABLE \"%s\" RENAME TO \"%s\";", tbl, newTbl)
	}

	return &AlterTableStmt{
		BaseStmt: stmts.NewBaseStmt(sql, "alter_table", flavor),
	}, nil
}
