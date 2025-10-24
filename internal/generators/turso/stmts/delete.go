package stmts

import (
	"fmt"
	"sqlsmith-go/internal/common"
)

// DeleteStmt represents a DELETE statement.
type DeleteStmt struct {
	sql string
}

func (s *DeleteStmt) SQL() string  { return s.sql }
func (s *DeleteStmt) Type() string { return "delete" }

// GenDelete generates a simple DELETE statement targeting a pseudo-random table.
// It may include a WHERE clause or delete all rows.
func GenDelete(lcg *common.LCG) (Stmt, error) {
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
