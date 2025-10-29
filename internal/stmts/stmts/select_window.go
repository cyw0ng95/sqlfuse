package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/stmts/helper"
	"strings"
)

// GenSelectWithWindowFunction generates a SELECT statement with window functions (OVER clause)
// Note: Turso LibSQL does NOT support window functions (OVER clause).
// This function will generate alternative queries for unsupported flavors.
func GenSelectWithWindowFunction(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	// Create a context to check flavor support
	ctx := NewGenContext(db, lcg, 0)
	return genSelectWithWindowFunctionInternal(ctx)
}

// genSelectWithWindowFunctionInternal is the internal implementation that uses GenContext
func genSelectWithWindowFunctionInternal(ctx *GenContext) (SelectStmt, error) {
	// Check if window functions are supported
	if !ctx.SupportsFeature("window_functions") {
		// For flavors that don't support window functions (like Turso),
		// generate a simple SELECT with row numbering via ROWID instead
		if ctx.DB == nil {
			return SelectStmt{sql: "SELECT 1 AS id, 1 AS row_num;", flavor: GetDefaultFlavor()}, nil
		}

		tables, err := helper.GetAllTablesAndCols(ctx.DB)
		if err != nil || len(tables) == 0 {
			return SelectStmt{sql: "SELECT 1 AS id, 1 AS row_num;", flavor: GetDefaultFlavor()}, nil
		}

		rnd := ctx.Intn
		tbl := tables[rnd(len(tables))]
		if len(tbl.Cols) == 0 {
			return SelectStmt{sql: "SELECT 1 AS id, 1 AS row_num;", flavor: GetDefaultFlavor()}, nil
		}

		// Generate a simple SELECT with ROWID for row numbering
		numCols := 1 + rnd(min(3, len(tbl.Cols)))
		selectedCols := make([]string, 0, numCols+1)
		selected := make(map[int]struct{})

		for len(selectedCols) < numCols {
			idx := rnd(len(tbl.Cols))
			if _, ok := selected[idx]; ok {
				continue
			}
			selected[idx] = struct{}{}
			selectedCols = append(selectedCols, quoteIdent(tbl.Cols[idx].Name))
		}

		// Add ROWID as alternative to ROW_NUMBER()
		selectedCols = append(selectedCols, "ROWID AS row_num")

		limit := 1 + rnd(50)
		sql := fmt.Sprintf("SELECT %s FROM %s LIMIT %d;",
			strings.Join(selectedCols, ", "), quoteIdent(tbl.Name), limit)
		return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
	}

	// Original window function logic for flavors that support it
	if ctx.DB == nil {
		return genSelectWindowFunctionLiteral(ctx.LCG), nil
	}

	tables, err := helper.GetAllTablesAndCols(ctx.DB)
	if err != nil || len(tables) == 0 {
		return genSelectWindowFunctionLiteral(ctx.LCG), nil
	}

	rnd := ctx.Intn
	tbl := tables[rnd(len(tables))]

	if len(tbl.Cols) == 0 {
		return genSelectWindowFunctionLiteral(ctx.LCG), nil
	}

	// Find numeric columns for window aggregation
	var numericCols []helper.ColumnInfo
	for _, col := range tbl.Cols {
		if isNumericType(col.Type) {
			numericCols = append(numericCols, col)
		}
	}

	// Choose a window function
	windowFuncs := []string{
		"ROW_NUMBER",
		"RANK",
		"DENSE_RANK",
		"PERCENT_RANK",
		"CUME_DIST",
		"NTILE",
		"LAG",
		"LEAD",
		"FIRST_VALUE",
		"LAST_VALUE",
		"NTH_VALUE",
	}

	funcName := windowFuncs[rnd(len(windowFuncs))]
	var windowExpr string

	// Build the window function expression
	switch funcName {
	case "ROW_NUMBER", "RANK", "DENSE_RANK", "PERCENT_RANK", "CUME_DIST":
		// These functions don't take arguments
		windowExpr = fmt.Sprintf("%s()", funcName)
	case "NTILE":
		// NTILE requires an integer argument
		buckets := 2 + rnd(8) // 2-9 buckets
		windowExpr = fmt.Sprintf("NTILE(%d)", buckets)
	case "LAG", "LEAD":
		// LAG/LEAD can work with any column
		if len(tbl.Cols) > 0 {
			col := tbl.Cols[rnd(len(tbl.Cols))]
			offset := 1 + rnd(3) // offset 1-3
			windowExpr = fmt.Sprintf("%s(%s, %d)", funcName, quoteIdent(col.Name), offset)
		} else {
			windowExpr = fmt.Sprintf("%s(1, 1)", funcName)
		}
	case "FIRST_VALUE", "LAST_VALUE":
		// FIRST_VALUE/LAST_VALUE work with any column
		if len(tbl.Cols) > 0 {
			col := tbl.Cols[rnd(len(tbl.Cols))]
			windowExpr = fmt.Sprintf("%s(%s)", funcName, quoteIdent(col.Name))
		} else {
			windowExpr = fmt.Sprintf("%s(1)", funcName)
		}
	case "NTH_VALUE":
		// NTH_VALUE requires column and N
		if len(tbl.Cols) > 0 {
			col := tbl.Cols[rnd(len(tbl.Cols))]
			n := 1 + rnd(5) // 1-5
			windowExpr = fmt.Sprintf("NTH_VALUE(%s, %d)", quoteIdent(col.Name), n)
		} else {
			windowExpr = fmt.Sprintf("NTH_VALUE(1, 2)")
		}
	default:
		windowExpr = "ROW_NUMBER()"
	}

	// Build the OVER clause with PARTITION BY and/or ORDER BY
	var overClause string
	clauseType := rnd(4)

	switch clauseType {
	case 0:
		// Just OVER()
		overClause = "OVER ()"
	case 1:
		// OVER (ORDER BY col)
		if len(tbl.Cols) > 0 {
			col := tbl.Cols[rnd(len(tbl.Cols))]
			orderDir := "ASC"
			if rnd(2) == 0 {
				orderDir = "DESC"
			}
			overClause = fmt.Sprintf("OVER (ORDER BY %s %s)", quoteIdent(col.Name), orderDir)
		} else {
			overClause = "OVER ()"
		}
	case 2:
		// OVER (PARTITION BY col)
		if len(tbl.Cols) > 1 {
			col := tbl.Cols[rnd(len(tbl.Cols))]
			overClause = fmt.Sprintf("OVER (PARTITION BY %s)", quoteIdent(col.Name))
		} else {
			overClause = "OVER ()"
		}
	default:
		// OVER (PARTITION BY col1 ORDER BY col2)
		if len(tbl.Cols) > 1 {
			partCol := tbl.Cols[rnd(len(tbl.Cols))]
			orderCol := tbl.Cols[rnd(len(tbl.Cols))]
			orderDir := "ASC"
			if rnd(2) == 0 {
				orderDir = "DESC"
			}
			overClause = fmt.Sprintf("OVER (PARTITION BY %s ORDER BY %s %s)",
				quoteIdent(partCol.Name), quoteIdent(orderCol.Name), orderDir)
		} else {
			overClause = "OVER ()"
		}
	}

	// Optionally add frame specification for aggregate functions
	if funcName == "FIRST_VALUE" || funcName == "LAST_VALUE" || funcName == "NTH_VALUE" {
		if rnd(2) == 0 { // 50% chance of adding frame
			frames := []string{
				"ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW",
				"ROWS BETWEEN CURRENT ROW AND UNBOUNDED FOLLOWING",
				"ROWS BETWEEN 1 PRECEDING AND 1 FOLLOWING",
				"RANGE BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW",
			}
			frame := frames[rnd(len(frames))]
			// Insert frame before the closing parenthesis
			overClause = strings.TrimSuffix(overClause, ")") + " " + frame + ")"
		}
	}

	// Select some regular columns along with the window function
	numRegularCols := 1 + rnd(min(3, len(tbl.Cols)))
	selectedCols := make([]string, 0, numRegularCols+1)
	selected := make(map[int]struct{})

	for len(selectedCols) < numRegularCols {
		idx := rnd(len(tbl.Cols))
		if _, ok := selected[idx]; ok {
			continue
		}
		selected[idx] = struct{}{}
		selectedCols = append(selectedCols, quoteIdent(tbl.Cols[idx].Name))
	}

	// Add the window function expression
	windowAlias := fmt.Sprintf("win_func_%d", rnd(1000))
	selectedCols = append(selectedCols, fmt.Sprintf("%s %s AS %s", windowExpr, overClause, quoteIdent(windowAlias)))

	limit := 1 + rnd(50)
	sql := fmt.Sprintf("SELECT %s FROM %s LIMIT %d;",
		strings.Join(selectedCols, ", "), quoteIdent(tbl.Name), limit)
	return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}

// genSelectWindowFunctionLiteral generates a SELECT with window function using literal values
func genSelectWindowFunctionLiteral(lcg *common.LCG) SelectStmt {
	windowExprs := []string{
		"ROW_NUMBER() OVER (ORDER BY 1)",
		"RANK() OVER (ORDER BY 1 DESC)",
		"DENSE_RANK() OVER (ORDER BY 1)",
		"NTILE(4) OVER (ORDER BY 1)",
		"LAG(1, 1) OVER (ORDER BY 1)",
		"LEAD(1, 1) OVER (ORDER BY 1)",
		"FIRST_VALUE(1) OVER (ORDER BY 1 ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW)",
		"LAST_VALUE(1) OVER (ORDER BY 1 ROWS BETWEEN CURRENT ROW AND UNBOUNDED FOLLOWING)",
	}

	funcIdx := lcg.Intn(len(windowExprs))
	sql := fmt.Sprintf("SELECT %s AS win_result;", windowExprs[funcIdx])
	return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}
}

// GenSelectWithMultipleWindows generates a SELECT with multiple window functions
// Note: Turso LibSQL does NOT support window functions.
// This function will generate alternative queries for unsupported flavors.
func GenSelectWithMultipleWindows(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	// Create a context to check flavor support
	ctx := NewGenContext(db, lcg, 0)
	return genSelectWithMultipleWindowsInternal(ctx)
}

// genSelectWithMultipleWindowsInternal is the internal implementation that uses GenContext
func genSelectWithMultipleWindowsInternal(ctx *GenContext) (SelectStmt, error) {
	// Check if window functions are supported
	if !ctx.SupportsFeature("window_functions") {
		// For flavors that don't support window functions,
		// generate a simple SELECT with aggregates and GROUP BY instead
		if ctx.DB == nil {
			return SelectStmt{sql: "SELECT 1 AS id, COUNT(*) AS cnt;", flavor: GetDefaultFlavor()}, nil
		}

		tables, err := helper.GetAllTablesAndCols(ctx.DB)
		if err != nil || len(tables) == 0 {
			return SelectStmt{sql: "SELECT 1 AS id, COUNT(*) AS cnt;", flavor: GetDefaultFlavor()}, nil
		}

		rnd := ctx.Intn
		tbl := tables[rnd(len(tables))]
		if len(tbl.Cols) == 0 {
			return SelectStmt{sql: "SELECT 1 AS id, COUNT(*) AS cnt;", flavor: GetDefaultFlavor()}, nil
		}

		// Generate GROUP BY with aggregates as alternative
		if len(tbl.Cols) > 0 {
			groupCol := tbl.Cols[rnd(len(tbl.Cols))]
			sql := fmt.Sprintf("SELECT %s, COUNT(*) AS cnt FROM %s GROUP BY %s LIMIT %d;",
				quoteIdent(groupCol.Name), quoteIdent(tbl.Name), quoteIdent(groupCol.Name), 1+rnd(30))
			return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
		}
	}

	// Original multiple windows logic for flavors that support it
	if ctx.DB == nil {
		return genSelectMultipleWindowsLiteral(ctx.LCG), nil
	}

	tables, err := helper.GetAllTablesAndCols(ctx.DB)
	if err != nil || len(tables) == 0 {
		return genSelectMultipleWindowsLiteral(ctx.LCG), nil
	}

	rnd := ctx.Intn
	tbl := tables[rnd(len(tables))]

	if len(tbl.Cols) == 0 {
		return genSelectMultipleWindowsLiteral(ctx.LCG), nil
	}

	// Build 2-3 different window expressions
	numWindows := 2 + rnd(2) // 2 or 3 windows
	windowExprs := make([]string, 0, numWindows)

	// Select base columns
	numRegularCols := 1 + rnd(min(2, len(tbl.Cols)))
	selectedCols := make([]string, 0, numRegularCols)
	selected := make(map[int]struct{})

	for len(selectedCols) < numRegularCols {
		idx := rnd(len(tbl.Cols))
		if _, ok := selected[idx]; ok {
			continue
		}
		selected[idx] = struct{}{}
		selectedCols = append(selectedCols, quoteIdent(tbl.Cols[idx].Name))
	}

	// Add window functions
	windowFuncTypes := []string{"ROW_NUMBER", "RANK", "SUM", "AVG", "COUNT"}
	for i := 0; i < numWindows; i++ {
		funcType := windowFuncTypes[rnd(len(windowFuncTypes))]
		var funcExpr string

		if funcType == "ROW_NUMBER" || funcType == "RANK" {
			funcExpr = fmt.Sprintf("%s()", funcType)
		} else if funcType == "COUNT" {
			funcExpr = "COUNT(*)"
		} else {
			// SUM, AVG need numeric column
			var numCol helper.ColumnInfo
			found := false
			for _, col := range tbl.Cols {
				if isNumericType(col.Type) {
					numCol = col
					found = true
					break
				}
			}
			if found {
				funcExpr = fmt.Sprintf("%s(%s)", funcType, quoteIdent(numCol.Name))
			} else {
				funcExpr = fmt.Sprintf("%s(1)", funcType)
			}
		}

		// Different OVER clause for each window
		var overClause string
		if i == 0 && len(tbl.Cols) > 0 {
			// First: simple ORDER BY
			col := tbl.Cols[rnd(len(tbl.Cols))]
			overClause = fmt.Sprintf("OVER (ORDER BY %s)", quoteIdent(col.Name))
		} else if i == 1 && len(tbl.Cols) > 1 {
			// Second: PARTITION BY
			col := tbl.Cols[rnd(len(tbl.Cols))]
			overClause = fmt.Sprintf("OVER (PARTITION BY %s)", quoteIdent(col.Name))
		} else if len(tbl.Cols) > 1 {
			// Third: both
			partCol := tbl.Cols[rnd(len(tbl.Cols))]
			orderCol := tbl.Cols[rnd(len(tbl.Cols))]
			overClause = fmt.Sprintf("OVER (PARTITION BY %s ORDER BY %s DESC)",
				quoteIdent(partCol.Name), quoteIdent(orderCol.Name))
		} else {
			overClause = "OVER ()"
		}

		windowAlias := fmt.Sprintf("win%d", i+1)
		windowExprs = append(windowExprs, fmt.Sprintf("%s %s AS %s", funcExpr, overClause, quoteIdent(windowAlias)))
	}

	allCols := append(selectedCols, windowExprs...)
	limit := 1 + rnd(30)
	sql := fmt.Sprintf("SELECT %s FROM %s LIMIT %d;",
		strings.Join(allCols, ", "), quoteIdent(tbl.Name), limit)
	return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}

func genSelectMultipleWindowsLiteral(lcg *common.LCG) SelectStmt {
	sql := "SELECT 1 AS id, ROW_NUMBER() OVER (ORDER BY 1) AS row_num, RANK() OVER (ORDER BY 1) AS rank_num;"
	return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}
}
