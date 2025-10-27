package stmts

import (
	"database/sql"
	"fmt"
	"sync"
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
	GenerateFn    func(ctx *GenContext) (Stmt, error)
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

// GeneratorRegistry provides a centralized registry for statement generators.
// This allows dynamic registration and lookup of generators by name.
type GeneratorRegistry struct {
	mu         sync.RWMutex
	generators map[string]StmtGenerator
}

// NewGeneratorRegistry creates a new empty generator registry.
func NewGeneratorRegistry() *GeneratorRegistry {
	return &GeneratorRegistry{
		generators: make(map[string]StmtGenerator),
	}
}

// Register adds or updates a generator in the registry.
func (r *GeneratorRegistry) Register(name string, gen StmtGenerator) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.generators[name] = gen
}

// Get retrieves a generator by name. Returns nil if not found.
func (r *GeneratorRegistry) Get(name string) StmtGenerator {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.generators[name]
}

// Has checks if a generator with the given name exists.
func (r *GeneratorRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.generators[name]
	return exists
}

// Names returns all registered generator names.
func (r *GeneratorRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.generators))
	for name := range r.generators {
		names = append(names, name)
	}
	return names
}

// DefaultRegistry returns a registry pre-populated with all standard generators.
func DefaultRegistry() *GeneratorRegistry {
	reg := NewGeneratorRegistry()

	// Register all standard generators
	reg.Register("pragma", &PragmaGenerator{})
	reg.Register("insert", &InsertGenerator{})
	reg.Register("select", &SelectGenerator{})
	reg.Register("update", &UpdateGenerator{})
	reg.Register("delete", &DeleteGenerator{})
	reg.Register("create_table", &CreateTableGenerator{})
	reg.Register("drop_table", &DropTableGenerator{})
	reg.Register("alter_table", &AlterTableGenerator{})
	reg.Register("create_view", &CreateViewGenerator{})
	reg.Register("drop_view", &DropViewGenerator{})

	return reg
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

// GenerateStmt is a convenience function that generates a statement using a named generator.
// It uses the default registry to look up the generator.
func GenerateStmt(ctx *GenContext, generatorName string) (Stmt, error) {
	reg := DefaultRegistry()
	gen := reg.Get(generatorName)
	if gen == nil {
		return nil, fmt.Errorf("generator %q not found", generatorName)
	}
	if !gen.CanGenerate(ctx) {
		return nil, fmt.Errorf("generator %q cannot generate with current context", generatorName)
	}
	return gen.Generate(ctx)
}
