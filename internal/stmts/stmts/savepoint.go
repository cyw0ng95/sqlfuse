package stmts

import (
	"fmt"
	"sqlfuse/internal/common"
)

// SavepointGenerator is a StmtGenerator for SAVEPOINT statements.
type SavepointGenerator struct{}

// Generate implements StmtGenerator for SAVEPOINT statements.
func (g *SavepointGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenSavepoint(ctx.LCG), nil
}

// CanGenerate implements StmtGenerator. SAVEPOINT can always be generated.
func (g *SavepointGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// ReleaseGenerator is a StmtGenerator for RELEASE SAVEPOINT statements.
type ReleaseGenerator struct{}

// Generate implements StmtGenerator for RELEASE SAVEPOINT statements.
func (g *ReleaseGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenReleaseSavepoint(ctx.LCG), nil
}

// CanGenerate implements StmtGenerator. RELEASE SAVEPOINT can always be generated.
func (g *ReleaseGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// SavepointStmt represents a SAVEPOINT or RELEASE SAVEPOINT statement.
// It embeds BaseStmt to avoid boilerplate method implementations.
type SavepointStmt struct {
	*BaseStmt
}

// GenSavepoint generates a SAVEPOINT statement.
// SAVEPOINT creates a transaction savepoint that can be rolled back to.
// According to SQLite documentation: https://sqlite.org/lang_savepoint.html
// Turso COMPAT.md: Yes (full support).
func GenSavepoint(lcg *common.LCG) Stmt {
	lcg = EnsureLCG(lcg)

	// Generate a savepoint name
	savepointName := fmt.Sprintf("sp_%d", lcg.Uint64()%10000)

	sql := fmt.Sprintf("SAVEPOINT \"%s\";", savepointName)
	return &SavepointStmt{
		BaseStmt: NewBaseStmt(sql, "savepoint", GetDefaultFlavor()),
	}
}

// GenReleaseSavepoint generates a RELEASE SAVEPOINT statement.
// RELEASE removes a savepoint and commits the changes since the savepoint.
// According to SQLite documentation: https://sqlite.org/lang_savepoint.html
// Turso COMPAT.md: Yes (full support).
func GenReleaseSavepoint(lcg *common.LCG) Stmt {
	lcg = EnsureLCG(lcg)

	// Generate a savepoint name (matching the naming scheme from GenSavepoint)
	savepointName := fmt.Sprintf("sp_%d", lcg.Uint64()%10000)

	// RELEASE can optionally include the SAVEPOINT keyword
	variants := []string{
		fmt.Sprintf("RELEASE SAVEPOINT \"%s\";", savepointName),
		fmt.Sprintf("RELEASE \"%s\";", savepointName),
	}

	sql := variants[lcg.Intn(len(variants))]
	return &SavepointStmt{
		BaseStmt: NewBaseStmt(sql, "release", GetDefaultFlavor()),
	}
}

// GenRollbackToSavepoint generates a ROLLBACK TO SAVEPOINT statement.
// This rolls back to a previously created savepoint.
func GenRollbackToSavepoint(lcg *common.LCG) Stmt {
	lcg = EnsureLCG(lcg)

	// Generate a savepoint name (matching the naming scheme from GenSavepoint)
	savepointName := fmt.Sprintf("sp_%d", lcg.Uint64()%10000)

	// ROLLBACK TO can optionally include the SAVEPOINT keyword
	variants := []string{
		fmt.Sprintf("ROLLBACK TO SAVEPOINT \"%s\";", savepointName),
		fmt.Sprintf("ROLLBACK TO \"%s\";", savepointName),
	}

	sql := variants[lcg.Intn(len(variants))]
	return &SavepointStmt{
		BaseStmt: NewBaseStmt(sql, "rollback_to_savepoint", GetDefaultFlavor()),
	}
}
