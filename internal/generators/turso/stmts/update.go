package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/generators/turso/helper"
	"sqlsmith-go/internal/generators/turso/types"
	"strings"
)

// UpdateGenerator is a StmtGenerator for UPDATE statements.
type UpdateGenerator struct{}

// Generate implements StmtGenerator for UPDATE statements.
func (g *UpdateGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return genUpdateInternal(ctx.LCG)
}

// CanGenerate implements StmtGenerator. UPDATE can always be generated (creates synthetic tables).
func (g *UpdateGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// UpdateStmt represents an UPDATE statement.
type UpdateStmt struct {
	sql string
}

func (s *UpdateStmt) SQL() string  { return s.sql }
func (s *UpdateStmt) Type() string { return "update" }

// GenUpdate generates an UPDATE statement using actual schema information.
// It supports complex SET expressions and WHERE clauses.
func GenUpdate(db *sql.DB, lcg *common.LCG) (Stmt, error) {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}

	// Try to get actual tables from schema
	tables, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tables) == 0 {
		// Fallback to simple update without schema
		return genUpdateFallback(lcg), nil
	}

	rnd := lcg.Intn
	tbl := tables[rnd(len(tables))]
	if len(tbl.Cols) == 0 {
		return genUpdateFallback(lcg), nil
	}

	// Filter out id column from SET clause
	setCols := []helper.ColumnInfo{}
	for _, c := range tbl.Cols {
		if !strings.EqualFold(c.Name, "id") {
			setCols = append(setCols, c)
		}
	}

	if len(setCols) == 0 {
		return genUpdateFallback(lcg), nil
	}

	// Choose 1..min(3, len(setCols)) columns to update
	maxCols := len(setCols)
	if maxCols > 3 {
		maxCols = 3
	}
	nCols := 1 + rnd(maxCols)
	if nCols > len(setCols) {
		nCols = len(setCols)
	}

	selected := make(map[int]struct{}, nCols)
	setExprs := make([]string, 0, nCols)
	for len(setExprs) < nCols {
		idx := rnd(len(setCols))
		if _, ok := selected[idx]; ok {
			continue
		}
		selected[idx] = struct{}{}
		col := setCols[idx]
		
		// Generate different types of SET expressions
		choice := rnd(10)
		var expr string
		switch choice {
		case 0:
			// Simple value assignment
			val := types.ValueForType(col.Type, lcg, col.Name)
			expr = fmt.Sprintf("%s = %s", quoteIdent(col.Name), val)
		case 1:
			// Arithmetic expression (for numeric columns)
			if isNumericType(col.Type) {
				val := types.ValueForType(col.Type, lcg, col.Name)
				ops := []string{"+", "-", "*", "/"}
				op := ops[rnd(len(ops))]
				expr = fmt.Sprintf("%s = %s %s %s", quoteIdent(col.Name), quoteIdent(col.Name), op, val)
			} else {
				val := types.ValueForType(col.Type, lcg, col.Name)
				expr = fmt.Sprintf("%s = %s", quoteIdent(col.Name), val)
			}
		case 2:
			// CASE expression
			val1 := types.ValueForType(col.Type, lcg, col.Name)
			val2 := types.ValueForType(col.Type, lcg, col.Name)
			expr = fmt.Sprintf("%s = CASE WHEN %s IS NULL THEN %s ELSE %s END", 
				quoteIdent(col.Name), quoteIdent(col.Name), val1, val2)
		case 3:
			// COALESCE expression
			val := types.ValueForType(col.Type, lcg, col.Name)
			expr = fmt.Sprintf("%s = COALESCE(%s, %s)", quoteIdent(col.Name), quoteIdent(col.Name), val)
		case 4:
			// CAST expression
			targetTypes := []string{"TEXT", "INTEGER", "REAL"}
			targetType := targetTypes[rnd(len(targetTypes))]
			val := types.ValueForType(col.Type, lcg, col.Name)
			expr = fmt.Sprintf("%s = CAST(%s AS %s)", quoteIdent(col.Name), val, targetType)
		case 5:
			// NULL assignment
			expr = fmt.Sprintf("%s = NULL", quoteIdent(col.Name))
		case 6:
			// String concatenation (for text columns)
			if isTextType(col.Type) {
				val := types.ValueForType(col.Type, lcg, col.Name)
				expr = fmt.Sprintf("%s = %s || %s", quoteIdent(col.Name), quoteIdent(col.Name), val)
			} else {
				val := types.ValueForType(col.Type, lcg, col.Name)
				expr = fmt.Sprintf("%s = %s", quoteIdent(col.Name), val)
			}
		case 7:
			// UPPER/LOWER for text columns
			if isTextType(col.Type) {
				fn := "UPPER"
				if rnd(2) == 0 {
					fn = "LOWER"
				}
				expr = fmt.Sprintf("%s = %s(%s)", quoteIdent(col.Name), fn, quoteIdent(col.Name))
			} else {
				val := types.ValueForType(col.Type, lcg, col.Name)
				expr = fmt.Sprintf("%s = %s", quoteIdent(col.Name), val)
			}
		case 8:
			// ABS/ROUND for numeric columns
			if isNumericType(col.Type) {
				fn := "ABS"
				if rnd(2) == 0 {
					fn = "ROUND"
				}
				expr = fmt.Sprintf("%s = %s(%s)", quoteIdent(col.Name), fn, quoteIdent(col.Name))
			} else {
				val := types.ValueForType(col.Type, lcg, col.Name)
				expr = fmt.Sprintf("%s = %s", quoteIdent(col.Name), val)
			}
		default:
			// Default to simple assignment
			val := types.ValueForType(col.Type, lcg, col.Name)
			expr = fmt.Sprintf("%s = %s", quoteIdent(col.Name), val)
		}
		setExprs = append(setExprs, expr)
	}

	// Build WHERE clause with complex conditions
	where := ""
	if rnd(4) > 0 { // 75% chance of WHERE clause
		whereClauses := make([]string, 0, 3)
		numConditions := 1 + rnd(3)
		
		for i := 0; i < numConditions && i < len(tbl.Cols); i++ {
			col := tbl.Cols[rnd(len(tbl.Cols))]
			condChoice := rnd(8)
			var cond string
			
			switch condChoice {
			case 0:
				// Simple equality
				val := types.ValueForType(col.Type, lcg, col.Name)
				cond = fmt.Sprintf("%s = %s", quoteIdent(col.Name), val)
			case 1:
				// Inequality
				val := types.ValueForType(col.Type, lcg, col.Name)
				ops := []string{"<", "<=", ">", ">=", "!=", "<>"}
				op := ops[rnd(len(ops))]
				cond = fmt.Sprintf("%s %s %s", quoteIdent(col.Name), op, val)
			case 2:
				// IS NULL / IS NOT NULL
				if rnd(2) == 0 {
					cond = fmt.Sprintf("%s IS NULL", quoteIdent(col.Name))
				} else {
					cond = fmt.Sprintf("%s IS NOT NULL", quoteIdent(col.Name))
				}
			case 3:
				// LIKE (for text columns)
				if isTextType(col.Type) {
					patterns := []string{"'%test%'", "'test%'", "'%test'", "'_test%'"}
					pattern := patterns[rnd(len(patterns))]
					cond = fmt.Sprintf("%s LIKE %s", quoteIdent(col.Name), pattern)
				} else {
					val := types.ValueForType(col.Type, lcg, col.Name)
					cond = fmt.Sprintf("%s = %s", quoteIdent(col.Name), val)
				}
			case 4:
				// BETWEEN (for numeric columns)
				if isNumericType(col.Type) {
					val1 := types.ValueForType(col.Type, lcg, col.Name)
					val2 := types.ValueForType(col.Type, lcg, col.Name)
					cond = fmt.Sprintf("%s BETWEEN %s AND %s", quoteIdent(col.Name), val1, val2)
				} else {
					val := types.ValueForType(col.Type, lcg, col.Name)
					cond = fmt.Sprintf("%s = %s", quoteIdent(col.Name), val)
				}
			case 5:
				// IN clause
				vals := make([]string, 2+rnd(4))
				for j := range vals {
					vals[j] = types.ValueForType(col.Type, lcg, col.Name)
				}
				cond = fmt.Sprintf("%s IN (%s)", quoteIdent(col.Name), strings.Join(vals, ", "))
			case 6:
				// Comparison with expression
				if isNumericType(col.Type) {
					val := types.ValueForType(col.Type, lcg, col.Name)
					cond = fmt.Sprintf("%s > %s * 2", quoteIdent(col.Name), val)
				} else {
					val := types.ValueForType(col.Type, lcg, col.Name)
					cond = fmt.Sprintf("%s = %s", quoteIdent(col.Name), val)
				}
			default:
				// Simple equality as default
				val := types.ValueForType(col.Type, lcg, col.Name)
				cond = fmt.Sprintf("%s = %s", quoteIdent(col.Name), val)
			}
			whereClauses = append(whereClauses, cond)
		}
		
		// Combine conditions with AND/OR
		if len(whereClauses) > 0 {
			if len(whereClauses) == 1 {
				where = " WHERE " + whereClauses[0]
			} else {
				connector := "AND"
				if rnd(4) == 0 { // 25% chance of OR
					connector = "OR"
				}
				where = " WHERE " + strings.Join(whereClauses, " "+connector+" ")
			}
		}
	}

	sql := fmt.Sprintf("UPDATE %s SET %s%s;", quoteIdent(tbl.Name), strings.Join(setExprs, ", "), where)
	return &UpdateStmt{sql: sql}, nil
}

// genUpdateFallback generates a simple UPDATE without schema information
func genUpdateFallback(lcg *common.LCG) *UpdateStmt {
	tbl := fmt.Sprintf("tbl_%d", lcg.Uint64()%1000000)
	n := 1 + lcg.Intn(3)
	sets := make([]string, 0, n)
	for i := 0; i < n; i++ {
		col := fmt.Sprintf("col%d", 1+lcg.Intn(6))
		v := genLiteral(lcg)
		sets = append(sets, fmt.Sprintf("\"%s\" = %s", col, v))
	}

	where := ""
	if lcg.Intn(2) == 0 {
		col := fmt.Sprintf("col%d", 1+lcg.Intn(6))
		where = fmt.Sprintf(" WHERE \"%s\" = %s", col, genLiteral(lcg))
	}

	sql := fmt.Sprintf("UPDATE \"%s\" SET %s%s;", tbl, join(sets, ", "), where)
	return &UpdateStmt{sql: sql}
}

func genLiteral(lcg *common.LCG) string {
	choice := lcg.Intn(5)
	switch choice {
	case 0:
		return types.IntLiteral(lcg, "")
	case 1:
		return types.StringLiteral(lcg, "")
	case 2:
		return types.RealLiteral(lcg)
	case 3:
		return "NULL"
	default:
		return types.BlobLiteral(lcg)
	}
}

func join(a []string, sep string) string {
	if len(a) == 0 {
		return ""
	}
	res := a[0]
	for i := 1; i < len(a); i++ {
		res += sep + a[i]
	}
	return res
}
