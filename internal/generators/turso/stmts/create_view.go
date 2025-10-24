package stmts

import (
	"fmt"
	"sqlsmith-go/internal/common"
)

// CreateViewStmt represents a CREATE VIEW statement.
type CreateViewStmt struct {
	sql string
}

func (s *CreateViewStmt) SQL() string  { return s.sql }
func (s *CreateViewStmt) Type() string { return "create_view" }

// GenCreateView generates a simple CREATE VIEW statement that selects a constant.
// Using a constant SELECT avoids depending on existing tables.
func GenCreateView(lcg *common.LCG) (Stmt, error) {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}

	view := fmt.Sprintf("view_%d", lcg.Uint64()%1000000)
	sql := fmt.Sprintf("CREATE VIEW IF NOT EXISTS \"%s\" AS SELECT 1 AS col1;", view)
	return &CreateViewStmt{sql: sql}, nil
}
