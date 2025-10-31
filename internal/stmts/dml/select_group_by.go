package dml

import (
	"sqlfuse/internal/stmts/stmts"
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
	"sqlfuse/internal/stmts/helper"
)

// GenSelectGroupBy generates a SELECT with GROUP BY aggregation.
func GenSelectGroupBy(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db, "sqlite")
	if err != nil || len(tbls) == 0 {
		return SelectStmt{sql: "SELECT 1;", flavor: stmts.GetDefaultFlavor()}, nil
	}

	var rnd func(int) int
	if lcg != nil {
		rnd = lcg.Intn
	} else {
		rnd = func(n int) int { return 0 }
	}

	tbl := tbls[rnd(len(tbls))]
	if len(tbl.Cols) == 0 {
		return SelectStmt{sql: fmt.Sprintf("SELECT * FROM %s;", stmts.QuoteIdent(tbl.Name)), flavor: stmts.GetDefaultFlavor()}, nil
	}

	// pick a group by column and an aggregate
	grp := tbl.Cols[rnd(len(tbl.Cols))]
	sql := fmt.Sprintf("SELECT %s, COUNT(1) FROM %s GROUP BY %s;", stmts.QuoteIdent(grp.Name), stmts.QuoteIdent(tbl.Name), stmts.QuoteIdent(grp.Name))
	return SelectStmt{sql: sql, flavor: stmts.GetDefaultFlavor()}, nil
}
