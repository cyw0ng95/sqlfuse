package stmts

import (
	"fmt"
	"sqlsmith-go/internal/common"
)

// AlterTableGenerator is a StmtGenerator for ALTER TABLE statements.
type AlterTableGenerator struct{}

// Generate implements StmtGenerator for ALTER TABLE statements.
func (g *AlterTableGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return genAlterTableInternal(ctx.LCG)
}

// CanGenerate implements StmtGenerator. ALTER TABLE can always be generated.
func (g *AlterTableGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// AlterTableStmt represents an ALTER TABLE statement.
type AlterTableStmt struct {
	sql string
}

func (s *AlterTableStmt) SQL() string  { return s.sql }
func (s *AlterTableStmt) Type() string { return "alter_table" }

// GenAlterTable generates simple ALTER TABLE statements:
// - ADD COLUMN
// - RENAME COLUMN
// - RENAME TABLE
// It targets lightweight pseudo-random table and column names produced by the LCG.
// This function is kept for backward compatibility with existing code.
func GenAlterTable(lcg *common.LCG) (Stmt, error) {
	return genAlterTableInternal(lcg)
}

// genAlterTableInternal is the internal implementation used by both old and new interfaces.
func genAlterTableInternal(lcg *common.LCG) (Stmt, error) {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}

	op := lcg.Intn(3)
	tbl := fmt.Sprintf("tbl_%d", lcg.Uint64()%1000000)
	types := []string{"INTEGER", "TEXT", "REAL", "BLOB"}

	switch op {
	case 0: // ADD COLUMN
		col := fmt.Sprintf("col%d", 1+lcg.Intn(100))
		t := types[lcg.Intn(len(types))]
		sql := fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN \"%s\" %s;", tbl, col, t)
		return &AlterTableStmt{sql: sql}, nil
	case 1: // RENAME COLUMN
		old := fmt.Sprintf("col%d", 1+lcg.Intn(100))
		new := fmt.Sprintf("col%d", 101+lcg.Intn(100))
		sql := fmt.Sprintf("ALTER TABLE \"%s\" RENAME COLUMN \"%s\" TO \"%s\";", tbl, old, new)
		return &AlterTableStmt{sql: sql}, nil
	default: // RENAME TABLE
		newTbl := fmt.Sprintf("tbl_%d", lcg.Uint64()%1000000)
		sql := fmt.Sprintf("ALTER TABLE \"%s\" RENAME TO \"%s\";", tbl, newTbl)
		return &AlterTableStmt{sql: sql}, nil
	}
}
