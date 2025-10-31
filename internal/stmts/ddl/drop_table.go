package ddl

import (
	"sqlfuse/internal/stmts/stmts"
	"fmt"
	"sqlfuse/internal/common"
)

// DropTableGenerator is a stmts.StmtGenerator for DROP TABLE statements.
type DropTableGenerator struct{}

// Generate implements stmts.StmtGenerator for DROP TABLE statements.
func (g *DropTableGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	return genDropTableWithFlavor(ctx.LCG, ctx.Flavor)
}

// CanGenerate implements stmts.StmtGenerator. DROP TABLE can always be generated.
func (g *DropTableGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	return true
}

// DropTableStmt represents a DROP TABLE statement.
// It embeds stmts.BaseStmt to avoid boilerplate method implementations.
type DropTableStmt struct {
	*stmts.BaseStmt
}

// GenDropTable generates a DROP TABLE IF EXISTS statement for a pseudo-random
// table name using the provided LCG.
// This function is kept for backward compatibility with existing code.
func GenDropTable(lcg *common.LCG) (stmts.Stmt, error) {
	return genDropTableInternal(lcg)
}

// genDropTableInternal is the internal implementation used by both old and new interfaces.
func genDropTableInternal(lcg *common.LCG) (stmts.Stmt, error) {
	return genDropTableWithFlavor(lcg, stmts.GetDefaultFlavor())
}

// genDropTableWithFlavor creates a DROP TABLE statement with flavor support.
func genDropTableWithFlavor(lcg *common.LCG, flavor stmts.FlavorConfig) (stmts.Stmt, error) {
	lcg = stmts.EnsureLCG(lcg)
	flavor = stmts.EnsureFlavor(flavor)

	// Use same naming scheme as GenCreateTable to sometimes target recently created tables.
	tbl := fmt.Sprintf("tbl_%d", lcg.Uint64()%1000000)
	sql := fmt.Sprintf("DROP TABLE IF EXISTS \"%s\";", tbl)
	return &DropTableStmt{
		BaseStmt: stmts.NewBaseStmt(sql, "drop_table", flavor),
	}, nil
}
