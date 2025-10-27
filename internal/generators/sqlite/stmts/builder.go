package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/generators/sqlite/helper"
	"strings"
)

// SelectBuilder provides a fluent interface for building SELECT statements.
// This enables composable and flexible SELECT generation.
type SelectBuilder struct {
	ctx         *GenContext
	tables      []helper.TableInfo
	selectExprs []string
	fromClause  string
	whereClauses []string
	groupBy     []string
	having      string
	orderBy     []string
	limit       int
	offset      int
	err         error
}

// NewSelectBuilder creates a new SELECT statement builder.
func NewSelectBuilder(ctx *GenContext) *SelectBuilder {
	return &SelectBuilder{
		ctx:          ctx,
		selectExprs:  []string{},
		whereClauses: []string{},
		groupBy:      []string{},
		orderBy:      []string{},
		limit:        -1,
		offset:       -1,
	}
}

// WithTables sets the tables available for the SELECT statement.
func (b *SelectBuilder) WithTables(tables []helper.TableInfo) *SelectBuilder {
	if b.err != nil {
		return b
	}
	b.tables = tables
	return b
}

// Select adds columns or expressions to the SELECT list.
func (b *SelectBuilder) Select(exprs ...string) *SelectBuilder {
	if b.err != nil {
		return b
	}
	b.selectExprs = append(b.selectExprs, exprs...)
	return b
}

// SelectColumns adds specific columns from a table to the SELECT list.
func (b *SelectBuilder) SelectColumns(table helper.TableInfo, count int) *SelectBuilder {
	if b.err != nil {
		return b
	}
	if count <= 0 || len(table.Cols) == 0 {
		return b
	}
	
	// Select up to 'count' random columns
	maxCols := count
	if maxCols > len(table.Cols) {
		maxCols = len(table.Cols)
	}
	
	selected := make(map[int]bool)
	for len(selected) < maxCols {
		idx := b.ctx.Intn(len(table.Cols))
		if !selected[idx] {
			selected[idx] = true
			b.selectExprs = append(b.selectExprs, quoteIdent(table.Cols[idx].Name))
		}
	}
	
	return b
}

// From sets the FROM clause to a table name.
func (b *SelectBuilder) From(tableName string) *SelectBuilder {
	if b.err != nil {
		return b
	}
	b.fromClause = quoteIdent(tableName)
	return b
}

// FromTable sets the FROM clause using a TableInfo.
func (b *SelectBuilder) FromTable(table helper.TableInfo) *SelectBuilder {
	return b.From(table.Name)
}

// FromSubquery sets the FROM clause to a subquery with an alias.
func (b *SelectBuilder) FromSubquery(subquery string, alias string) *SelectBuilder {
	if b.err != nil {
		return b
	}
	b.fromClause = fmt.Sprintf("(%s) AS %s", subquery, quoteIdent(alias))
	return b
}

// Where adds a WHERE condition (multiple conditions are AND-ed).
func (b *SelectBuilder) Where(condition string) *SelectBuilder {
	if b.err != nil {
		return b
	}
	if condition != "" {
		b.whereClauses = append(b.whereClauses, condition)
	}
	return b
}

// GroupBy adds columns to GROUP BY clause.
func (b *SelectBuilder) GroupBy(columns ...string) *SelectBuilder {
	if b.err != nil {
		return b
	}
	b.groupBy = append(b.groupBy, columns...)
	return b
}

// Having sets the HAVING clause.
func (b *SelectBuilder) Having(condition string) *SelectBuilder {
	if b.err != nil {
		return b
	}
	b.having = condition
	return b
}

// OrderBy adds columns to ORDER BY clause.
func (b *SelectBuilder) OrderBy(columns ...string) *SelectBuilder {
	if b.err != nil {
		return b
	}
	b.orderBy = append(b.orderBy, columns...)
	return b
}

// Limit sets the LIMIT clause.
func (b *SelectBuilder) Limit(limit int) *SelectBuilder {
	if b.err != nil {
		return b
	}
	b.limit = limit
	return b
}

// Offset sets the OFFSET clause.
func (b *SelectBuilder) Offset(offset int) *SelectBuilder {
	if b.err != nil {
		return b
	}
	b.offset = offset
	return b
}

// Build constructs the final SELECT statement.
func (b *SelectBuilder) Build() (*SelectStmt, error) {
	if b.err != nil {
		return nil, b.err
	}
	
	// Build SELECT clause
	selectClause := "*"
	if len(b.selectExprs) > 0 {
		selectClause = strings.Join(b.selectExprs, ", ")
	}
	
	// Build FROM clause
	if b.fromClause == "" {
		return nil, fmt.Errorf("FROM clause is required")
	}
	
	// Start building SQL
	sql := fmt.Sprintf("SELECT %s FROM %s", selectClause, b.fromClause)
	
	// Add WHERE clause
	if len(b.whereClauses) > 0 {
		sql += " WHERE " + strings.Join(b.whereClauses, " AND ")
	}
	
	// Add GROUP BY clause
	if len(b.groupBy) > 0 {
		sql += " GROUP BY " + strings.Join(b.groupBy, ", ")
	}
	
	// Add HAVING clause
	if b.having != "" {
		sql += " HAVING " + b.having
	}
	
	// Add ORDER BY clause
	if len(b.orderBy) > 0 {
		sql += " ORDER BY " + strings.Join(b.orderBy, ", ")
	}
	
	// Add LIMIT clause
	if b.limit >= 0 {
		sql += fmt.Sprintf(" LIMIT %d", b.limit)
	}
	
	// Add OFFSET clause
	if b.offset >= 0 {
		sql += fmt.Sprintf(" OFFSET %d", b.offset)
	}
	
	sql += ";"
	
	return &SelectStmt{sql: sql}, nil
}

// BuildSQL is a shorthand for Build().SQL().
func (b *SelectBuilder) BuildSQL() (string, error) {
	stmt, err := b.Build()
	if err != nil {
		return "", err
	}
	return stmt.SQL(), nil
}

// InsertBuilder provides a fluent interface for building INSERT statements.
type InsertBuilder struct {
	ctx        *GenContext
	table      string
	columns    []string
	values     [][]string
	onConflict string
	err        error
}

// NewInsertBuilder creates a new INSERT statement builder.
func NewInsertBuilder(ctx *GenContext) *InsertBuilder {
	return &InsertBuilder{
		ctx:    ctx,
		values: [][]string{},
	}
}

// Into sets the target table for the INSERT.
func (b *InsertBuilder) Into(table string) *InsertBuilder {
	if b.err != nil {
		return b
	}
	b.table = table
	return b
}

// Columns sets the column list for the INSERT.
func (b *InsertBuilder) Columns(columns ...string) *InsertBuilder {
	if b.err != nil {
		return b
	}
	b.columns = columns
	return b
}

// Values adds a row of values to insert.
func (b *InsertBuilder) Values(values ...string) *InsertBuilder {
	if b.err != nil {
		return b
	}
	b.values = append(b.values, values)
	return b
}

// OnConflict sets the ON CONFLICT clause.
func (b *InsertBuilder) OnConflict(clause string) *InsertBuilder {
	if b.err != nil {
		return b
	}
	b.onConflict = clause
	return b
}

// Build constructs the final INSERT statement.
func (b *InsertBuilder) Build() (*InsertStmt, error) {
	if b.err != nil {
		return nil, b.err
	}
	
	if b.table == "" {
		return nil, fmt.Errorf("table name is required")
	}
	
	// Handle DEFAULT VALUES case
	if len(b.columns) == 0 || len(b.values) == 0 {
		sql := fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", quoteIdent(b.table))
		return &InsertStmt{sql: sql}, nil
	}
	
	// Build column list
	quotedCols := make([]string, len(b.columns))
	for i, col := range b.columns {
		quotedCols[i] = quoteIdent(col)
	}
	colList := strings.Join(quotedCols, ", ")
	
	// Build values list
	valueRows := make([]string, len(b.values))
	for i, row := range b.values {
		valueRows[i] = fmt.Sprintf("(%s)", strings.Join(row, ", "))
	}
	valuesList := strings.Join(valueRows, ", ")
	
	// Build SQL
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", quoteIdent(b.table), colList, valuesList)
	
	// Add ON CONFLICT clause if present
	if b.onConflict != "" {
		sql += " " + b.onConflict
	}
	
	sql += ";"
	
	return &InsertStmt{sql: sql}, nil
}

// BuildSQL is a shorthand for Build().SQL().
func (b *InsertBuilder) BuildSQL() (string, error) {
	stmt, err := b.Build()
	if err != nil {
		return "", err
	}
	return stmt.SQL(), nil
}

// RandomSelectBuilder creates a SELECT statement with random features enabled.
type RandomSelectBuilder struct {
	ctx *GenContext
	db  *sql.DB
}

// NewRandomSelectBuilder creates a builder that generates random SELECT statements.
func NewRandomSelectBuilder(ctx *GenContext, db *sql.DB) *RandomSelectBuilder {
	return &RandomSelectBuilder{
		ctx: ctx,
		db:  db,
	}
}

// Build generates a random SELECT statement using available tables.
func (r *RandomSelectBuilder) Build() (*SelectStmt, error) {
	tables, err := helper.GetAllTablesAndCols(r.db)
	if err != nil || len(tables) == 0 {
		return &SelectStmt{sql: "SELECT 1;"}, nil
	}
	
	// Pick a random table
	table := tables[r.ctx.Intn(len(tables))]
	
	builder := NewSelectBuilder(r.ctx).
		FromTable(table).
		Limit(1 + r.ctx.Intn(50))
	
	// Add 1-3 random columns
	if len(table.Cols) > 0 {
		numCols := 1 + r.ctx.Intn(3)
		builder.SelectColumns(table, numCols)
	} else {
		builder.Select("*")
	}
	
	// Maybe add WHERE clause
	if r.ctx.Intn(2) == 0 && len(table.Cols) > 0 {
		exprGen := NewExprGenerator(r.ctx)
		whereExpr := exprGen.GenWhereExpr([]helper.TableInfo{table})
		if whereExpr != "" {
			builder.Where(whereExpr)
		}
	}
	
	// Maybe add ORDER BY
	if r.ctx.Intn(3) == 0 && len(table.Cols) > 0 {
		col := table.Cols[r.ctx.Intn(len(table.Cols))]
		order := "ASC"
		if r.ctx.Intn(2) == 0 {
			order = "DESC"
		}
		builder.OrderBy(fmt.Sprintf("%s %s", quoteIdent(col.Name), order))
	}
	
	return builder.Build()
}
