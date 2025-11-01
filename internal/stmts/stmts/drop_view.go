package stmts

import (
	"fmt"
	"sqlfuse/internal/common"
)

// DropViewGenerator is a StmtGenerator for DROP VIEW statements.
type DropViewGenerator struct{}

// Generate implements StmtGenerator for DROP VIEW statements.
func (g *DropViewGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return genDropViewInternal(ctx.LCG)
}

// CanGenerate implements StmtGenerator. DROP VIEW can always be generated.
func (g *DropViewGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// DropViewStmt represents a DROP VIEW statement.
// It embeds BaseStmt to avoid boilerplate method implementations.
type DropViewStmt struct {
	*BaseStmt
}

// GenDropView generates a DROP VIEW IF EXISTS statement for a pseudo-random view name.
// This function is kept for backward compatibility with existing code.
func GenDropView(lcg *common.LCG) (Stmt, error) {
	return genDropViewInternal(lcg)
}

// genDropViewInternal is the internal implementation used by both old and new interfaces.
func genDropViewInternal(lcg *common.LCG) (Stmt, error) {
	lcg = EnsureLCG(lcg)

	view := fmt.Sprintf("view_%d", lcg.Uint64()%1000000)
	sql := fmt.Sprintf("DROP VIEW IF EXISTS \"%s\";", view)
	return &DropViewStmt{
		BaseStmt: NewBaseStmt(sql, "drop_view", GetDefaultFlavor()),
	}, nil
}
