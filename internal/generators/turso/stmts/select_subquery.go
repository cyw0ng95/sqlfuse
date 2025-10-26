package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/generators/turso/helper"
	"sqlsmith-go/internal/generators/turso/types"
	"strings"
)

// GenSelectSubquery generates a SELECT with a subquery in WHERE clause.
func GenSelectSubquery(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tbls) < 2 {
		// Need at least 2 tables for a meaningful subquery
		return SelectStmt{sql: "SELECT 1;"}, nil
	}

	var rnd func(int) int
	if lcg != nil {
		rnd = lcg.Intn
	} else {
		rnd = func(n int) int { return 0 }
	}

	// Pick outer query table
	outerTbl := tbls[rnd(len(tbls))]
	if len(outerTbl.Cols) == 0 {
		return SelectStmt{sql: fmt.Sprintf("SELECT * FROM %s;", quoteIdent(outerTbl.Name))}, nil
	}

	// Pick inner query table (different from outer if possible)
	innerTbl := tbls[rnd(len(tbls))]
	if len(tbls) > 1 && innerTbl.Name == outerTbl.Name {
		innerTbl = tbls[(rnd(len(tbls)-1)+1)%len(tbls)]
	}
	if len(innerTbl.Cols) == 0 {
		return SelectStmt{sql: fmt.Sprintf("SELECT * FROM %s;", quoteIdent(outerTbl.Name))}, nil
	}

	// Select columns for outer query
	maxCols := len(outerTbl.Cols)
	if maxCols > 3 {
		maxCols = 3
	}
	nCols := 1 + rnd(maxCols)
	if nCols > len(outerTbl.Cols) {
		nCols = len(outerTbl.Cols)
	}

	selected := make(map[int]struct{}, nCols)
	cols := make([]helper.ColumnInfo, 0, nCols)
	for len(cols) < nCols {
		idx := rnd(len(outerTbl.Cols))
		if _, ok := selected[idx]; ok {
			continue
		}
		selected[idx] = struct{}{}
		cols = append(cols, outerTbl.Cols[idx])
	}

	// Try to find a common column between tables for correlation
	var outerCol, innerCol helper.ColumnInfo
	foundCommon := false
	for _, oc := range outerTbl.Cols {
		for _, ic := range innerTbl.Cols {
			if strings.EqualFold(oc.Name, ic.Name) {
				outerCol = oc
				innerCol = ic
				foundCommon = true
				break
			}
		}
		if foundCommon {
			break
		}
	}

	// If no common column, just use first columns
	if !foundCommon {
		outerCol = outerTbl.Cols[0]
		innerCol = innerTbl.Cols[0]
	}

	// Use EXISTS subquery with correlation (LibSQL doesn't support IN with subquery)
	// Build correlation condition
	correlationWhere := fmt.Sprintf("%s = %s.%s", 
		quoteIdent(innerCol.Name), 
		quoteIdent(outerTbl.Name), 
		quoteIdent(outerCol.Name))
	
	// Optionally add an AND condition to the subquery
	if rnd(2) == 0 && len(innerTbl.Cols) > 1 {
		// Add a simple additional condition to the subquery
		whereCol := innerTbl.Cols[rnd(len(innerTbl.Cols))]
		val := types.ValueForType(whereCol.Type, lcg, whereCol.Name)
		correlationWhere = fmt.Sprintf("%s AND %s > %s", correlationWhere, quoteIdent(whereCol.Name), val)
	}

	// Use EXISTS subquery with correlation
	subquery := fmt.Sprintf("SELECT 1 FROM %s WHERE %s LIMIT 1",
		quoteIdent(innerTbl.Name), correlationWhere)
	where := fmt.Sprintf(" WHERE EXISTS (%s)", subquery)

	limit := 1 + rnd(50)
	sql := fmt.Sprintf("SELECT %s FROM %s%s LIMIT %d;", joinCols(cols), quoteIdent(outerTbl.Name), where, limit)
	return SelectStmt{sql: sql}, nil
}
