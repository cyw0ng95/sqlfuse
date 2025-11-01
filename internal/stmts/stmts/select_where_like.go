package stmts

import (
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
	"sqlfuse/internal/stmts/helper"
	"strings"
)

// GenSelectWhereLike generates a SELECT with a LIKE predicate on a text column.
func GenSelectWhereLike(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db, "sqlite")
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
		return SelectStmt{sql: fmt.Sprintf("SELECT * FROM %s;", QuoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
	}

	// find a text-like column
	textIdx := -1
	for i, c := range tbl.Cols {
		if isTextType(c.Type) || containsTypeHintSimple(c.Name, "name", "email", "title", "desc") {
			textIdx = i
			break
		}
	}

	// pick columns to select (include the text column if found)
	cols := []helper.ColumnInfo{}
	if textIdx >= 0 {
		cols = append(cols, tbl.Cols[textIdx])
	}
	// fill rest up to 3
	for len(cols) < 3 && len(cols) < len(tbl.Cols) {
		idx := rnd(len(tbl.Cols))
		// avoid duplicates
		d := false
		for _, ex := range cols {
			if ex.Name == tbl.Cols[idx].Name {
				d = true
				break
			}
		}
		if d {
			continue
		}
		cols = append(cols, tbl.Cols[idx])
	}

	if len(cols) == 0 {
		// fallback
		cols = append(cols, tbl.Cols[0])
	}

	// build LIKE pattern
	var likeVal string
	if textIdx >= 0 {
		likeVal = fmt.Sprintf("'%%%s%%'", "a")
	} else {
		likeVal = "'%a%'"
	}

	where := fmt.Sprintf(" WHERE %s LIKE %s", QuoteIdent(cols[0].Name), likeVal)
	limit := 1 + rnd(50)
	sql := fmt.Sprintf("SELECT %s FROM %s%s LIMIT %d;", joinCols(cols), QuoteIdent(tbl.Name), where, limit)
	return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}

func isTextType(t string) bool {
	if t == "" {
		return false
	}
	up := strings.ToUpper(t)
	return strings.Contains(up, "CHAR") || strings.Contains(up, "CLOB") || strings.Contains(up, "TEXT") || strings.Contains(up, "VARCHAR")
}
