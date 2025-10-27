package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/generators/turso/helper"
	"sqlsmith-go/internal/generators/turso/types"
	"strings"
)

// DeleteGenerator is a StmtGenerator for DELETE statements.
type DeleteGenerator struct{}

// Generate implements StmtGenerator for DELETE statements.
func (g *DeleteGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return genDeleteInternal(ctx.LCG)
}

// CanGenerate implements StmtGenerator. DELETE can always be generated (creates synthetic tables).
func (g *DeleteGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// DeleteStmt represents a DELETE statement.
type DeleteStmt struct {
	sql string
}

func (s *DeleteStmt) SQL() string  { return s.sql }
func (s *DeleteStmt) Type() string { return "delete" }

// GenDelete generates a DELETE statement using actual schema information.
// It supports complex WHERE clauses with multiple conditions.
func GenDelete(db *sql.DB, lcg *common.LCG) (Stmt, error) {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}

	// Try to get actual tables from schema
	tables, err := helper.GetAllTablesAndCols(db)
	if err != nil || len(tables) == 0 {
		// Fallback to simple delete without schema
		return genDeleteFallback(lcg), nil
	}

	rnd := lcg.Intn
	tbl := tables[rnd(len(tables))]
	
	// Build WHERE clause with complex conditions
	where := ""
	if len(tbl.Cols) > 0 && rnd(5) > 0 { // 80% chance of WHERE clause
		whereClauses := make([]string, 0, 4)
		numConditions := 1 + rnd(4) // 1-4 conditions
		
		for i := 0; i < numConditions && i < len(tbl.Cols); i++ {
			col := tbl.Cols[rnd(len(tbl.Cols))]
			condChoice := rnd(12)
			var cond string
			
			switch condChoice {
			case 0:
				// Simple equality
				val := types.ValueForType(col.Type, lcg, col.Name)
				cond = fmt.Sprintf("%s = %s", quoteIdent(col.Name), val)
			case 1:
				// Less than
				val := types.ValueForType(col.Type, lcg, col.Name)
				cond = fmt.Sprintf("%s < %s", quoteIdent(col.Name), val)
			case 2:
				// Greater than
				val := types.ValueForType(col.Type, lcg, col.Name)
				cond = fmt.Sprintf("%s > %s", quoteIdent(col.Name), val)
			case 3:
				// IS NULL
				cond = fmt.Sprintf("%s IS NULL", quoteIdent(col.Name))
			case 4:
				// IS NOT NULL
				cond = fmt.Sprintf("%s IS NOT NULL", quoteIdent(col.Name))
			case 5:
				// LIKE (for text columns)
				if isTextType(col.Type) {
					patterns := []string{"'%old%'", "'temp%'", "'%_deleted'", "'test%'"}
					pattern := patterns[rnd(len(patterns))]
					cond = fmt.Sprintf("%s LIKE %s", quoteIdent(col.Name), pattern)
				} else {
					val := types.ValueForType(col.Type, lcg, col.Name)
					cond = fmt.Sprintf("%s = %s", quoteIdent(col.Name), val)
				}
			case 6:
				// NOT LIKE
				if isTextType(col.Type) {
					cond = fmt.Sprintf("%s NOT LIKE '%%keep%%'", quoteIdent(col.Name))
				} else {
					val := types.ValueForType(col.Type, lcg, col.Name)
					cond = fmt.Sprintf("%s != %s", quoteIdent(col.Name), val)
				}
			case 7:
				// BETWEEN (for numeric columns)
				if isNumericType(col.Type) {
					val1 := types.ValueForType(col.Type, lcg, col.Name)
					val2 := types.ValueForType(col.Type, lcg, col.Name)
					cond = fmt.Sprintf("%s BETWEEN %s AND %s", quoteIdent(col.Name), val1, val2)
				} else {
					val := types.ValueForType(col.Type, lcg, col.Name)
					cond = fmt.Sprintf("%s = %s", quoteIdent(col.Name), val)
				}
			case 8:
				// NOT BETWEEN
				if isNumericType(col.Type) {
					val1 := types.ValueForType(col.Type, lcg, col.Name)
					val2 := types.ValueForType(col.Type, lcg, col.Name)
					cond = fmt.Sprintf("%s NOT BETWEEN %s AND %s", quoteIdent(col.Name), val1, val2)
				} else {
					val := types.ValueForType(col.Type, lcg, col.Name)
					cond = fmt.Sprintf("%s != %s", quoteIdent(col.Name), val)
				}
			case 9:
				// IN clause
				vals := make([]string, 2+rnd(5))
				for j := range vals {
					vals[j] = types.ValueForType(col.Type, lcg, col.Name)
				}
				cond = fmt.Sprintf("%s IN (%s)", quoteIdent(col.Name), strings.Join(vals, ", "))
			case 10:
				// NOT IN clause
				vals := make([]string, 2+rnd(3))
				for j := range vals {
					vals[j] = types.ValueForType(col.Type, lcg, col.Name)
				}
				cond = fmt.Sprintf("%s NOT IN (%s)", quoteIdent(col.Name), strings.Join(vals, ", "))
			case 11:
				// Complex expression (for numeric columns)
				if isNumericType(col.Type) {
					val := types.ValueForType(col.Type, lcg, col.Name)
					ops := []string{" * 2 <", " / 2 >", " + 10 =", " - 5 !="}
					op := ops[rnd(len(ops))]
					cond = fmt.Sprintf("%s%s %s", quoteIdent(col.Name), op, val)
				} else {
					val := types.ValueForType(col.Type, lcg, col.Name)
					cond = fmt.Sprintf("%s = %s", quoteIdent(col.Name), val)
				}
			default:
				// Default to simple equality
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
				// Mix AND and OR operators for more complexity
				if rnd(3) == 0 { // 33% chance of all OR
					where = " WHERE " + strings.Join(whereClauses, " OR ")
				} else if rnd(3) == 0 { // 33% chance of mixed
					// Create mixed AND/OR conditions
					combined := whereClauses[0]
					for i := 1; i < len(whereClauses); i++ {
						if rnd(2) == 0 {
							combined += " AND " + whereClauses[i]
						} else {
							combined += " OR " + whereClauses[i]
						}
					}
					where = " WHERE " + combined
				} else { // Default to all AND
					where = " WHERE " + strings.Join(whereClauses, " AND ")
				}
			}
		}
	}

	// Optionally add LIMIT clause
	limitClause := ""
	if rnd(3) == 0 { // 33% chance of LIMIT
		limitClause = fmt.Sprintf(" LIMIT %d", 1+rnd(100))
	}

	sql := fmt.Sprintf("DELETE FROM %s%s%s;", quoteIdent(tbl.Name), where, limitClause)
	return &DeleteStmt{sql: sql}, nil
}

// genDeleteFallback generates a simple DELETE without schema information
func genDeleteFallback(lcg *common.LCG) *DeleteStmt {
	tbl := fmt.Sprintf("tbl_%d", lcg.Uint64()%1000000)
	where := ""
	if lcg.Intn(2) == 0 {
		col := fmt.Sprintf("col%d", 1+lcg.Intn(6))
		where = fmt.Sprintf(" WHERE \"%s\" = %d", col, lcg.Intn(100))
	}

	sql := fmt.Sprintf("DELETE FROM \"%s\"%s;", tbl, where)
	return &DeleteStmt{sql: sql}
}
