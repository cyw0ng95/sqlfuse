package dml

import (
	"sqlfuse/internal/stmts/stmts"
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
	"sqlfuse/internal/stmts/helper"
)

// GenSelectInnerJoin generates a SELECT using INNER JOIN with a simple ON clause.
func GenSelectInnerJoin(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db, "sqlite")
	if err != nil || len(tbls) < 2 {
		return SelectStmt{sql: "SELECT 1;", flavor: stmts.GetDefaultFlavor()}, nil
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
	if len(t1.Cols) == 0 || len(t2.Cols) == 0 {
		return SelectStmt{sql: "SELECT 1;", flavor: stmts.GetDefaultFlavor()}, nil
	}

	// build ON using first common column name if any
	on := "1=1"
	for _, c1 := range t1.Cols {
		for _, c2 := range t2.Cols {
			if stringsEqualFold(c1.Name, c2.Name) {
				on = fmt.Sprintf("a.%s = b.%s", stmts.QuoteIdent(c1.Name), stmts.QuoteIdent(c2.Name))
				break
			}
		}
		if on != "1=1" {
			break
		}
	}

	// select a few columns
	cols := []string{fmt.Sprintf("a.%s", stmts.QuoteIdent(t1.Cols[rnd(len(t1.Cols))].Name))}
	if len(t2.Cols) > 0 {
		cols = append(cols, fmt.Sprintf("b.%s", stmts.QuoteIdent(t2.Cols[rnd(len(t2.Cols))].Name)))
	}

	limit := 1 + rnd(50)
	sql := fmt.Sprintf("SELECT %s FROM %s AS a INNER JOIN %s AS b ON %s LIMIT %d;",
		joinStrings(cols, ", "), stmts.QuoteIdent(t1.Name), stmts.QuoteIdent(t2.Name), on, limit)
	return SelectStmt{sql: sql, flavor: stmts.GetDefaultFlavor()}, nil
}

func stringsEqualFold(a, b string) bool {
	if len(a) != len(b) {
		return stringsToUpper(a) == stringsToUpper(b)
	}
	return stringsToUpper(a) == stringsToUpper(b)
}
