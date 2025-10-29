package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/stmts/helper"
)

// GenSelectGroupBy generates a SELECT with GROUP BY aggregation.
func GenSelectGroupBy(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db)
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

	// pick a group by column and an aggregate
	grp := tbl.Cols[rnd(len(tbl.Cols))]
	sql := fmt.Sprintf("SELECT %s, COUNT(1) FROM %s GROUP BY %s;", quoteIdent(grp.Name), quoteIdent(tbl.Name), quoteIdent(grp.Name))
	return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}
