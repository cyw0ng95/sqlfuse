package other

import (
	"sqlfuse/internal/stmts/stmts"
	"fmt"
	"sqlfuse/internal/common"
)

// SavepointGenerator is a stmts.StmtGenerator for SAVEPOINT statements.
type SavepointGenerator struct{}

// Generate implements stmts.StmtGenerator for SAVEPOINT statements.
func (g *SavepointGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	return GenSavepoint(ctx.LCG), nil
}

// CanGenerate implements stmts.StmtGenerator. SAVEPOINT can always be generated.
func (g *SavepointGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	return true
}

// ReleaseGenerator is a stmts.StmtGenerator for RELEASE SAVEPOINT statements.
type ReleaseGenerator struct{}

// Generate implements stmts.StmtGenerator for RELEASE SAVEPOINT statements.
func (g *ReleaseGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	return GenReleaseSavepoint(ctx.LCG), nil
}

// CanGenerate implements stmts.StmtGenerator. RELEASE SAVEPOINT can always be generated.
func (g *ReleaseGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	return true
}

// SavepointStmt represents a SAVEPOINT or RELEASE SAVEPOINT statement.
// It embeds stmts.BaseStmt to avoid boilerplate method implementations.
type SavepointStmt struct {
	*stmts.BaseStmt
}

// GenSavepoint generates a SAVEPOINT statement.
// SAVEPOINT creates a transaction savepoint that can be rolled back to.
// According to SQLite documentation: https://sqlite.org/lang_savepoint.html
// Turso COMPAT.md: Yes (full support).
func GenSavepoint(lcg *common.LCG) stmts.Stmt {
	lcg = stmts.EnsureLCG(lcg)

	// Generate a savepoint name
	savepointName := fmt.Sprintf("sp_%d", lcg.Uint64()%10000)

	sql := fmt.Sprintf("SAVEPOINT \"%s\";", savepointName)
	return &SavepointStmt{
		BaseStmt: stmts.NewBaseStmt(sql, "savepoint", stmts.GetDefaultFlavor()),
	}
}

// GenReleaseSavepoint generates a RELEASE SAVEPOINT statement.
// RELEASE removes a savepoint and commits the changes since the savepoint.
// According to SQLite documentation: https://sqlite.org/lang_savepoint.html
// Turso COMPAT.md: Yes (full support).
func GenReleaseSavepoint(lcg *common.LCG) stmts.Stmt {
	lcg = stmts.EnsureLCG(lcg)

	// Generate a savepoint name (matching the naming scheme from GenSavepoint)
	savepointName := fmt.Sprintf("sp_%d", lcg.Uint64()%10000)

	// RELEASE can optionally include the SAVEPOINT keyword
	variants := []string{
		fmt.Sprintf("RELEASE SAVEPOINT \"%s\";", savepointName),
		fmt.Sprintf("RELEASE \"%s\";", savepointName),
	}

	sql := variants[lcg.Intn(len(variants))]
	return &SavepointStmt{
		BaseStmt: stmts.NewBaseStmt(sql, "release", stmts.GetDefaultFlavor()),
	}
}

// GenRollbackToSavepoint generates a ROLLBACK TO SAVEPOINT statement.
// This rolls back to a previously created savepoint.
func GenRollbackToSavepoint(lcg *common.LCG) stmts.Stmt {
	lcg = stmts.EnsureLCG(lcg)

	// Generate a savepoint name (matching the naming scheme from GenSavepoint)
	savepointName := fmt.Sprintf("sp_%d", lcg.Uint64()%10000)

	// ROLLBACK TO can optionally include the SAVEPOINT keyword
	variants := []string{
		fmt.Sprintf("ROLLBACK TO SAVEPOINT \"%s\";", savepointName),
		fmt.Sprintf("ROLLBACK TO \"%s\";", savepointName),
	}

	sql := variants[lcg.Intn(len(variants))]
	return &SavepointStmt{
		BaseStmt: stmts.NewBaseStmt(sql, "rollback_to_savepoint", stmts.GetDefaultFlavor()),
	}
}
