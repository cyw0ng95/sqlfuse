package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/generators/sqlite/helper"
)

// GenSelectCrossJoin generates a SELECT across two tables using CROSS JOIN.
func GenSelectCrossJoin(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
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

	// pick up to 2 cols from each, alias as a/b
	cols := []string{}
	maxA := min(2, len(t1.Cols))
	for k := 0; k < maxA; k++ {
		c := t1.Cols[rnd(len(t1.Cols))]
		cols = append(cols, fmt.Sprintf("a.%s", quoteIdent(c.Name)))
	}
	maxB := min(2, len(t2.Cols))
	for k := 0; k < maxB; k++ {
		c := t2.Cols[rnd(len(t2.Cols))]
		cols = append(cols, fmt.Sprintf("b.%s", quoteIdent(c.Name)))
	}
	if len(cols) == 0 {
		cols = append(cols, "*")
	}

	limit := 1 + rnd(50)
	sql := fmt.Sprintf("SELECT %s FROM %s AS a CROSS JOIN %s AS b LIMIT %d;",
		joinStrings(cols, ", "), quoteIdent(t1.Name), quoteIdent(t2.Name), limit)
	return SelectStmt{sql: sql}, nil
}

func joinStrings(a []string, sep string) string {
	if len(a) == 0 {
		return ""
	}
	res := a[0]
	for i := 1; i < len(a); i++ {
		res += sep + a[i]
	}
	return res
}
