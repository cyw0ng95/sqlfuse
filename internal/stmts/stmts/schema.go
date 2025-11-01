package stmts

import (
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
)

// CreateSchemaGenerator is a StmtGenerator for CREATE SCHEMA statements.
type CreateSchemaGenerator struct{}

// Generate implements StmtGenerator for CREATE SCHEMA statements.
func (g *CreateSchemaGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenCreateSchema(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. CREATE SCHEMA can always be generated.
func (g *CreateSchemaGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// CreateSchemaStmt represents a CREATE SCHEMA statement.
type CreateSchemaStmt struct {
	*BaseStmt
}

// GenCreateSchema generates a CREATE SCHEMA statement.
// Reference: https://duckdb.org/docs/stable/sql/statements/create_schema
func GenCreateSchema(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = EnsureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	schemaName := fmt.Sprintf("schema_%d", lcg.Uint64()%10000)

	var sql string
	if lcg.Intn(2) == 0 {
		// 50% chance: with IF NOT EXISTS
		sql = fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS \"%s\";", schemaName)
	} else {
		sql = fmt.Sprintf("CREATE SCHEMA \"%s\";", schemaName)
	}

	return &CreateSchemaStmt{
		BaseStmt: NewBaseStmt(sql, "create_schema", flavor),
	}, nil
}

// DropSchemaGenerator is a StmtGenerator for DROP SCHEMA statements.
type DropSchemaGenerator struct{}

// Generate implements StmtGenerator for DROP SCHEMA statements.
func (g *DropSchemaGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenDropSchema(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. DROP SCHEMA can always be generated.
func (g *DropSchemaGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// DropSchemaStmt represents a DROP SCHEMA statement.
type DropSchemaStmt struct {
	*BaseStmt
}

// GenDropSchema generates a DROP SCHEMA statement.
// Reference: https://duckdb.org/docs/stable/sql/statements/drop
func GenDropSchema(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = EnsureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	schemaName := fmt.Sprintf("schema_%d", lcg.Uint64()%10000)

	var sql string
	choice := lcg.Intn(4)

	if choice == 0 {
		// 25% chance: with IF EXISTS
		sql = fmt.Sprintf("DROP SCHEMA IF EXISTS \"%s\";", schemaName)
	} else if choice == 1 {
		// 25% chance: with CASCADE
		sql = fmt.Sprintf("DROP SCHEMA IF EXISTS \"%s\" CASCADE;", schemaName)
	} else if choice == 2 {
		// 25% chance: with RESTRICT
		sql = fmt.Sprintf("DROP SCHEMA IF EXISTS \"%s\" RESTRICT;", schemaName)
	} else {
		// 25% chance: basic DROP
		sql = fmt.Sprintf("DROP SCHEMA IF EXISTS \"%s\";", schemaName)
	}

	return &DropSchemaStmt{
		BaseStmt: NewBaseStmt(sql, "drop_schema", flavor),
	}, nil
}
