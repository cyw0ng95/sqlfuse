package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
)

// SetGenerator is a StmtGenerator for SET statements.
type SetGenerator struct{}

// Generate implements StmtGenerator for SET statements.
func (g *SetGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenSet(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. SET can always be generated.
func (g *SetGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// SetStmt represents a SET statement.
type SetStmt struct {
	*BaseStmt
}

// GenSet generates a SET statement.
// SET is DuckDB's way of configuring database settings.
// Reference: https://duckdb.org/docs/stable/sql/statements/set
func GenSet(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	// Common DuckDB settings
	settings := []struct {
		name   string
		values []string
	}{
		{"memory_limit", []string{"'1GB'", "'500MB'", "'2GB'"}},
		{"threads", []string{"1", "2", "4", "8"}},
		{"enable_profiling", []string{"true", "false"}},
		{"enable_progress_bar", []string{"true", "false"}},
		{"default_null_order", []string{"'NULLS FIRST'", "'NULLS LAST'"}},
		{"preserve_insertion_order", []string{"true", "false"}},
		{"temp_directory", []string{"'/tmp/duckdb'", "'/tmp'"}},
	}

	setting := settings[lcg.Intn(len(settings))]
	value := setting.values[lcg.Intn(len(setting.values))]

	sql := fmt.Sprintf("SET %s = %s;", setting.name, value)

	return &SetStmt{
		BaseStmt: NewBaseStmt(sql, "set", flavor),
	}, nil
}

// ResetGenerator is a StmtGenerator for RESET statements.
type ResetGenerator struct{}

// Generate implements StmtGenerator for RESET statements.
func (g *ResetGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenReset(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. RESET can always be generated.
func (g *ResetGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// ResetStmt represents a RESET statement.
type ResetStmt struct {
	*BaseStmt
}

// GenReset generates a RESET statement.
// RESET returns a setting to its default value.
// Reference: https://duckdb.org/docs/stable/sql/statements/set
func GenReset(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	// Settings that can be reset
	settings := []string{
		"memory_limit",
		"threads",
		"enable_profiling",
		"enable_progress_bar",
		"default_null_order",
		"preserve_insertion_order",
		"temp_directory",
	}

	setting := settings[lcg.Intn(len(settings))]
	sql := fmt.Sprintf("RESET %s;", setting)

	return &ResetStmt{
		BaseStmt: NewBaseStmt(sql, "reset", flavor),
	}, nil
}
