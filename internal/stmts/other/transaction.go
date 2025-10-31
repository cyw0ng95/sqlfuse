package other

import (
	"sqlfuse/internal/stmts/stmts"
	"sqlfuse/internal/common"
)

// BeginTransactionGenerator is a stmts.StmtGenerator for BEGIN TRANSACTION statements.
type BeginTransactionGenerator struct{}

// Generate implements stmts.StmtGenerator for BEGIN TRANSACTION statements.
func (g *BeginTransactionGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	return GenBeginTransaction(ctx.LCG), nil
}

// CanGenerate implements stmts.StmtGenerator. BEGIN TRANSACTION can always be generated.
func (g *BeginTransactionGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	return true
}

// CommitTransactionGenerator is a stmts.StmtGenerator for COMMIT TRANSACTION statements.
type CommitTransactionGenerator struct{}

// Generate implements stmts.StmtGenerator for COMMIT TRANSACTION statements.
func (g *CommitTransactionGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	return GenCommitTransaction(ctx.LCG), nil
}

// CanGenerate implements stmts.StmtGenerator. COMMIT TRANSACTION can always be generated.
func (g *CommitTransactionGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	return true
}

// RollbackTransactionGenerator is a stmts.StmtGenerator for ROLLBACK TRANSACTION statements.
type RollbackTransactionGenerator struct{}

// Generate implements stmts.StmtGenerator for ROLLBACK TRANSACTION statements.
func (g *RollbackTransactionGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	return GenRollbackTransaction(ctx.LCG), nil
}

// CanGenerate implements stmts.StmtGenerator. ROLLBACK TRANSACTION can always be generated.
func (g *RollbackTransactionGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	return true
}

// TransactionStmt represents transaction control statements (BEGIN, COMMIT, ROLLBACK).
// It embeds stmts.BaseStmt to avoid boilerplate method implementations.
type TransactionStmt struct {
	*stmts.BaseStmt
}

// GenBeginTransaction generates a BEGIN TRANSACTION statement.
// According to Turso COMPAT.md: Partial support (no transaction names).
func GenBeginTransaction(lcg *common.LCG) stmts.Stmt {
	lcg = stmts.EnsureLCG(lcg)

	// Turso supports BEGIN, BEGIN TRANSACTION, BEGIN DEFERRED, BEGIN IMMEDIATE, BEGIN EXCLUSIVE
	// but not named transactions
	variants := []string{
		"BEGIN;",
		"BEGIN TRANSACTION;",
		"BEGIN DEFERRED;",
		"BEGIN DEFERRED TRANSACTION;",
		"BEGIN IMMEDIATE;",
		"BEGIN IMMEDIATE TRANSACTION;",
		"BEGIN EXCLUSIVE;",
		"BEGIN EXCLUSIVE TRANSACTION;",
	}

	sql := variants[lcg.Intn(len(variants))]
	return &TransactionStmt{
		BaseStmt: stmts.NewBaseStmt(sql, "transaction", stmts.GetDefaultFlavor()),
	}
}

// GenCommitTransaction generates a COMMIT TRANSACTION statement.
// According to Turso COMPAT.md: Partial support (no transaction names).
func GenCommitTransaction(lcg *common.LCG) stmts.Stmt {
	lcg = stmts.EnsureLCG(lcg)

	// Turso supports COMMIT and COMMIT TRANSACTION but not named transactions
	// END TRANSACTION is an alias for COMMIT TRANSACTION
	variants := []string{
		"COMMIT;",
		"COMMIT TRANSACTION;",
		"END;",
		"END TRANSACTION;",
	}

	sql := variants[lcg.Intn(len(variants))]
	return &TransactionStmt{
		BaseStmt: stmts.NewBaseStmt(sql, "transaction", stmts.GetDefaultFlavor()),
	}
}

// GenRollbackTransaction generates a ROLLBACK TRANSACTION statement.
// According to Turso COMPAT.md: Yes (full support).
func GenRollbackTransaction(lcg *common.LCG) stmts.Stmt {
	lcg = stmts.EnsureLCG(lcg)

	// Turso supports ROLLBACK and ROLLBACK TRANSACTION but not named transactions
	variants := []string{
		"ROLLBACK;",
		"ROLLBACK TRANSACTION;",
	}

	sql := variants[lcg.Intn(len(variants))]
	return &TransactionStmt{
		BaseStmt: stmts.NewBaseStmt(sql, "transaction", stmts.GetDefaultFlavor()),
	}
}
