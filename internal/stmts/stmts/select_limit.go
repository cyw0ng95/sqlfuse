package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/stmts/helper"
)

// GenSelectLimit generates a SELECT with a LIMIT clause (different limit ranges).
func GenSelectLimit(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db, "")
	if err != nil || len(tbls) == 0 {
		return SelectStmt{sql: "SELECT 1;", flavor: GetDefaultFlavor()}, nil
	}

	var rnd func(int) int
	if lcg != nil {
		rnd = lcg.Intn
	} else {
		rnd = func(n int) int { return 0 }
	}

	tbl := tbls[rnd(len(tbls))]
	if len(tbl.Cols) == 0 {
		return SelectStmt{sql: fmt.Sprintf("SELECT * FROM %s;", quoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
	}

	// pick columns
	cols := []helper.ColumnInfo{}
	for i := 0; i < 1+rnd(min(3, len(tbl.Cols))); i++ {
		cols = append(cols, tbl.Cols[rnd(len(tbl.Cols))])
	}

	limit := 1 + rnd(1000)
	sql := fmt.Sprintf("SELECT %s FROM %s LIMIT %d;", joinCols(cols), quoteIdent(tbl.Name), limit)
	return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}
