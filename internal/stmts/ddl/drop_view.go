package ddl

import (
	"sqlfuse/internal/stmts/stmts"
	"fmt"
	"sqlfuse/internal/common"
)

// DropViewGenerator is a stmts.StmtGenerator for DROP VIEW statements.
type DropViewGenerator struct{}

// Generate implements stmts.StmtGenerator for DROP VIEW statements.
func (g *DropViewGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	return genDropViewInternal(ctx.LCG)
}

// CanGenerate implements stmts.StmtGenerator. DROP VIEW can always be generated.
func (g *DropViewGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	return true
}

// DropViewStmt represents a DROP VIEW statement.
// It embeds stmts.BaseStmt to avoid boilerplate method implementations.
type DropViewStmt struct {
	*stmts.BaseStmt
}

// GenDropView generates a DROP VIEW IF EXISTS statement for a pseudo-random view name.
// This function is kept for backward compatibility with existing code.
func GenDropView(lcg *common.LCG) (stmts.Stmt, error) {
	return genDropViewInternal(lcg)
}

// genDropViewInternal is the internal implementation used by both old and new interfaces.
func genDropViewInternal(lcg *common.LCG) (stmts.Stmt, error) {
	lcg = stmts.EnsureLCG(lcg)

	view := fmt.Sprintf("view_%d", lcg.Uint64()%1000000)
	sql := fmt.Sprintf("DROP VIEW IF EXISTS \"%s\";", view)
	return &DropViewStmt{
		BaseStmt: stmts.NewBaseStmt(sql, "drop_view", stmts.GetDefaultFlavor()),
	}, nil
}
