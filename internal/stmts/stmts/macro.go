package stmts

import (
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
)

// CreateMacroGenerator is a StmtGenerator for CREATE MACRO statements.
type CreateMacroGenerator struct{}

// Generate implements StmtGenerator for CREATE MACRO statements.
func (g *CreateMacroGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenCreateMacro(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. CREATE MACRO can always be generated.
func (g *CreateMacroGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// CreateMacroStmt represents a CREATE MACRO statement.
type CreateMacroStmt struct {
	*BaseStmt
}

// GenCreateMacro generates a CREATE MACRO statement.
// Macros are DuckDB's way of creating reusable SQL functions.
// Reference: https://duckdb.org/docs/stable/sql/statements/create_macro
func GenCreateMacro(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	macroName := fmt.Sprintf("macro_%d", lcg.Uint64()%10000)

	var sql string
	choice := lcg.Intn(4)

	if choice == 0 {
		// Simple scalar macro
		sql = fmt.Sprintf("CREATE OR REPLACE MACRO %s(x) AS x + 1;", macroName)
	} else if choice == 1 {
		// Macro with multiple parameters
		sql = fmt.Sprintf("CREATE OR REPLACE MACRO %s(a, b) AS a * b + 10;", macroName)
	} else if choice == 2 {
		// String manipulation macro
		sql = fmt.Sprintf("CREATE OR REPLACE MACRO %s(s) AS upper(s);", macroName)
	} else {
		// Macro with CASE expression
		sql = fmt.Sprintf("CREATE OR REPLACE MACRO %s(x) AS CASE WHEN x > 0 THEN 'positive' ELSE 'non-positive' END;", macroName)
	}

	return &CreateMacroStmt{
		BaseStmt: NewBaseStmt(sql, "create_macro", flavor),
	}, nil
}

// DropMacroGenerator is a StmtGenerator for DROP MACRO statements.
type DropMacroGenerator struct{}

// Generate implements StmtGenerator for DROP MACRO statements.
func (g *DropMacroGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenDropMacro(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. DROP MACRO can always be generated.
func (g *DropMacroGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// DropMacroStmt represents a DROP MACRO statement.
type DropMacroStmt struct {
	*BaseStmt
}

// GenDropMacro generates a DROP MACRO statement.
// Reference: https://duckdb.org/docs/stable/sql/statements/drop
func GenDropMacro(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	macroName := fmt.Sprintf("macro_%d", lcg.Uint64()%10000)
	sql := fmt.Sprintf("DROP MACRO IF EXISTS %s;", macroName)

	return &DropMacroStmt{
		BaseStmt: NewBaseStmt(sql, "drop_macro", flavor),
	}, nil
}

// CallGenerator is a StmtGenerator for CALL statements.
type CallGenerator struct{}

// Generate implements StmtGenerator for CALL statements.
func (g *CallGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenCall(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. CALL can always be generated.
func (g *CallGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// CallStmt represents a CALL statement.
type CallStmt struct {
	*BaseStmt
}

// GenCall generates a CALL statement.
// CALL is used to invoke table-producing macros or functions.
// Reference: https://duckdb.org/docs/stable/sql/statements/call
func GenCall(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	// Call a built-in or user-defined macro/function
	choice := lcg.Intn(3)
	var sql string

	if choice == 0 {
		// Call a hypothetical user macro
		macroName := fmt.Sprintf("macro_%d", lcg.Uint64()%100)
		sql = fmt.Sprintf("CALL %s(%d);", macroName, lcg.Intn(100))
	} else if choice == 1 {
		// Call with multiple arguments
		macroName := fmt.Sprintf("macro_%d", lcg.Uint64()%100)
		sql = fmt.Sprintf("CALL %s(%d, %d);", macroName, lcg.Intn(100), lcg.Intn(100))
	} else {
		// Call with string argument
		macroName := fmt.Sprintf("macro_%d", lcg.Uint64()%100)
		sql = fmt.Sprintf("CALL %s('test');", macroName)
	}

	return &CallStmt{
		BaseStmt: NewBaseStmt(sql, "call", flavor),
	}, nil
}
