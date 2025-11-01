package stmts

import (
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
	"sqlfuse/internal/stmts/helper"
	"sqlfuse/internal/stmts/types"
	"strings"
)

// GenSelectWithCTE generates a SELECT statement with Common Table Expression (WITH clause)
func GenSelectWithCTE(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tables, err := helper.GetAllTablesAndCols(db, "sqlite")
	if err != nil || len(tables) == 0 {
		return genSelectCTELiteral(lcg), nil
	}

	rnd := lcg.Intn
	tbl := tables[rnd(len(tables))]

	if len(tbl.Cols) == 0 {
		return genSelectCTELiteral(lcg), nil
	}

	// Create a CTE name
	cteName := fmt.Sprintf("cte_%d", rnd(1000))

	// Build the CTE query - simple SELECT from existing table
	cteSelectCols := make([]string, 0, min(3, len(tbl.Cols)))
	numCteCols := 1 + rnd(min(3, len(tbl.Cols)))
	selected := make(map[int]struct{})

	for len(cteSelectCols) < numCteCols {
		idx := rnd(len(tbl.Cols))
		if _, ok := selected[idx]; ok {
			continue
		}
		selected[idx] = struct{}{}
		cteSelectCols = append(cteSelectCols, QuoteIdent(tbl.Cols[idx].Name))
	}

	// Add optional WHERE clause to CTE
	cteWhere := ""
	if rnd(2) == 0 && len(tbl.Cols) > 0 {
		col := tbl.Cols[rnd(len(tbl.Cols))]
		val := types.ValueForType(col.Type, lcg, col.Name)
		cteWhere = fmt.Sprintf(" WHERE %s > %s", QuoteIdent(col.Name), val)
	}

	cteQuery := fmt.Sprintf("SELECT %s FROM %s%s",
		strings.Join(cteSelectCols, ", "), QuoteIdent(tbl.Name), cteWhere)

	// Main query that uses the CTE
	mainSelectCols := "*"
	if rnd(2) == 0 {
		// Select specific columns from CTE
		mainSelectCols = strings.Join(cteSelectCols, ", ")
	}

	// Add optional WHERE in main query
	mainWhere := ""
	if rnd(3) == 0 && len(cteSelectCols) > 0 {
		// Reference a column from CTE
		colName := strings.Trim(cteSelectCols[rnd(len(cteSelectCols))], "\"")
		mainWhere = fmt.Sprintf(" WHERE \"%s\" IS NOT NULL", colName)
	}

	limit := 1 + rnd(50)
	sql := fmt.Sprintf("WITH %s AS (%s) SELECT %s FROM %s%s LIMIT %d;",
		QuoteIdent(cteName), cteQuery, mainSelectCols, QuoteIdent(cteName), mainWhere, limit)

	return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}

// genSelectCTELiteral generates a simple CTE with literal values
func genSelectCTELiteral(lcg *common.LCG) SelectStmt {
	ctes := []string{
		"WITH numbers AS (SELECT 1 AS n UNION ALL SELECT 2 UNION ALL SELECT 3) SELECT * FROM numbers;",
		"WITH data AS (SELECT 'test' AS value, 1 AS id) SELECT * FROM data WHERE id > 0;",
		"WITH filtered AS (SELECT 1 AS x, 2 AS y) SELECT x * y AS result FROM filtered;",
	}
	sql := ctes[lcg.Intn(len(ctes))]
	return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}
}

// GenSelectWithMultipleCTE generates a SELECT with multiple CTEs
func GenSelectWithMultipleCTE(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tables, err := helper.GetAllTablesAndCols(db, "sqlite")
	if err != nil || len(tables) == 0 {
		return genSelectMultipleCTELiteral(lcg), nil
	}

	rnd := lcg.Intn
	if len(tables) < 1 {
		return genSelectMultipleCTELiteral(lcg), nil
	}

	// Create 2 CTEs
	cte1Name := fmt.Sprintf("cte1_%d", rnd(1000))
	cte2Name := fmt.Sprintf("cte2_%d", rnd(1000))

	tbl1 := tables[rnd(len(tables))]
	if len(tbl1.Cols) == 0 {
		return genSelectMultipleCTELiteral(lcg), nil
	}

	// First CTE - simple select
	numCols1 := min(2, len(tbl1.Cols))
	cols1 := make([]string, 0, numCols1)
	for i := 0; i < numCols1 && i < len(tbl1.Cols); i++ {
		cols1 = append(cols1, QuoteIdent(tbl1.Cols[i].Name))
	}

	cte1Query := fmt.Sprintf("SELECT %s FROM %s LIMIT %d",
		strings.Join(cols1, ", "), QuoteIdent(tbl1.Name), 10+rnd(40))

	// Second CTE - can reference first CTE or another table
	var cte2Query string
	if rnd(2) == 0 {
		// Reference first CTE
		cte2Query = fmt.Sprintf("SELECT * FROM %s WHERE ROWID %% 2 = 0", QuoteIdent(cte1Name))
	} else {
		// Use another table
		tbl2 := tables[rnd(len(tables))]
		if len(tbl2.Cols) > 0 {
			numCols2 := min(2, len(tbl2.Cols))
			cols2 := make([]string, 0, numCols2)
			for i := 0; i < numCols2 && i < len(tbl2.Cols); i++ {
				cols2 = append(cols2, QuoteIdent(tbl2.Cols[i].Name))
			}
			cte2Query = fmt.Sprintf("SELECT %s FROM %s LIMIT %d",
				strings.Join(cols2, ", "), QuoteIdent(tbl2.Name), 5+rnd(20))
		} else {
			cte2Query = "SELECT 1 AS id, 'test' AS value"
		}
	}

	// Main query combines both CTEs
	mainQuery := ""
	combineType := rnd(3)
	switch combineType {
	case 0:
		// SELECT from first CTE only
		mainQuery = fmt.Sprintf("SELECT * FROM %s", QuoteIdent(cte1Name))
	case 1:
		// SELECT from second CTE only
		mainQuery = fmt.Sprintf("SELECT * FROM %s", QuoteIdent(cte2Name))
	default:
		// UNION both CTEs
		mainQuery = fmt.Sprintf("SELECT * FROM %s UNION SELECT * FROM %s",
			QuoteIdent(cte1Name), QuoteIdent(cte2Name))
	}

	limit := 1 + rnd(50)
	sql := fmt.Sprintf("WITH %s AS (%s), %s AS (%s) %s LIMIT %d;",
		QuoteIdent(cte1Name), cte1Query, QuoteIdent(cte2Name), cte2Query, mainQuery, limit)

	return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}

func genSelectMultipleCTELiteral(lcg *common.LCG) SelectStmt {
	sql := "WITH a AS (SELECT 1 AS x), b AS (SELECT 2 AS y) SELECT a.x, b.y FROM a, b;"
	return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}
}

// GenSelectWithRecursiveCTE generates a SELECT with recursive CTE
// Note: Turso LibSQL does NOT support RECURSIVE keyword in CTEs.
// This function will generate non-recursive CTEs for Turso flavor.
func GenSelectWithRecursiveCTE(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	// Create a default context to check flavor support
	// This maintains backward compatibility while supporting flavor checks
	ctx := NewGenContext(db, lcg, 0)
	return genSelectWithRecursiveCTEInternal(ctx)
}

// genSelectWithRecursiveCTEInternal is the internal implementation that uses GenContext
func genSelectWithRecursiveCTEInternal(ctx *GenContext) (SelectStmt, error) {
	rnd := ctx.Intn
	cteName := fmt.Sprintf("recursive_cte_%d", rnd(1000))

	// Check if RECURSIVE CTEs are supported by the current flavor
	if !ctx.SupportsFeature("cte_recursive") {
		// For flavors that don't support RECURSIVE (like Turso),
		// generate a regular CTE with multiple SELECT UNION instead
		sql := fmt.Sprintf("WITH %s AS (SELECT 1 AS n UNION ALL SELECT 2 UNION ALL SELECT 3) SELECT * FROM %s;",
			QuoteIdent(cteName), QuoteIdent(cteName))
		return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
	}

	// Recursive CTEs follow pattern: WITH RECURSIVE name AS (base UNION ALL recursive)
	// Simple recursive examples
	recursivePatterns := []string{
		// Count from 1 to N
		fmt.Sprintf("WITH RECURSIVE %s(n) AS (SELECT 1 UNION ALL SELECT n+1 FROM %s WHERE n < %d) SELECT * FROM %s;",
			QuoteIdent(cteName), QuoteIdent(cteName), 5+rnd(15), QuoteIdent(cteName)),

		// Fibonacci-like sequence
		fmt.Sprintf("WITH RECURSIVE %s(a, b) AS (SELECT 0, 1 UNION ALL SELECT b, a+b FROM %s WHERE b < %d) SELECT * FROM %s;",
			QuoteIdent(cteName), QuoteIdent(cteName), 100+rnd(900), QuoteIdent(cteName)),

		// Powers of 2
		fmt.Sprintf("WITH RECURSIVE %s(n, val) AS (SELECT 1, 1 UNION ALL SELECT n+1, val*2 FROM %s WHERE n < %d) SELECT * FROM %s;",
			QuoteIdent(cteName), QuoteIdent(cteName), 5+rnd(10), QuoteIdent(cteName)),
	}

	sql := recursivePatterns[rnd(len(recursivePatterns))]
	return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}
