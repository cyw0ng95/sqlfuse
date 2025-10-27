package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/generators/sqlite/helper"
	"strings"
)

// GenSelectWithExpressions generates a SELECT statement with various expression types in SELECT list
func GenSelectWithExpressions(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tbls) == 0 {
		return SelectStmt{sql: "SELECT 1;"}, nil
	}

	ctx := NewGenContext(db, lcg, 2)
	eg := NewExprGenerator(ctx)

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

	// Generate 1-3 expressions for the SELECT list
	numExprs := 1 + rnd(3)
	exprs := make([]string, 0, numExprs)

	for i := 0; i < numExprs; i++ {
		expr := eg.GenRandomExpr(tbls)
		exprs = append(exprs, expr)
	}

	limit := 1 + rnd(50)
	sql := fmt.Sprintf("SELECT %s FROM %s LIMIT %d;", strings.Join(exprs, ", "), quoteIdent(tbl.Name), limit)
	return SelectStmt{sql: sql}, nil
}

// GenSelectWhereCast generates a SELECT with CAST in WHERE clause
func GenSelectWhereCast(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tbls) == 0 {
		return SelectStmt{sql: "SELECT 1;"}, nil
	}

	ctx := NewGenContext(db, lcg, 2)
	eg := NewExprGenerator(ctx)

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

	// Build WHERE clause with CAST
	castExpr := eg.GenCastExpr(tbls)
	where := fmt.Sprintf(" WHERE %s IS NOT NULL", castExpr)

	limit := 1 + rnd(50)
	sql := fmt.Sprintf("SELECT * FROM %s%s LIMIT %d;", quoteIdent(tbl.Name), where, limit)
	return SelectStmt{sql: sql}, nil
}

// GenSelectWhereBetween generates a SELECT with BETWEEN in WHERE clause
func GenSelectWhereBetween(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tbls) == 0 {
		return SelectStmt{sql: "SELECT 1;"}, nil
	}

	ctx := NewGenContext(db, lcg, 2)
	eg := NewExprGenerator(ctx)

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

	// Build WHERE clause with BETWEEN
	betweenExpr := eg.GenBetweenExpr(tbls, rnd(2) == 0)
	where := fmt.Sprintf(" WHERE %s", betweenExpr)

	limit := 1 + rnd(50)
	sql := fmt.Sprintf("SELECT * FROM %s%s LIMIT %d;", quoteIdent(tbl.Name), where, limit)
	return SelectStmt{sql: sql}, nil
}

// GenSelectWhereGlob generates a SELECT with GLOB in WHERE clause
func GenSelectWhereGlob(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tbls) == 0 {
		return SelectStmt{sql: "SELECT 1;"}, nil
	}

	ctx := NewGenContext(db, lcg, 2)
	eg := NewExprGenerator(ctx)

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

	// Build WHERE clause with GLOB
	globExpr := eg.GenGlobExpr(tbls, rnd(2) == 0)
	where := fmt.Sprintf(" WHERE %s", globExpr)

	limit := 1 + rnd(50)
	sql := fmt.Sprintf("SELECT * FROM %s%s LIMIT %d;", quoteIdent(tbl.Name), where, limit)
	return SelectStmt{sql: sql}, nil
}

// GenSelectWithCollate generates a SELECT with COLLATE in ORDER BY clause
func GenSelectWithCollate(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tbls) == 0 {
		return SelectStmt{sql: "SELECT 1;"}, nil
	}

	ctx := NewGenContext(db, lcg, 2)
	eg := NewExprGenerator(ctx)

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

	// Find a text column for COLLATE
	var textCol *helper.ColumnInfo
	for _, col := range tbl.Cols {
		colTypeUpper := strings.ToUpper(col.Type)
		if strings.Contains(colTypeUpper, "TEXT") || strings.Contains(colTypeUpper, "CHAR") {
			textCol = &col
			break
		}
	}

	orderBy := ""
	if textCol != nil {
		collateExpr := eg.GenCollateExpr(tbls)
		orderBy = fmt.Sprintf(" ORDER BY %s", collateExpr)
	}

	limit := 1 + rnd(50)
	sql := fmt.Sprintf("SELECT * FROM %s%s LIMIT %d;", quoteIdent(tbl.Name), orderBy, limit)
	return SelectStmt{sql: sql}, nil
}

// GenSelectWithUnaryOp generates a SELECT with unary operators in expressions
func GenSelectWithUnaryOp(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tbls) == 0 {
		return SelectStmt{sql: "SELECT 1;"}, nil
	}

	ctx := NewGenContext(db, lcg, 2)
	eg := NewExprGenerator(ctx)

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

	// Generate unary expression for SELECT list
	unaryExpr := eg.GenUnaryExpr(tbls)

	limit := 1 + rnd(50)
	sql := fmt.Sprintf("SELECT %s FROM %s LIMIT %d;", unaryExpr, quoteIdent(tbl.Name), limit)
	return SelectStmt{sql: sql}, nil
}

// GenSelectWithBinaryOp generates a SELECT with binary operators in expressions
func GenSelectWithBinaryOp(db *sql.DB, lcg *common.LCG) (SelectStmt, error) {
	tbls, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tbls) == 0 {
		return SelectStmt{sql: "SELECT 1;"}, nil
	}

	ctx := NewGenContext(db, lcg, 2)
	eg := NewExprGenerator(ctx)

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

	// Generate binary expression for SELECT list
	binaryExpr := eg.GenBinaryExpr(tbls)

	limit := 1 + rnd(50)
	sql := fmt.Sprintf("SELECT %s FROM %s LIMIT %d;", binaryExpr, quoteIdent(tbl.Name), limit)
	return SelectStmt{sql: sql}, nil
}
