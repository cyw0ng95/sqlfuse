package stmts

import (
	"fmt"
	"sqlsmith-go/internal/generators/turso/helper"
	"sqlsmith-go/internal/generators/turso/types"
	"strings"
)

// Expr provides centralized expression generation for various SQL expressions.
// This module covers expression types from https://github.com/tursodatabase/turso/blob/main/COMPAT.md#expressions
//
// Supported expressions (as per Turso COMPAT.md):
//   - literals
//   - unary operators (+, -, ~, NOT)
//   - binary operators (excluding %, !<, !>)
//   - (expr) - parenthesized expressions
//   - CAST (expr AS type)
//   - COLLATE (partial - custom collations not supported)
//   - (NOT) LIKE
//   - (NOT) GLOB
//   - IS (NOT)
//   - IS (NOT) DISTINCT FROM
//   - (NOT) BETWEEN ... AND ...
//   - CASE WHEN THEN ELSE END
//
// Unsupported expressions (generated for testing but will fail on Turso):
//   - (NOT) REGEXP - not supported
//   - (NOT) MATCH - not supported
//   - (NOT) IN (subquery) - not supported
//   - (NOT) EXISTS (subquery) - not supported
//   - agg() FILTER (WHERE ...) - incorrectly ignored
//   - ... OVER (...) - incorrectly ignored

// GenCastExpr generates a CAST expression: CAST(expr AS type)
func (eg *ExprGenerator) GenCastExpr(tbls []helper.TableInfo) string {
	if len(tbls) == 0 {
		return "CAST(1 AS INTEGER)"
	}

	tbl := tbls[eg.ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		return "CAST(1 AS INTEGER)"
	}

	col := tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]

	// Target types for CAST
	targetTypes := []string{"INTEGER", "TEXT", "REAL", "BLOB"}
	targetType := targetTypes[eg.ctx.Intn(len(targetTypes))]

	return fmt.Sprintf("CAST(%s AS %s)", quoteIdent(col.Name), targetType)
}

// GenBetweenExpr generates a BETWEEN expression: expr (NOT) BETWEEN val1 AND val2
func (eg *ExprGenerator) GenBetweenExpr(tbls []helper.TableInfo, useNot bool) string {
	if len(tbls) == 0 {
		if useNot {
			return "1 NOT BETWEEN 0 AND 10"
		}
		return "1 BETWEEN 0 AND 10"
	}

	tbl := tbls[eg.ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		if useNot {
			return "1 NOT BETWEEN 0 AND 10"
		}
		return "1 BETWEEN 0 AND 10"
	}

	// Find a numeric column for BETWEEN
	var numericCols []helper.ColumnInfo
	for _, col := range tbl.Cols {
		if isNumericType(col.Type) {
			numericCols = append(numericCols, col)
		}
	}

	var col helper.ColumnInfo
	if len(numericCols) > 0 {
		col = numericCols[eg.ctx.Intn(len(numericCols))]
	} else {
		col = tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]
	}

	val1 := types.ValueForType(col.Type, eg.ctx.LCG, col.Name)
	val2 := types.ValueForType(col.Type, eg.ctx.LCG, col.Name)

	notClause := ""
	if useNot {
		notClause = "NOT "
	}

	return fmt.Sprintf("%s %sBETWEEN %s AND %s", quoteIdent(col.Name), notClause, val1, val2)
}

// GenGlobExpr generates a GLOB expression: expr (NOT) GLOB pattern
func (eg *ExprGenerator) GenGlobExpr(tbls []helper.TableInfo, useNot bool) string {
	if len(tbls) == 0 {
		if useNot {
			return "'test' NOT GLOB 'te*'"
		}
		return "'test' GLOB 'te*'"
	}

	tbl := tbls[eg.ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		if useNot {
			return "'test' NOT GLOB 'te*'"
		}
		return "'test' GLOB 'te*'"
	}

	// Find text columns for GLOB
	var textCols []helper.ColumnInfo
	for _, col := range tbl.Cols {
		colTypeUpper := strings.ToUpper(col.Type)
		if strings.Contains(colTypeUpper, "TEXT") || strings.Contains(colTypeUpper, "CHAR") {
			textCols = append(textCols, col)
		}
	}

	var col helper.ColumnInfo
	if len(textCols) > 0 {
		col = textCols[eg.ctx.Intn(len(textCols))]
	} else {
		col = tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]
	}

	// Generate GLOB patterns
	patterns := []string{"*", "?*", "[a-z]*", "*[0-9]", "A*"}
	pattern := patterns[eg.ctx.Intn(len(patterns))]

	notClause := ""
	if useNot {
		notClause = "NOT "
	}

	return fmt.Sprintf("%s %sGLOB '%s'", quoteIdent(col.Name), notClause, pattern)
}

// GenIsDistinctFromExpr generates an IS (NOT) DISTINCT FROM expression
func (eg *ExprGenerator) GenIsDistinctFromExpr(tbls []helper.TableInfo, useNot bool) string {
	if len(tbls) == 0 {
		if useNot {
			return "1 IS NOT DISTINCT FROM 1"
		}
		return "1 IS DISTINCT FROM NULL"
	}

	tbl := tbls[eg.ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		if useNot {
			return "1 IS NOT DISTINCT FROM 1"
		}
		return "1 IS DISTINCT FROM NULL"
	}

	col := tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]

	// For IS DISTINCT FROM, compare with NULL or another value
	var compareVal string
	if eg.ctx.Intn(2) == 0 {
		compareVal = "NULL"
	} else {
		compareVal = types.ValueForType(col.Type, eg.ctx.LCG, col.Name)
	}

	notClause := ""
	if useNot {
		notClause = "NOT "
	}

	return fmt.Sprintf("%s IS %sDISTINCT FROM %s", quoteIdent(col.Name), notClause, compareVal)
}

// GenCollateExpr generates a COLLATE expression: expr COLLATE collation_name
// IMPORTANT: Only default SQLite collations are supported by Turso (BINARY, NOCASE, RTRIM).
// Custom collations are NOT supported per COMPAT.md.
func (eg *ExprGenerator) GenCollateExpr(tbls []helper.TableInfo) string {
	if len(tbls) == 0 {
		return "'text' COLLATE NOCASE"
	}

	tbl := tbls[eg.ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		return "'text' COLLATE NOCASE"
	}

	// Find text columns for COLLATE
	var textCols []helper.ColumnInfo
	for _, col := range tbl.Cols {
		colTypeUpper := strings.ToUpper(col.Type)
		if strings.Contains(colTypeUpper, "TEXT") || strings.Contains(colTypeUpper, "CHAR") {
			textCols = append(textCols, col)
		}
	}

	var col helper.ColumnInfo
	if len(textCols) > 0 {
		col = textCols[eg.ctx.Intn(len(textCols))]
	} else {
		col = tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]
	}

	// Default SQLite collations only (Turso doesn't support custom collations)
	collations := []string{"BINARY", "NOCASE", "RTRIM"}
	collation := collations[eg.ctx.Intn(len(collations))]

	return fmt.Sprintf("%s COLLATE %s", quoteIdent(col.Name), collation)
}

// GenUnaryExpr generates a unary operator expression: +expr, -expr, ~expr, NOT expr
func (eg *ExprGenerator) GenUnaryExpr(tbls []helper.TableInfo) string {
	if len(tbls) == 0 {
		operators := []string{"+", "-", "~", "NOT"}
		op := operators[eg.ctx.Intn(len(operators))]
		if op == "NOT" {
			return "NOT (1 = 1)"
		}
		return fmt.Sprintf("%s1", op)
	}

	tbl := tbls[eg.ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		operators := []string{"+", "-", "~", "NOT"}
		op := operators[eg.ctx.Intn(len(operators))]
		if op == "NOT" {
			return "NOT (1 = 1)"
		}
		return fmt.Sprintf("%s1", op)
	}

	// Find numeric columns for +, -, ~ operators
	var numericCols []helper.ColumnInfo
	for _, col := range tbl.Cols {
		if isNumericType(col.Type) {
			numericCols = append(numericCols, col)
		}
	}

	operators := []string{"+", "-", "~", "NOT"}
	op := operators[eg.ctx.Intn(len(operators))]

	if op == "NOT" {
		// NOT works with boolean expressions
		col := tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]
		val := types.ValueForType(col.Type, eg.ctx.LCG, col.Name)
		return fmt.Sprintf("NOT (%s = %s)", quoteIdent(col.Name), val)
	}

	// +, -, ~ work with numeric values
	if len(numericCols) > 0 {
		col := numericCols[eg.ctx.Intn(len(numericCols))]
		return fmt.Sprintf("%s%s", op, quoteIdent(col.Name))
	}

	// Fallback to literal
	return fmt.Sprintf("%s1", op)
}

// GenBinaryExpr generates a binary operator expression
// IMPORTANT: Excludes unsupported Turso operators per COMPAT.md:
//   - % (modulo) - NOT SUPPORTED
//   - !< (not less than) - NOT SUPPORTED  
//   - !> (not greater than) - NOT SUPPORTED
func (eg *ExprGenerator) GenBinaryExpr(tbls []helper.TableInfo) string {
	if len(tbls) == 0 {
		return "1 + 1"
	}

	tbl := tbls[eg.ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		return "1 + 1"
	}

	// Find numeric columns for binary operators
	var numericCols []helper.ColumnInfo
	for _, col := range tbl.Cols {
		if isNumericType(col.Type) {
			numericCols = append(numericCols, col)
		}
	}

	// Supported binary operators (excluding %, !<, !> which are NOT supported by Turso)
	numericOps := []string{"+", "-", "*", "/", "&", "|", "<<", ">>"}
	comparisonOps := []string{"=", "!=", "<", "<=", ">", ">=", "<>"}
	logicalOps := []string{"AND", "OR"}

	var col1, col2 helper.ColumnInfo
	var op string

	opType := eg.ctx.Intn(3) // 0: numeric, 1: comparison, 2: logical
	switch opType {
	case 0: // Numeric operators
		if len(numericCols) >= 2 {
			col1 = numericCols[eg.ctx.Intn(len(numericCols))]
			col2 = numericCols[eg.ctx.Intn(len(numericCols))]
			op = numericOps[eg.ctx.Intn(len(numericOps))]
			return fmt.Sprintf("%s %s %s", quoteIdent(col1.Name), op, quoteIdent(col2.Name))
		}
		op = numericOps[eg.ctx.Intn(len(numericOps))]
		return fmt.Sprintf("1 %s 2", op)

	case 1: // Comparison operators
		col1 = tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]
		val := types.ValueForType(col1.Type, eg.ctx.LCG, col1.Name)
		op = comparisonOps[eg.ctx.Intn(len(comparisonOps))]
		return fmt.Sprintf("%s %s %s", quoteIdent(col1.Name), op, val)

	default: // Logical operators
		col1 = tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]
		col2 = tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]
		val1 := types.ValueForType(col1.Type, eg.ctx.LCG, col1.Name)
		val2 := types.ValueForType(col2.Type, eg.ctx.LCG, col2.Name)
		op = logicalOps[eg.ctx.Intn(len(logicalOps))]
		return fmt.Sprintf("(%s = %s) %s (%s = %s)", quoteIdent(col1.Name), val1, op, quoteIdent(col2.Name), val2)
	}
}

// GenParenthesizedExpr generates a parenthesized expression: (expr)
func (eg *ExprGenerator) GenParenthesizedExpr(tbls []helper.TableInfo) string {
	if len(tbls) == 0 {
		return "(1 + 1)"
	}

	tbl := tbls[eg.ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		return "(1 + 1)"
	}

	col := tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]
	val := types.ValueForType(col.Type, eg.ctx.LCG, col.Name)

	return fmt.Sprintf("(%s = %s)", quoteIdent(col.Name), val)
}

// GenIsNullExpr generates IS (NOT) NULL expression
func (eg *ExprGenerator) GenIsNullExpr(tbls []helper.TableInfo, useNot bool) string {
	if len(tbls) == 0 {
		if useNot {
			return "1 IS NOT NULL"
		}
		return "NULL IS NULL"
	}

	tbl := tbls[eg.ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		if useNot {
			return "1 IS NOT NULL"
		}
		return "NULL IS NULL"
	}

	col := tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]

	if useNot {
		return fmt.Sprintf("%s IS NOT NULL", quoteIdent(col.Name))
	}
	return fmt.Sprintf("%s IS NULL", quoteIdent(col.Name))
}

// GenRegexpExpr generates a REGEXP expression: expr (NOT) REGEXP pattern
// Note: REGEXP is not supported by Turso according to COMPAT.md
func (eg *ExprGenerator) GenRegexpExpr(tbls []helper.TableInfo, useNot bool) string {
	if len(tbls) == 0 {
		if useNot {
			return "'test' NOT REGEXP '^t.*'"
		}
		return "'test' REGEXP '^t.*'"
	}

	tbl := tbls[eg.ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		if useNot {
			return "'test' NOT REGEXP '^t.*'"
		}
		return "'test' REGEXP '^t.*'"
	}

	// Find text columns for REGEXP
	var textCols []helper.ColumnInfo
	for _, col := range tbl.Cols {
		colTypeUpper := strings.ToUpper(col.Type)
		if strings.Contains(colTypeUpper, "TEXT") || strings.Contains(colTypeUpper, "CHAR") {
			textCols = append(textCols, col)
		}
	}

	var col helper.ColumnInfo
	if len(textCols) > 0 {
		col = textCols[eg.ctx.Intn(len(textCols))]
	} else {
		col = tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]
	}

	// Generate regex patterns
	// Note: Using explicit character classes for better SQLite compatibility
	patterns := []string{"^[a-z]+$", ".*[0-9].*", "^test", "[A-Z]+", "[0-9]+"}
	pattern := patterns[eg.ctx.Intn(len(patterns))]

	notClause := ""
	if useNot {
		notClause = "NOT "
	}

	return fmt.Sprintf("%s %sREGEXP '%s'", quoteIdent(col.Name), notClause, pattern)
}

// GenMatchExpr generates a MATCH expression: expr (NOT) MATCH pattern
// Note: MATCH is not supported by Turso according to COMPAT.md
func (eg *ExprGenerator) GenMatchExpr(tbls []helper.TableInfo, useNot bool) string {
	if len(tbls) == 0 {
		if useNot {
			return "'test' NOT MATCH 'pattern'"
		}
		return "'test' MATCH 'pattern'"
	}

	tbl := tbls[eg.ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		if useNot {
			return "'test' NOT MATCH 'pattern'"
		}
		return "'test' MATCH 'pattern'"
	}

	// Find text columns for MATCH
	var textCols []helper.ColumnInfo
	for _, col := range tbl.Cols {
		colTypeUpper := strings.ToUpper(col.Type)
		if strings.Contains(colTypeUpper, "TEXT") || strings.Contains(colTypeUpper, "CHAR") {
			textCols = append(textCols, col)
		}
	}

	var col helper.ColumnInfo
	if len(textCols) > 0 {
		col = textCols[eg.ctx.Intn(len(textCols))]
	} else {
		col = tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]
	}

	// Generate match patterns
	patterns := []string{"pattern", "test", "value", "match"}
	pattern := patterns[eg.ctx.Intn(len(patterns))]

	notClause := ""
	if useNot {
		notClause = "NOT "
	}

	return fmt.Sprintf("%s %sMATCH '%s'", quoteIdent(col.Name), notClause, pattern)
}

// GenInSubqueryExpr generates an IN (subquery) expression: expr (NOT) IN (SELECT ...)
// Note: IN (subquery) is not supported by Turso according to COMPAT.md
func (eg *ExprGenerator) GenInSubqueryExpr(tbls []helper.TableInfo, useNot bool) string {
	if len(tbls) == 0 {
		if useNot {
			return "1 NOT IN (SELECT 1)"
		}
		return "1 IN (SELECT 1)"
	}

	tbl := tbls[eg.ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		if useNot {
			return "1 NOT IN (SELECT 1)"
		}
		return "1 IN (SELECT 1)"
	}

	col := tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]

	// Generate a simple subquery
	subquery := fmt.Sprintf("SELECT %s FROM %s LIMIT 5", quoteIdent(col.Name), quoteIdent(tbl.Name))

	notClause := ""
	if useNot {
		notClause = "NOT "
	}

	return fmt.Sprintf("%s %sIN (%s)", quoteIdent(col.Name), notClause, subquery)
}

// GenExistsSubqueryExpr generates an EXISTS (subquery) expression: (NOT) EXISTS (SELECT ...)
// Note: EXISTS (subquery) is not supported by Turso according to COMPAT.md
func (eg *ExprGenerator) GenExistsSubqueryExpr(tbls []helper.TableInfo, useNot bool) string {
	if len(tbls) == 0 {
		if useNot {
			return "NOT EXISTS (SELECT 1)"
		}
		return "EXISTS (SELECT 1)"
	}

	tbl := tbls[eg.ctx.Intn(len(tbls))]

	// Generate a simple subquery with WHERE clause if possible
	var subquery string
	if len(tbl.Cols) > 0 {
		col := tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]
		val := types.ValueForType(col.Type, eg.ctx.LCG, col.Name)
		subquery = fmt.Sprintf("SELECT 1 FROM %s WHERE %s = %s", quoteIdent(tbl.Name), quoteIdent(col.Name), val)
	} else {
		subquery = fmt.Sprintf("SELECT 1 FROM %s", quoteIdent(tbl.Name))
	}

	notClause := ""
	if useNot {
		notClause = "NOT "
	}

	return fmt.Sprintf("%sEXISTS (%s)", notClause, subquery)
}

// GenFilterExpr generates an aggregate function with FILTER clause: agg() FILTER (WHERE ...)
// Note: FILTER is not supported by Turso according to COMPAT.md (incorrectly ignored)
func (eg *ExprGenerator) GenFilterExpr(tbls []helper.TableInfo) string {
	if len(tbls) == 0 {
		return "COUNT(*) FILTER (WHERE 1 = 1)"
	}

	tbl := tbls[eg.ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		return "COUNT(*) FILTER (WHERE 1 = 1)"
	}

	// Choose an aggregate function
	aggregates := []string{"COUNT", "SUM", "AVG", "MIN", "MAX"}
	agg := aggregates[eg.ctx.Intn(len(aggregates))]

	// Generate aggregate argument
	var aggArg string
	if agg == "COUNT" && eg.ctx.Intn(2) == 0 {
		aggArg = "*"
	} else {
		// Use a numeric column for SUM/AVG or any column for others
		var cols []helper.ColumnInfo
		if agg == "SUM" || agg == "AVG" {
			// Prefer numeric columns
			for _, col := range tbl.Cols {
				if isNumericType(col.Type) {
					cols = append(cols, col)
				}
			}
		}
		if len(cols) == 0 {
			cols = tbl.Cols
		}
		col := cols[eg.ctx.Intn(len(cols))]
		aggArg = quoteIdent(col.Name)
	}

	// Generate WHERE condition for FILTER
	filterCol := tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]
	filterVal := types.ValueForType(filterCol.Type, eg.ctx.LCG, filterCol.Name)

	return fmt.Sprintf("%s(%s) FILTER (WHERE %s = %s)", agg, aggArg, quoteIdent(filterCol.Name), filterVal)
}

// GenOverExpr generates a window function with OVER clause: func() OVER (...)
// Note: OVER is not supported by Turso according to COMPAT.md (incorrectly ignored)
func (eg *ExprGenerator) GenOverExpr(tbls []helper.TableInfo) string {
	if len(tbls) == 0 {
		return "ROW_NUMBER() OVER (ORDER BY 1)"
	}

	tbl := tbls[eg.ctx.Intn(len(tbls))]
	if len(tbl.Cols) == 0 {
		return "ROW_NUMBER() OVER (ORDER BY 1)"
	}

	// Choose a window function
	windowFuncs := []string{"ROW_NUMBER", "RANK", "DENSE_RANK", "NTILE"}
	fn := windowFuncs[eg.ctx.Intn(len(windowFuncs))]

	// Generate function argument (NTILE requires argument)
	var fnArg string
	if fn == "NTILE" {
		fnArg = fmt.Sprintf("%d", 2+eg.ctx.Intn(10))
	}

	// Generate OVER clause
	col := tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]

	// Optionally add PARTITION BY
	var overClause string
	if eg.ctx.Intn(2) == 0 && len(tbl.Cols) > 1 {
		partCol := tbl.Cols[eg.ctx.Intn(len(tbl.Cols))]
		overClause = fmt.Sprintf("PARTITION BY %s ORDER BY %s", quoteIdent(partCol.Name), quoteIdent(col.Name))
	} else {
		overClause = fmt.Sprintf("ORDER BY %s", quoteIdent(col.Name))
	}

	return fmt.Sprintf("%s(%s) OVER (%s)", fn, fnArg, overClause)
}

// GenRandomExpr generates a random expression from all available expression types
func (eg *ExprGenerator) GenRandomExpr(tbls []helper.TableInfo) string {
	if len(tbls) == 0 {
		return "1"
	}

	// Choose random expression type
	// Note: IS DISTINCT FROM excluded as it's not supported by the SQLite ANTLR parser
	// Expression types: 0-8 are supported, 9-14 are unsupported (for testing)
	const numExprTypes = 15
	exprType := eg.ctx.Intn(numExprTypes)

	switch exprType {
	case 0:
		return eg.GenCastExpr(tbls)
	case 1:
		return eg.GenBetweenExpr(tbls, eg.ctx.Intn(2) == 0)
	case 2:
		return eg.GenGlobExpr(tbls, eg.ctx.Intn(2) == 0)
	case 3:
		return eg.GenCollateExpr(tbls)
	case 4:
		return eg.GenUnaryExpr(tbls)
	case 5:
		return eg.GenBinaryExpr(tbls)
	case 6:
		return eg.GenParenthesizedExpr(tbls)
	case 7:
		return eg.GenIsNullExpr(tbls, eg.ctx.Intn(2) == 0)
	case 8:
		// Existing CASE expression
		return eg.genCaseExpr(tbls)
	case 9:
		// REGEXP - unsupported by Turso
		return eg.GenRegexpExpr(tbls, eg.ctx.Intn(2) == 0)
	case 10:
		// MATCH - unsupported by Turso
		return eg.GenMatchExpr(tbls, eg.ctx.Intn(2) == 0)
	case 11:
		// IN (subquery) - unsupported by Turso
		return eg.GenInSubqueryExpr(tbls, eg.ctx.Intn(2) == 0)
	case 12:
		// EXISTS (subquery) - unsupported by Turso
		return eg.GenExistsSubqueryExpr(tbls, eg.ctx.Intn(2) == 0)
	case 13:
		// FILTER clause - unsupported by Turso (incorrectly ignored)
		return eg.GenFilterExpr(tbls)
	default:
		// OVER clause - unsupported by Turso (incorrectly ignored)
		return eg.GenOverExpr(tbls)
	}
}
