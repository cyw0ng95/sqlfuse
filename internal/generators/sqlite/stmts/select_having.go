package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/generators/sqlite/helper"
)

// GenSelectHaving generates a GROUP BY with HAVING predicate.
func GenSelectHaving(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tbls) == 0 {
		return SelectStmt{sql: "SELECT 1;"}, nil
	}

	var rnd func(int) int
	if lcg != nil {
		rnd = lcg.Intn
	} else {
		rnd = func(n int) int { return 0 }
	}

	tbl := tbls[rnd(len(tbls))]
	if len(tbl.Cols) == 0 {
		return SelectStmt{sql: fmt.Sprintf("SELECT * FROM %s;", quoteIdent(tbl.Name))}, nil
	}

	grp := tbl.Cols[rnd(len(tbl.Cols))]
	sql := fmt.Sprintf("SELECT %s, COUNT(1) as cnt FROM %s GROUP BY %s HAVING cnt > %d;", quoteIdent(grp.Name), quoteIdent(tbl.Name), quoteIdent(grp.Name), 1+rnd(10))
	return SelectStmt{sql: sql}, nil
}
