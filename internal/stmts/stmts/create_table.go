package stmts

import (
	"fmt"
	"sqlsmith-go/internal/common"
	"strings"
)

// CreateTableGenerator is a StmtGenerator for CREATE TABLE statements.
type CreateTableGenerator struct{}

// Generate implements StmtGenerator for CREATE TABLE statements.
func (g *CreateTableGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return genCreateTableWithFlavor(ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. CREATE TABLE can always be generated.
func (g *CreateTableGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// CreateTableStmt represents a CREATE TABLE statement.
type CreateTableStmt struct {
	sql    string
	flavor FlavorConfig
}

func (s *CreateTableStmt) SQL() string          { return s.sql }
func (s *CreateTableStmt) Type() string         { return "create_table" }
func (s *CreateTableStmt) Flavor() FlavorConfig { return s.flavor }

// GenCreateTable generates a simple CREATE TABLE statement using the provided LCG.
// This function is kept for backward compatibility with existing code.
// It produces 1..4 columns with common SQLite-compatible types.
func GenCreateTable(lcg *common.LCG) (Stmt, error) {
	return genCreateTableInternal(lcg)
}

// genCreateTableInternal is the internal implementation used by both old and new interfaces.
func genCreateTableInternal(lcg *common.LCG) (Stmt, error) {
	return genCreateTableWithFlavor(lcg, GetDefaultFlavor())
}

// genCreateTableWithFlavor creates a CREATE TABLE statement with flavor support.
func genCreateTableWithFlavor(lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	// choose number of columns 1..4
	n := 1 + lcg.Intn(4)
	cols := make([]string, 0, n)
	
	// Base types supported by all SQLite flavors
	types := []string{"INTEGER", "TEXT", "REAL", "BLOB"}
	
	// For go-sqlite3 flavor, JSON type is fully supported via JSON1 extension
	// which is enabled by default in standard SQLite builds
	if flavor != nil && flavor.Name() == "go-sqlite3" {
		types = append(types, "JSON")
	}
	
	for i := 0; i < n; i++ {
		colName := fmt.Sprintf("col%d", i+1)
		typeIdx := lcg.Intn(len(types))
		cols = append(cols, fmt.Sprintf("\"%s\" %s", colName, types[typeIdx]))
	}

	// generate a lightweight unique-ish table name
	tbl := fmt.Sprintf("tbl_%d", lcg.Uint64()%1000000)

	sql := fmt.Sprintf("CREATE TABLE IF NOT EXISTS \"%s\" (%s);", tbl, strings.Join(cols, ", "))
	return &CreateTableStmt{sql: sql, flavor: flavor}, nil
}
