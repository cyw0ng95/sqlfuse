package stmts

import (
	"fmt"
	"sqlsmith-go/internal/common"
)

// DropTableGenerator is a StmtGenerator for DROP TABLE statements.
type DropTableGenerator struct{}

// Generate implements StmtGenerator for DROP TABLE statements.
func (g *DropTableGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return genDropTableWithFlavor(ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. DROP TABLE can always be generated.
func (g *DropTableGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// DropTableStmt represents a DROP TABLE statement.
type DropTableStmt struct {
	sql    string
	flavor FlavorConfig
}

func (s *DropTableStmt) SQL() string          { return s.sql }
func (s *DropTableStmt) Type() string         { return "drop_table" }
func (s *DropTableStmt) Flavor() FlavorConfig { return s.flavor }

// GenDropTable generates a DROP TABLE IF EXISTS statement for a pseudo-random
// table name using the provided LCG.
// This function is kept for backward compatibility with existing code.
func GenDropTable(lcg *common.LCG) (Stmt, error) {
	return genDropTableInternal(lcg)
}

// genDropTableInternal is the internal implementation used by both old and new interfaces.
func genDropTableInternal(lcg *common.LCG) (Stmt, error) {
	return genDropTableWithFlavor(lcg, GetDefaultFlavor())
}

// genDropTableWithFlavor creates a DROP TABLE statement with flavor support.
func genDropTableWithFlavor(lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	// Use same naming scheme as GenCreateTable to sometimes target recently created tables.
	tbl := fmt.Sprintf("tbl_%d", lcg.Uint64()%1000000)
	sql := fmt.Sprintf("DROP TABLE IF EXISTS \"%s\";", tbl)
	return &DropTableStmt{sql: sql, flavor: flavor}, nil
}
