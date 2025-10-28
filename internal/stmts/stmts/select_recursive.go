package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/stmts/helper"
	"strings"
)

// GenSelectRecursive generates a SELECT statement with recursive/nested features.
// maxDepth controls how deep the recursion can go (0 = no recursion, simple SELECT).
func GenSelectRecursive(db *sql.DB, lcg *common.LCG, maxDepth int) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tbls) == 0 {
		return SelectStmt{sql: "SELECT 1;", flavor: GetDefaultFlavor()}, nil
	}

	ctx := NewGenContext(db, lcg, maxDepth)
	exprGen := NewExprGenerator(ctx)

	// Pick a primary table
	tbl := tbls[ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		return SelectStmt{sql: fmt.Sprintf("SELECT * FROM %s;", quoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
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

	return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}

// genFromClause generates a FROM clause, possibly with a subquery.
func genFromClause(ctx *GenContext, tbls []helper.TableInfo, defaultTbl helper.TableInfo) string {
	// Use subquery in FROM clause for recursion
	if ctx.CanRecurse() && ctx.Intn(3) == 0 && len(tbls) > 0 {
		subCtx := ctx.Descend()
		exprGen := NewExprGenerator(subCtx)
		subquery := exprGen.GenSubquery(tbls)
		return fmt.Sprintf("(%s) AS subq", subquery)
	}

	// Regular table reference
	return quoteIdent(defaultTbl.Name)
}

// GenSelectWithNestedCase generates a SELECT with nested CASE expressions.
func GenSelectWithNestedCase(db *sql.DB, lcg *common.LCG, maxDepth int) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tbls) == 0 {
		return SelectStmt{sql: "SELECT 1;", flavor: GetDefaultFlavor()}, nil
	}

	ctx := NewGenContext(db, lcg, maxDepth)
	tbl := tbls[ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		return SelectStmt{sql: fmt.Sprintf("SELECT * FROM %s;", quoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
	}

	// Generate nested CASE expression
	caseExpr := genNestedCase(ctx, tbl, 0)

	// Add some regular columns
	selectCols := []string{caseExpr}
	numCols := 1 + ctx.Intn(2)
	for i := 0; i < numCols && i < len(tbl.Cols); i++ {
		selectCols = append(selectCols, quoteIdent(tbl.Cols[i].Name))
	}

	limit := 1 + ctx.Intn(50)
	sql := fmt.Sprintf("SELECT %s FROM %s LIMIT %d;",
		strings.Join(selectCols, ", "), quoteIdent(tbl.Name), limit)

	return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}

// genNestedCase generates a potentially nested CASE expression.
func genNestedCase(ctx *GenContext, tbl helper.TableInfo, depth int) string {
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
			quoteIdent(col.Name), nestedCase, depth)
	}

	// Simple CASE
	return fmt.Sprintf("CASE WHEN %s IS NOT NULL THEN 'present' ELSE 'absent' END AS case_%d",
		quoteIdent(col.Name), depth)
}

// GenSelectWithComplexJoin generates a SELECT with potentially nested subqueries in joins.
func GenSelectWithComplexJoin(db *sql.DB, lcg *common.LCG, maxDepth int) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tbls) < 2 {
		return SelectStmt{sql: "SELECT 1;", flavor: GetDefaultFlavor()}, nil
	}

	ctx := NewGenContext(db, lcg, maxDepth)

	// Pick two tables for joining
	tbl1 := tbls[ctx.Intn(len(tbls))]
	tbl2 := tbls[ctx.Intn(len(tbls))]

	// Ensure we have different tables if possible
	if len(tbls) > 1 && tbl1.Name == tbl2.Name {
		tbl2 = tbls[(ctx.Intn(len(tbls)-1)+1)%len(tbls)]
	}

	if len(tbl1.Cols) == 0 || len(tbl2.Cols) == 0 {
		return SelectStmt{sql: "SELECT 1;", flavor: GetDefaultFlavor()}, nil
	}

	// Select columns from both tables
	selectCols := []string{
		fmt.Sprintf("%s.%s", quoteIdent(tbl1.Name), quoteIdent(tbl1.Cols[0].Name)),
		fmt.Sprintf("%s.%s", quoteIdent(tbl2.Name), quoteIdent(tbl2.Cols[0].Name)),
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
		quoteIdent(tbl1.Name), quoteIdent(joinCol1),
		quoteIdent(tbl2.Name), quoteIdent(joinCol2))

	// Don't add WHERE clause for joins to avoid column ambiguity
	// (would need qualified column names which is complex)
	whereClause := ""

	limit := 1 + ctx.Intn(50)
	sql := fmt.Sprintf("SELECT %s FROM %s INNER JOIN %s ON %s%s LIMIT %d;",
		strings.Join(selectCols, ", "),
		quoteIdent(tbl1.Name),
		quoteIdent(tbl2.Name),
		joinCond,
		whereClause,
		limit)

	return SelectStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}
