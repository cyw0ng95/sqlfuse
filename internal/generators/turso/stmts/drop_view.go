package stmts

import (
	"fmt"
	"sqlsmith-go/internal/common"
)

// DropViewStmt represents a DROP VIEW statement.
type DropViewStmt struct {
	sql string
}

func (s *DropViewStmt) SQL() string  { return s.sql }
func (s *DropViewStmt) Type() string { return "drop_view" }

// GenDropView generates a DROP VIEW IF EXISTS statement for a pseudo-random view name.
func GenDropView(lcg *common.LCG) (Stmt, error) {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}

	view := fmt.Sprintf("view_%d", lcg.Uint64()%1000000)
	sql := fmt.Sprintf("DROP VIEW IF EXISTS \"%s\";", view)
	return &DropViewStmt{sql: sql}, nil
}
