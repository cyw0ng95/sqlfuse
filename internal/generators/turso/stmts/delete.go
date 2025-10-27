package stmts

import (
	"fmt"
	"sqlsmith-go/internal/common"
)

// DeleteGenerator is a StmtGenerator for DELETE statements.
type DeleteGenerator struct{}

// Generate implements StmtGenerator for DELETE statements.
func (g *DeleteGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return genDeleteInternal(ctx.LCG)
}

// CanGenerate implements StmtGenerator. DELETE can always be generated (creates synthetic tables).
func (g *DeleteGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// DeleteStmt represents a DELETE statement.
type DeleteStmt struct {
	sql string
}

func (s *DeleteStmt) SQL() string  { return s.sql }
func (s *DeleteStmt) Type() string { return "delete" }

// GenDelete generates a simple DELETE statement targeting a pseudo-random table.
// It may include a WHERE clause or delete all rows.
// This function is kept for backward compatibility with existing code.
func GenDelete(lcg *common.LCG) (Stmt, error) {
	return genDeleteInternal(lcg)
}

// genDeleteInternal is the internal implementation used by both old and new interfaces.
func genDeleteInternal(lcg *common.LCG) (Stmt, error) {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}

	tbl := fmt.Sprintf("tbl_%d", lcg.Uint64()%1000000)
	where := ""
	if lcg.Intn(2) == 0 {
		col := fmt.Sprintf("col%d", 1+lcg.Intn(6))
		// simple equality condition
		where = fmt.Sprintf(" WHERE \"%s\" = %d", col, lcg.Intn(100))
	}

	sql := fmt.Sprintf("DELETE FROM \"%s\"%s;", tbl, where)
	return &DeleteStmt{sql: sql}, nil
}
