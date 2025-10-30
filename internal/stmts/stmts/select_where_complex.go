package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/stmts/helper"
	"sqlsmith-go/internal/stmts/types"
	"strings"
)

// GenSelectWhereComplex generates a SELECT with complex WHERE clause using AND/OR combinations.
func GenSelectWhereComplex(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
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
		return SelectStmt{sql: fmt.Sprintf("SELECT * FROM %s;", quoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
	}

	// pick up to 3 columns for selection
	maxCols := len(tbl.Cols)
	if maxCols > 3 {
		maxCols = 3
	}
	nCols := 1 + rnd(maxCols)
	if nCols > len(tbl.Cols) {
		nCols = len(tbl.Cols)
	}

	selected := make(map[int]struct{}, nCols)
	cols := make([]helper.ColumnInfo, 0, nCols)
	for len(cols) < nCols {
		idx := rnd(len(tbl.Cols))
		if _, ok := selected[idx]; ok {
			continue
		}
		selected[idx] = struct{}{}
		cols = append(cols, tbl.Cols[idx])
	}

	// Build complex WHERE clause with 2-4 conditions
	numConditions := 2 + rnd(3) // 2..4 conditions
	conditions := []string{}

	for i := 0; i < numConditions && i < len(tbl.Cols); i++ {
		col := tbl.Cols[rnd(len(tbl.Cols))]
		val := types.ValueForType(col.Type, lcg, col.Name)

		// Choose a random operator
		operators := []string{"=", ">", "<", ">=", "<=", "!="}
		if strings.Contains(strings.ToUpper(col.Type), "TEXT") || strings.Contains(strings.ToUpper(col.Type), "CHAR") {
			operators = []string{"=", "!=", "LIKE"}
		}
		op := operators[rnd(len(operators))]

		// For LIKE operator, modify the value to include wildcards
		if op == "LIKE" && val != "NULL" {
			// Only modify if val is a properly quoted string literal
			if strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'") && len(val) >= 2 {
				unquoted := strings.TrimSuffix(strings.TrimPrefix(val, "'"), "'")
				val = fmt.Sprintf("'%%%s%%'", escapeSingle(unquoted))
			}
			// else: leave val as-is (could be unquoted or invalid, but don't modify)
		}

		conditions = append(conditions, fmt.Sprintf("%s %s %s", quoteIdent(col.Name), op, val))
	}

	// Combine conditions with AND/OR
	where := ""
	if len(conditions) > 0 {
		// Randomly choose AND or OR for combining
		connector := " AND "
		if rnd(2) == 0 {
			connector = " OR "
		}

		// For more complexity, sometimes mix AND and OR
		if len(conditions) > 2 && rnd(2) == 0 {
			// Mix: (cond1 AND cond2) OR cond3
			where = fmt.Sprintf(" WHERE (%s AND %s)", conditions[0], conditions[1])
			for i := 2; i < len(conditions); i++ {
				where += fmt.Sprintf(" OR %s", conditions[i])
			}
		} else {
			where = fmt.Sprintf(" WHERE %s", strings.Join(conditions, connector))
		}
	}

	limit := 1 + rnd(50)
	sql := fmt.Sprintf("SELECT %s FROM %s%s LIMIT %d;", joinCols(cols), quoteIdent(tbl.Name), where, limit)
	return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}
