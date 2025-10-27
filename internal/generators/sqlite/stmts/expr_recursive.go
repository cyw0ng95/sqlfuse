package stmts

import (
	"fmt"
	"sqlsmith-go/internal/generators/sqlite/helper"
	"sqlsmith-go/internal/generators/sqlite/types"
	"strings"
)

// ExprGenerator provides recursive expression generation capabilities.
type ExprGenerator struct {
	ctx *GenContext
}

// NewExprGenerator creates a new expression generator with the given context.
func NewExprGenerator(ctx *GenContext) *ExprGenerator {
	return &ExprGenerator{ctx: ctx}
}

// GenWhereExpr generates a WHERE clause expression, potentially recursive.
// Returns the expression without the "WHERE" keyword.
func (eg *ExprGenerator) GenWhereExpr(tbls []helper.TableInfo) string {
	if len(tbls) == 0 {
		return ""
	}

	// Choose whether to generate a simple or complex expression
	if !eg.ctx.CanRecurse() || eg.ctx.Intn(3) == 0 {
		// Generate simple condition
		return eg.genSimpleCondition(tbls)
	}

	// Generate complex condition with AND/OR
	numConditions := 2 + eg.ctx.Intn(2) // 2-3 conditions
	conditions := make([]string, 0, numConditions)

	for i := 0; i < numConditions; i++ {
		if eg.ctx.CanRecurse() && eg.ctx.Intn(4) == 0 {
			// Occasionally nest deeper
			subCtx := eg.ctx.Descend()
			subGen := NewExprGenerator(subCtx)
			subExpr := subGen.GenWhereExpr(tbls)
			if subExpr != "" {
				conditions = append(conditions, fmt.Sprintf("(%s)", subExpr))
			}
		} else {
			// Generate simple condition
			cond := eg.genSimpleCondition(tbls)
			if cond != "" {
				conditions = append(conditions, cond)
			}
		}
	}

	if len(conditions) == 0 {
		return ""
	}

	// Combine with AND or OR
	connector := " AND "
	if eg.ctx.Intn(2) == 0 {
		connector = " OR "
	}

	return strings.Join(conditions, connector)
}

// genSimpleCondition generates a simple comparison condition.
func (eg *ExprGenerator) genSimpleCondition(tbls []helper.TableInfo) string {
	if len(tbls) == 0 {
		return ""
	}

	tbl := tbls[eg.ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		return ""
	}

	col := tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]
	val := types.ValueForType(col.Type, eg.ctx.LCG, col.Name)

	// Choose operator
	operators := []string{"=", ">", "<", ">=", "<=", "!="}
	if strings.Contains(strings.ToUpper(col.Type), "TEXT") || strings.Contains(strings.ToUpper(col.Type), "CHAR") {
		operators = []string{"=", "!="}
	}
	op := operators[eg.ctx.Intn(len(operators))]

	return fmt.Sprintf("%s %s %s", quoteIdent(col.Name), op, val)
}

// GenSelectExpr generates a SELECT list expression, potentially with nested expressions.
func (eg *ExprGenerator) GenSelectExpr(tbls []helper.TableInfo, maxExprs int) []string {
	if len(tbls) == 0 || maxExprs <= 0 {
		return []string{"1"}
	}

	exprs := make([]string, 0, maxExprs)
	numExprs := 1 + eg.ctx.Intn(maxExprs)

	for i := 0; i < numExprs; i++ {
		if eg.ctx.CanRecurse() && eg.ctx.Intn(3) == 0 {
			// Generate a complex expression (CASE, arithmetic, etc.)
			expr := eg.genComplexExpr(tbls)
			if expr != "" {
				exprs = append(exprs, expr)
			}
		} else {
			// Simple column reference
			tbl := tbls[eg.ctx.Intn(len(tbls))]
			if len(tbl.Cols) > 0 {
				col := tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]
				exprs = append(exprs, quoteIdent(col.Name))
			}
		}
	}

	if len(exprs) == 0 {
		return []string{"1"}
	}

	return exprs
}

// genComplexExpr generates a complex expression like CASE, arithmetic, or function calls.
func (eg *ExprGenerator) genComplexExpr(tbls []helper.TableInfo) string {
	if len(tbls) == 0 {
		return ""
	}

	exprType := eg.ctx.Intn(3)
	switch exprType {
	case 0:
		// CASE expression
		return eg.genCaseExpr(tbls)
	case 1:
		// Arithmetic expression
		return eg.genArithmeticExpr(tbls)
	default:
		// Function call
		return eg.genFunctionExpr(tbls)
	}
}

// genCaseExpr generates a CASE expression.
func (eg *ExprGenerator) genCaseExpr(tbls []helper.TableInfo) string {
	tbl := tbls[eg.ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		return ""
	}

	col := tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]
	val := types.ValueForType(col.Type, eg.ctx.LCG, col.Name)

	// Simple CASE expression
	return fmt.Sprintf("CASE WHEN %s = %s THEN 1 ELSE 0 END",
		quoteIdent(col.Name), val)
}

// genArithmeticExpr generates an arithmetic expression.
func (eg *ExprGenerator) genArithmeticExpr(tbls []helper.TableInfo) string {
	// Find numeric columns
	var numericCols []helper.ColumnInfo
	for _, tbl := range tbls {
		for _, col := range tbl.Cols {
			if isNumericType(col.Type) {
				numericCols = append(numericCols, col)
			}
		}
	}

	if len(numericCols) == 0 {
		return "1 + 1"
	}

	col := numericCols[eg.ctx.Intn(len(numericCols))]
	operators := []string{"+", "-", "*"}
	op := operators[eg.ctx.Intn(len(operators))]
	val := eg.ctx.Intn(100) + 1

	return fmt.Sprintf("%s %s %d", quoteIdent(col.Name), op, val)
}

// genFunctionExpr generates a function call expression.
func (eg *ExprGenerator) genFunctionExpr(tbls []helper.TableInfo) string {
	functions := []string{"LENGTH", "UPPER", "LOWER", "ABS", "ROUND"}
	fn := functions[eg.ctx.Intn(len(functions))]

	tbl := tbls[eg.ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		return fmt.Sprintf("%s('test')", fn)
	}

	col := tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]
	return fmt.Sprintf("%s(%s)", fn, quoteIdent(col.Name))
}

// GenSubquery generates a SELECT subquery, potentially recursive.
func (eg *ExprGenerator) GenSubquery(tbls []helper.TableInfo) string {
	if !eg.ctx.CanRecurse() || len(tbls) == 0 {
		return "SELECT 1"
	}

	tbl := tbls[eg.ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		return fmt.Sprintf("SELECT * FROM %s LIMIT 1", quoteIdent(tbl.Name))
	}

	// Select a few columns
	numCols := 1 + eg.ctx.Intn(2) // 1-2 columns
	cols := make([]string, 0, numCols)
	used := make(map[int]bool)

	for i := 0; i < numCols && i < len(tbl.Cols); i++ {
		idx := eg.ctx.Intn(len(tbl.Cols))
		if !used[idx] {
			used[idx] = true
			cols = append(cols, quoteIdent(tbl.Cols[idx].Name))
		}
	}

	if len(cols) == 0 {
		cols = []string{"*"}
	}

	// Optionally add WHERE clause
	where := ""
	if eg.ctx.CanRecurse() && eg.ctx.Intn(2) == 0 {
		subCtx := eg.ctx.Descend()
		subGen := NewExprGenerator(subCtx)
		whereExpr := subGen.GenWhereExpr([]helper.TableInfo{tbl})
		if whereExpr != "" {
			where = fmt.Sprintf(" WHERE %s", whereExpr)
		}
	}

	limit := 1 + eg.ctx.Intn(10)
	return fmt.Sprintf("SELECT %s FROM %s%s LIMIT %d",
		strings.Join(cols, ", "), quoteIdent(tbl.Name), where, limit)
}
