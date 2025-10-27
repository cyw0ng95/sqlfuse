package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/generators/sqlite/helper"
)

// GenSelectOuterJoin generates a SELECT using LEFT OUTER JOIN (SQLite-compatible).
func GenSelectOuterJoin(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tbls) < 2 {
		return SelectStmt{sql: "SELECT 1;"}, nil
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
		return SelectStmt{sql: "SELECT 1;"}, nil
	}

	on := "1=1"
	for _, c1 := range t1.Cols {
		for _, c2 := range t2.Cols {
			if stringsEqualFold(c1.Name, c2.Name) {
				on = fmt.Sprintf("a.%s = b.%s", quoteIdent(c1.Name), quoteIdent(c2.Name))
				break
			}
		}
		if on != "1=1" {
			break
		}
	}

	cols := []string{fmt.Sprintf("a.%s", quoteIdent(t1.Cols[rnd(len(t1.Cols))].Name))}
	if len(t2.Cols) > 0 {
		cols = append(cols, fmt.Sprintf("b.%s", quoteIdent(t2.Cols[rnd(len(t2.Cols))].Name)))
	}

	limit := 1 + rnd(50)
	sql := fmt.Sprintf("SELECT %s FROM %s AS a LEFT OUTER JOIN %s AS b ON %s LIMIT %d;",
		joinStrings(cols, ", "), quoteIdent(t1.Name), quoteIdent(t2.Name), on, limit)
	return SelectStmt{sql: sql}, nil
}
