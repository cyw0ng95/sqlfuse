package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/generators/turso/helper"
)

// GenSelectJoinUsing generates a SELECT with JOIN ... USING(col) if a common column exists.
func GenSelectJoinUsing(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
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

	using := ""
	for _, c1 := range t1.Cols {
		for _, c2 := range t2.Cols {
			if stringsEqualFold(c1.Name, c2.Name) {
				using = c1.Name
				break
			}
		}
		if using != "" {
			break
		}
	}

	cols := []string{fmt.Sprintf("a.%s", quoteIdent(t1.Cols[rnd(len(t1.Cols))].Name))}
	if len(t2.Cols) > 0 {
		cols = append(cols, fmt.Sprintf("b.%s", quoteIdent(t2.Cols[rnd(len(t2.Cols))].Name)))
	}

	limit := 1 + rnd(50)
	var sql string
	if using != "" {
		sql = fmt.Sprintf("SELECT %s FROM %s AS a JOIN %s AS b USING(%s) LIMIT %d;",
			joinStrings(cols, ", "), quoteIdent(t1.Name), quoteIdent(t2.Name), quoteIdent(using), limit)
	} else {
		// fallback to INNER JOIN ON 1=1 if no common column
		sql = fmt.Sprintf("SELECT %s FROM %s AS a INNER JOIN %s AS b ON 1=1 LIMIT %d;",
			joinStrings(cols, ", "), quoteIdent(t1.Name), quoteIdent(t2.Name), limit)
	}
	return SelectStmt{sql: sql}, nil
}
