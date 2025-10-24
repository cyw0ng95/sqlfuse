package stmts

import (
	"fmt"
	"sqlsmith-go/internal/common"
)

// DropTableStmt represents a DROP TABLE statement.
type DropTableStmt struct {
	sql string
}

func (s *DropTableStmt) SQL() string  { return s.sql }
func (s *DropTableStmt) Type() string { return "drop_table" }

// GenDropTable generates a DROP TABLE IF EXISTS statement for a pseudo-random
// table name using the provided LCG.
func GenDropTable(lcg *common.LCG) (Stmt, error) {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}

	// Use same naming scheme as GenCreateTable to sometimes target recently created tables.
	tbl := fmt.Sprintf("tbl_%d", lcg.Uint64()%1000000)
	sql := fmt.Sprintf("DROP TABLE IF EXISTS \"%s\";", tbl)
	return &DropTableStmt{sql: sql}, nil
}
