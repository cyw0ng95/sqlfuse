package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
)

// CreateTypeGenerator is a StmtGenerator for CREATE TYPE statements.
type CreateTypeGenerator struct{}

// Generate implements StmtGenerator for CREATE TYPE statements.
func (g *CreateTypeGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenCreateType(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. CREATE TYPE can always be generated.
func (g *CreateTypeGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// CreateTypeStmt represents a CREATE TYPE statement.
type CreateTypeStmt struct {
	*BaseStmt
}

// GenCreateType generates a CREATE TYPE statement.
// DuckDB supports ENUM types.
// Reference: https://duckdb.org/docs/stable/sql/statements/create_type
func GenCreateType(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	typeName := fmt.Sprintf("type_%d", lcg.Uint64()%10000)
	
	// DuckDB supports ENUM types
	choice := lcg.Intn(3)
	var sql string
	
	if choice == 0 {
		// Simple ENUM with a few values
		sql = fmt.Sprintf("CREATE TYPE %s AS ENUM ('value1', 'value2', 'value3');", typeName)
	} else if choice == 1 {
		// ENUM with status values
		sql = fmt.Sprintf("CREATE TYPE %s AS ENUM ('pending', 'active', 'completed', 'cancelled');", typeName)
	} else {
		// ENUM with day names
		sql = fmt.Sprintf("CREATE TYPE %s AS ENUM ('Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday');", typeName)
	}

	return &CreateTypeStmt{
		BaseStmt: NewBaseStmt(sql, "create_type", flavor),
	}, nil
}

// DropTypeGenerator is a StmtGenerator for DROP TYPE statements.
type DropTypeGenerator struct{}

// Generate implements StmtGenerator for DROP TYPE statements.
func (g *DropTypeGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenDropType(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. DROP TYPE can always be generated.
func (g *DropTypeGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// DropTypeStmt represents a DROP TYPE statement.
type DropTypeStmt struct {
	*BaseStmt
}

// GenDropType generates a DROP TYPE statement.
// Reference: https://duckdb.org/docs/stable/sql/statements/drop
func GenDropType(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	typeName := fmt.Sprintf("type_%d", lcg.Uint64()%10000)
	sql := fmt.Sprintf("DROP TYPE IF EXISTS %s;", typeName)

	return &DropTypeStmt{
		BaseStmt: NewBaseStmt(sql, "drop_type", flavor),
	}, nil
}
