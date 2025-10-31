package ddl

import (
	"sqlfuse/internal/stmts/stmts"
	"fmt"
	"sqlfuse/internal/common"
)

// CreateViewGenerator is a stmts.StmtGenerator for CREATE VIEW statements.
type CreateViewGenerator struct{}

// Generate implements stmts.StmtGenerator for CREATE VIEW statements.
func (g *CreateViewGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	return genCreateViewInternal(ctx.LCG)
}

// CanGenerate implements stmts.StmtGenerator. CREATE VIEW can always be generated.
func (g *CreateViewGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	return true
}

// CreateViewStmt represents a CREATE VIEW statement.
// It embeds stmts.BaseStmt to avoid boilerplate method implementations.
type CreateViewStmt struct {
	*stmts.BaseStmt
}

// GenCreateView generates a simple CREATE VIEW statement that selects a constant.
// Using a constant SELECT avoids depending on existing tables.
// This function is kept for backward compatibility with existing code.
func GenCreateView(lcg *common.LCG) (stmts.Stmt, error) {
	return genCreateViewInternal(lcg)
}

// genCreateViewInternal is the internal implementation used by both old and new interfaces.
func genCreateViewInternal(lcg *common.LCG) (stmts.Stmt, error) {
	lcg = stmts.EnsureLCG(lcg)

	view := fmt.Sprintf("view_%d", lcg.Uint64()%1000000)
	sql := fmt.Sprintf("CREATE VIEW IF NOT EXISTS \"%s\" AS SELECT 1 AS col1;", view)
	return &CreateViewStmt{
		BaseStmt: stmts.NewBaseStmt(sql, "create_view", stmts.GetDefaultFlavor()),
	}, nil
}
