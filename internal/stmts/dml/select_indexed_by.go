package dml

import (
	"sqlfuse/internal/stmts/stmts"
	"database/sql"
	"fmt"
	"sqlfuse/internal/stmts/helper"
)

// SelectIndexedByGenerator generates SELECT statements with INDEXED BY clause
type SelectIndexedByGenerator struct{}

// Generate creates a SELECT statement with INDEXED BY clause
func (g *SelectIndexedByGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	return genSelectIndexedBy(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements stmts.StmtGenerator
func (g *SelectIndexedByGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	tables, err := helper.GetAllTablesAndCols(ctx.DB, ctx.Flavor.Name())
	return err == nil && len(tables) > 0
}

// genSelectIndexedBy generates a SELECT with INDEXED BY clause
func genSelectIndexedBy(db *sql.DB, lcg interface{ Intn(int) int }, flavor stmts.FlavorConfig) (stmts.Stmt, error) {
	if flavor == nil {
		flavor = stmts.GetDefaultFlavor()
	}
	
	tables, err := helper.GetAllTablesAndCols(db, flavor.Name())
	if err != nil || len(tables) == 0 {
		return &SelectStmt{sql: "SELECT 1;", flavor: flavor}, nil
	}

	// Pick a random table
	table := tables[lcg.Intn(len(tables))]
	if len(table.Cols) == 0 {
		return &SelectStmt{sql: "SELECT 1;", flavor: flavor}, nil
	}

	// Generate a plausible index name
	// Common patterns: idx_tablename_columnname
	col := table.Cols[lcg.Intn(len(table.Cols))]
	indexName := fmt.Sprintf("idx_%s_%s", table.Name, col.Name)
	
	// Truncate if too long (SQLite limit is 1000 characters, but keep it reasonable)
	if len(indexName) > 50 {
		indexName = indexName[:50]
	}

	// Select random columns
	maxCols := len(table.Cols)
	if maxCols > 3 {
		maxCols = 3
	}
	nCols := 1 + lcg.Intn(maxCols)
	
	colNames := ""
	for i := 0; i < nCols; i++ {
		if i > 0 {
			colNames += ", "
		}
		colNames += stmts.QuoteIdent(table.Cols[i].Name)
	}

	sql := fmt.Sprintf("SELECT %s FROM %s INDEXED BY %s",
		colNames,
		stmts.QuoteIdent(table.Name),
		stmts.QuoteIdent(indexName))

	// Optionally add WHERE clause
	if lcg.Intn(100) < 50 { // 50% chance
		whereCol := table.Cols[lcg.Intn(len(table.Cols))]
		sql += fmt.Sprintf(" WHERE %s > %d", stmts.QuoteIdent(whereCol.Name), lcg.Intn(100))
	}

	// Optionally add LIMIT
	if lcg.Intn(100) < 30 { // 30% chance
		sql += fmt.Sprintf(" LIMIT %d", 1+lcg.Intn(100))
	}

	sql += ";"

	return &SelectStmt{sql: sql, flavor: flavor}, nil
}

// SelectNotIndexedGenerator generates SELECT statements with NOT INDEXED clause
type SelectNotIndexedGenerator struct{}

// Generate creates a SELECT statement with NOT INDEXED clause
func (g *SelectNotIndexedGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	return genSelectNotIndexed(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements stmts.StmtGenerator
func (g *SelectNotIndexedGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	tables, err := helper.GetAllTablesAndCols(ctx.DB, ctx.Flavor.Name())
	return err == nil && len(tables) > 0
}

// genSelectNotIndexed generates a SELECT with NOT INDEXED clause
func genSelectNotIndexed(db *sql.DB, lcg interface{ Intn(int) int }, flavor stmts.FlavorConfig) (stmts.Stmt, error) {
	if flavor == nil {
		flavor = stmts.GetDefaultFlavor()
	}
	
	tables, err := helper.GetAllTablesAndCols(db, flavor.Name())
	if err != nil || len(tables) == 0 {
		return &SelectStmt{sql: "SELECT 1;", flavor: flavor}, nil
	}

	// Pick a random table
	table := tables[lcg.Intn(len(tables))]
	if len(table.Cols) == 0 {
		return &SelectStmt{sql: "SELECT 1;", flavor: flavor}, nil
	}

	// Select random columns
	maxCols := len(table.Cols)
	if maxCols > 3 {
		maxCols = 3
	}
	nCols := 1 + lcg.Intn(maxCols)
	
	colNames := ""
	for i := 0; i < nCols; i++ {
		if i > 0 {
			colNames += ", "
		}
		colNames += stmts.QuoteIdent(table.Cols[i].Name)
	}

	sql := fmt.Sprintf("SELECT %s FROM %s NOT INDEXED",
		colNames,
		stmts.QuoteIdent(table.Name))

	// Optionally add WHERE clause
	if lcg.Intn(100) < 50 { // 50% chance
		whereCol := table.Cols[lcg.Intn(len(table.Cols))]
		sql += fmt.Sprintf(" WHERE %s IS NOT NULL", stmts.QuoteIdent(whereCol.Name))
	}

	// Optionally add LIMIT
	if lcg.Intn(100) < 30 { // 30% chance
		sql += fmt.Sprintf(" LIMIT %d", 1+lcg.Intn(100))
	}

	sql += ";"

	return &SelectStmt{sql: sql, flavor: flavor}, nil
}

// NewSelectIndexedByGenerator creates a new INDEXED BY SELECT generator
func NewSelectIndexedByGenerator() stmts.StmtGenerator {
	return &SelectIndexedByGenerator{}
}

// NewSelectNotIndexedGenerator creates a new NOT INDEXED SELECT generator
func NewSelectNotIndexedGenerator() stmts.StmtGenerator {
	return &SelectNotIndexedGenerator{}
}
