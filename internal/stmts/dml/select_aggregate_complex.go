package dml

import (
	"sqlfuse/internal/stmts/stmts"
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
	"sqlfuse/internal/stmts/helper"
	"strings"
)

// GenSelectAggregateComplex generates a SELECT with multiple aggregate functions.
func GenSelectAggregateComplex(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
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

	// Build multiple aggregate expressions
	aggregates := []string{}
	aggregateFuncs := []string{"COUNT", "SUM", "AVG", "MIN", "MAX"}

	// Add 2-4 different aggregate functions
	numAggs := 2 + rnd(3) // 2..4
	used := make(map[string]bool)

	for i := 0; i < numAggs && len(aggregates) < 5; i++ {
		col := tbl.Cols[rnd(len(tbl.Cols))]
		aggFunc := aggregateFuncs[rnd(len(aggregateFuncs))]

		// COUNT can be used on any column or with *
		if aggFunc == "COUNT" && rnd(2) == 0 {
			key := "COUNT(*)"
			if !used[key] {
				aggregates = append(aggregates, "COUNT(*) AS total_count")
				used[key] = true
			}
			continue
		}

		// For SUM and AVG, prefer numeric columns
		if (aggFunc == "SUM" || aggFunc == "AVG") && !isNumericType(col.Type) && !containsTypeHintSimple(col.Name, "id", "num", "count", "amount", "price", "quantity") {
			// Try to find a numeric column
			found := false
			for _, c := range tbl.Cols {
				if isNumericType(c.Type) || containsTypeHintSimple(c.Name, "id", "num", "count", "amount", "price", "quantity") {
					col = c
					found = true
					break
				}
			}
			if !found {
				aggFunc = "COUNT" // fallback to COUNT if no numeric column
			}
		}

		key := fmt.Sprintf("%s(%s)", aggFunc, col.Name)
		if !used[key] {
			aggregates = append(aggregates, fmt.Sprintf("%s(%s) AS %s_%s",
				aggFunc, stmts.QuoteIdent(col.Name), strings.ToLower(aggFunc), col.Name))
			used[key] = true
		}
	}

	// Ensure we have at least one aggregate
	if len(aggregates) == 0 {
		aggregates = append(aggregates, "COUNT(*) AS total_count")
	}

	// Optionally add GROUP BY
	groupBy := ""
	if rnd(2) == 0 && len(tbl.Cols) > 1 {
		// Find a non-numeric column for grouping (or any column)
		groupCol := tbl.Cols[rnd(len(tbl.Cols))]
		groupBy = fmt.Sprintf(" GROUP BY %s", stmts.QuoteIdent(groupCol.Name))
		// Add the grouping column to the select list
		aggregates = append([]string{stmts.QuoteIdent(groupCol.Name)}, aggregates...)
	}

	sql := fmt.Sprintf("SELECT %s FROM %s%s;", strings.Join(aggregates, ", "), stmts.QuoteIdent(tbl.Name), groupBy)
	return SelectStmt{sql: sql, flavor: stmts.GetDefaultFlavor()}, nil
}
