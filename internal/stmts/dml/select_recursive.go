package dml

import (
	"sqlfuse/internal/stmts/stmts"
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
	"sqlfuse/internal/stmts/helper"
	"sqlfuse/internal/stmts/other"
	"strings"
)

// GenSelectRecursive generates a SELECT statement with recursive/nested features.
// maxDepth controls how deep the recursion can go (0 = no recursion, simple SELECT).
func GenSelectRecursive(db *sql.DB, lcg *common.LCG, maxDepth int) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db, "sqlite")
	if err != nil || len(tbls) == 0 {
		return SelectStmt{sql: "SELECT 1;", flavor: stmts.GetDefaultFlavor()}, nil
	}

	ctx := stmts.NewGenContext(db, lcg, maxDepth)
	exprGen := other.NewExprGenerator(ctx)

	// Pick a primary table
	tbl := tbls[ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		return SelectStmt{sql: fmt.Sprintf("SELECT * FROM %s;", stmts.QuoteIdent(tbl.Name)), flavor: stmts.GetDefaultFlavor()}, nil
	}

	// Generate SELECT expressions (may include complex nested expressions)
	selectExprs := exprGen.GenSelectExpr([]helper.TableInfo{tbl}, 3)

	// Generate FROM clause (may include subquery)
	fromClause := genFromClause(ctx, tbls, tbl)

	// Generate WHERE clause (may include nested conditions)
	whereClause := ""
	if ctx.Intn(2) == 0 {
		whereExpr := exprGen.GenWhereExpr([]helper.TableInfo{tbl})
		if whereExpr != "" {
			whereClause = fmt.Sprintf(" WHERE %s", whereExpr)
		}
	}

	limit := 1 + ctx.Intn(50)
	sql := fmt.Sprintf("SELECT %s FROM %s%s LIMIT %d;",
		strings.Join(selectExprs, ", "), fromClause, whereClause, limit)

	return SelectStmt{sql: sql, flavor: stmts.GetDefaultFlavor()}, nil
}

// genFromClause generates a FROM clause, possibly with a subquery.
func genFromClause(ctx *stmts.GenContext, tbls []helper.TableInfo, defaultTbl helper.TableInfo) string {
	// Use subquery in FROM clause for recursion more frequently
	if ctx.CanRecurse() && ctx.Intn(2) == 0 && len(tbls) > 0 {
		subCtx := ctx.Descend()
		exprGen := other.NewExprGenerator(subCtx)
		subquery := exprGen.GenSubquery(tbls)
		return fmt.Sprintf("(%s) AS subq", subquery)
	}

	// Regular table reference
	return stmts.QuoteIdent(defaultTbl.Name)
}

// GenSelectWithNestedCase generates a SELECT with nested CASE expressions.
func GenSelectWithNestedCase(db *sql.DB, lcg *common.LCG, maxDepth int) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db, "sqlite")
	if err != nil || len(tbls) == 0 {
		return SelectStmt{sql: "SELECT 1;", flavor: stmts.GetDefaultFlavor()}, nil
	}

	ctx := stmts.NewGenContext(db, lcg, maxDepth)
	tbl := tbls[ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		return SelectStmt{sql: fmt.Sprintf("SELECT * FROM %s;", stmts.QuoteIdent(tbl.Name)), flavor: stmts.GetDefaultFlavor()}, nil
	}

	// Generate nested CASE expression
	caseExpr := genNestedCase(ctx, tbl, 0)

	// Add some regular columns
	selectCols := []string{caseExpr}
	numCols := 1 + ctx.Intn(2)
	for i := 0; i < numCols && i < len(tbl.Cols); i++ {
		selectCols = append(selectCols, stmts.QuoteIdent(tbl.Cols[i].Name))
	}

	limit := 1 + ctx.Intn(50)
	sql := fmt.Sprintf("SELECT %s FROM %s LIMIT %d;",
		strings.Join(selectCols, ", "), stmts.QuoteIdent(tbl.Name), limit)

	return SelectStmt{sql: sql, flavor: stmts.GetDefaultFlavor()}, nil
}

// genNestedCase generates a potentially nested CASE expression.
func genNestedCase(ctx *stmts.GenContext, tbl helper.TableInfo, depth int) string {
	if len(tbl.Cols) == 0 {
		return "CASE WHEN 1=1 THEN 'default' ELSE 'other' END"
	}

	col := tbl.Cols[ctx.Intn(len(tbl.Cols))]

	// Decide whether to nest deeper
	if ctx.CanRecurse() && ctx.Intn(2) == 0 {
		// Nested CASE in the THEN clause
		subCtx := ctx.Descend()
		nestedCase := genNestedCase(subCtx, tbl, depth+1)
		return fmt.Sprintf("CASE WHEN %s IS NOT NULL THEN (%s) ELSE 'null' END AS nested_case_%d",
			stmts.QuoteIdent(col.Name), nestedCase, depth)
	}

	// Simple CASE
	return fmt.Sprintf("CASE WHEN %s IS NOT NULL THEN 'present' ELSE 'absent' END AS case_%d",
		stmts.QuoteIdent(col.Name), depth)
}

// GenSelectWithComplexJoin generates a SELECT with potentially nested subqueries in joins.
func GenSelectWithComplexJoin(db *sql.DB, lcg *common.LCG, maxDepth int) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db, "sqlite")
	if err != nil || len(tbls) < 2 {
		return SelectStmt{sql: "SELECT 1;", flavor: stmts.GetDefaultFlavor()}, nil
	}

	ctx := stmts.NewGenContext(db, lcg, maxDepth)

	// Pick two tables for joining
	tbl1 := tbls[ctx.Intn(len(tbls))]
	tbl2 := tbls[ctx.Intn(len(tbls))]

	// Ensure we have different tables if possible
	if len(tbls) > 1 && tbl1.Name == tbl2.Name {
		tbl2 = tbls[(ctx.Intn(len(tbls)-1)+1)%len(tbls)]
	}

	if len(tbl1.Cols) == 0 || len(tbl2.Cols) == 0 {
		return SelectStmt{sql: "SELECT 1;", flavor: stmts.GetDefaultFlavor()}, nil
	}

	// Select columns from both tables
	selectCols := []string{
		fmt.Sprintf("%s.%s", stmts.QuoteIdent(tbl1.Name), stmts.QuoteIdent(tbl1.Cols[0].Name)),
		fmt.Sprintf("%s.%s", stmts.QuoteIdent(tbl2.Name), stmts.QuoteIdent(tbl2.Cols[0].Name)),
	}

	// Try to find common column for join condition
	joinCol1 := tbl1.Cols[0].Name
	joinCol2 := tbl2.Cols[0].Name

	for _, c1 := range tbl1.Cols {
		for _, c2 := range tbl2.Cols {
			if strings.EqualFold(c1.Name, c2.Name) {
				joinCol1 = c1.Name
				joinCol2 = c2.Name
				break
			}
		}
	}

	// Build join condition
	joinCond := fmt.Sprintf("%s.%s = %s.%s",
		stmts.QuoteIdent(tbl1.Name), stmts.QuoteIdent(joinCol1),
		stmts.QuoteIdent(tbl2.Name), stmts.QuoteIdent(joinCol2))

	// Don't add WHERE clause for joins to avoid column ambiguity
	// (would need qualified column names which is complex)
	whereClause := ""

	limit := 1 + ctx.Intn(50)
	sql := fmt.Sprintf("SELECT %s FROM %s INNER JOIN %s ON %s%s LIMIT %d;",
		strings.Join(selectCols, ", "),
		stmts.QuoteIdent(tbl1.Name),
		stmts.QuoteIdent(tbl2.Name),
		joinCond,
		whereClause,
		limit)

	return SelectStmt{sql: sql, flavor: stmts.GetDefaultFlavor()}, nil
}

// GenSelectDeeplyNested generates a SELECT with deeply nested subqueries.
// This creates complex queries with multiple levels of nesting for stress testing.
func GenSelectDeeplyNested(db *sql.DB, lcg *common.LCG, maxDepth int) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db, "sqlite")
	if err != nil || len(tbls) == 0 {
		return SelectStmt{sql: "SELECT 1;", flavor: stmts.GetDefaultFlavor()}, nil
	}

	ctx := stmts.NewGenContext(db, lcg, maxDepth)

	// Build a deeply nested query recursively
	sql := genDeeplyNestedQuery(ctx, tbls, 0)

	return SelectStmt{sql: sql, flavor: stmts.GetDefaultFlavor()}, nil
}

// genDeeplyNestedQuery recursively builds nested subqueries.
func genDeeplyNestedQuery(ctx *stmts.GenContext, tbls []helper.TableInfo, currentDepth int) string {
	if !ctx.CanRecurse() || len(tbls) == 0 {
		// Base case: simple SELECT
		tbl := tbls[ctx.Intn(len(tbls))]
		if len(tbl.Cols) == 0 {
			return "SELECT 1"
		}
		col := tbl.Cols[ctx.Intn(len(tbl.Cols))]
		limit := 1 + ctx.Intn(10)
		return fmt.Sprintf("SELECT %s FROM %s LIMIT %d",
			stmts.QuoteIdent(col.Name), stmts.QuoteIdent(tbl.Name), limit)
	}

	// Recursive case: build a subquery
	tbl := tbls[ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		return "SELECT 1"
	}

	// Descend to build the inner query
	subCtx := ctx.Descend()
	innerQuery := genDeeplyNestedQuery(subCtx, tbls, currentDepth+1)

	// Wrap it with an outer query
	col := tbl.Cols[ctx.Intn(len(tbl.Cols))]
	limit := 1 + ctx.Intn(20)

	// Optionally add WHERE clause
	whereClause := ""
	if ctx.Intn(2) == 0 {
		whereClause = fmt.Sprintf(" WHERE %s IS NOT NULL", stmts.QuoteIdent(col.Name))
	}

	return fmt.Sprintf("SELECT * FROM (%s) AS nested_%d%s LIMIT %d",
		innerQuery, currentDepth, whereClause, limit)
}
