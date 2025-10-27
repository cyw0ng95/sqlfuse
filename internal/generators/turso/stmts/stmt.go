package stmts

import (
	"database/sql"
)

// Stmt represents a SQL statement that can be generated and executed.
type Stmt interface {
	SQL() string
	Type() string // e.g. "pragma", "ddl", "dml", "select", "insert", etc.
}

// StmtGenerator is a unified interface for all statement generators.
// All statement generators should implement this interface for consistency.
type StmtGenerator interface {
	// Generate creates a new statement using the provided context.
	// Returns a Stmt or an error if generation failed.
	Generate(ctx *GenContext) (Stmt, error)

	// CanGenerate returns true if this generator can produce a valid statement
	// given the current context (e.g., requires tables to exist).
	CanGenerate(ctx *GenContext) bool
}

// GeneratorFunc is a function type that implements StmtGenerator.
// This allows using functions as generators for simpler cases.
type GeneratorFunc func(ctx *GenContext) (Stmt, error)

// Generate implements StmtGenerator for GeneratorFunc.
func (f GeneratorFunc) Generate(ctx *GenContext) (Stmt, error) {
	return f(ctx)
}

// CanGenerate implements StmtGenerator for GeneratorFunc.
// By default, function generators can always generate (returns true).
func (f GeneratorFunc) CanGenerate(ctx *GenContext) bool {
	return true
}

// StmtGeneratorWithCheck is a StmtGenerator that includes a custom CanGenerate check.
type StmtGeneratorWithCheck struct {
	GenerateFn   func(ctx *GenContext) (Stmt, error)
	CanGenerateFn func(ctx *GenContext) bool
}

// Generate implements StmtGenerator.
func (g *StmtGeneratorWithCheck) Generate(ctx *GenContext) (Stmt, error) {
	return g.GenerateFn(ctx)
}

// CanGenerate implements StmtGenerator.
func (g *StmtGeneratorWithCheck) CanGenerate(ctx *GenContext) bool {
	if g.CanGenerateFn == nil {
		return true
	}
	return g.CanGenerateFn(ctx)
}

// PragmaStmt represents a PRAGMA statement.
type PragmaStmt struct {
	sql string
}

func (p *PragmaStmt) SQL() string  { return p.sql }
func (p *PragmaStmt) Type() string { return "pragma" }

// Helper function to check if database has tables (used by generators that need tables).
func hasTables(db *sql.DB) bool {
	if db == nil {
		return false
	}
	// Quick check: query for any user table
	row := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	var count int
	if err := row.Scan(&count); err != nil {
		return false
	}
	return count > 0
}
