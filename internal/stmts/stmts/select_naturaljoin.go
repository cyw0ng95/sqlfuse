package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/stmts/helper"
)

// GenSelectNaturalJoin generates a SELECT with NATURAL JOIN between two tables.
func GenSelectNaturalJoin(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db, "")
	if err != nil || len(tbls) < 2 {
		return SelectStmt{sql: "SELECT 1;", flavor: GetDefaultFlavor()}, nil
	}

	var rnd func(int) int
	if lcg != nil {
		rnd = lcg.Intn
	} else {
		rnd = func(n int) int { return 0 }
	}

	i := rnd(len(tbls))
	j := rnd(len(tbls))
	if j == i && len(tbls) > 1 {
		j = (j + 1) % len(tbls)
	}
	t1 := tbls[i]
	t2 := tbls[j]
	limit := 1 + rnd(50)
	// NATURAL JOIN merges common columns automatically; select * keeps it simple.
	sql := fmt.Sprintf("SELECT * FROM %s NATURAL JOIN %s LIMIT %d;", quoteIdent(t1.Name), quoteIdent(t2.Name), limit)
	return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}
