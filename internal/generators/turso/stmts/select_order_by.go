package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/generators/turso/helper"
)

// GenSelectOrderBy generates a SELECT with ORDER BY on one or two columns.
func GenSelectOrderBy(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
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

	// pick columns
	cols := []helper.ColumnInfo{}
	cols = append(cols, tbl.Cols[rnd(len(tbl.Cols))])
	if len(tbl.Cols) > 1 && rnd(2) == 0 {
		cols = append(cols, tbl.Cols[rnd(len(tbl.Cols))])
	}

	// build order by
	order := quoteIdent(cols[0].Name)
	if len(cols) > 1 {
		order += ", " + quoteIdent(cols[1].Name)
	}

	sql := fmt.Sprintf("SELECT %s FROM %s ORDER BY %s;", joinCols(cols), quoteIdent(tbl.Name), order)
	return SelectStmt{sql: sql}, nil
}
