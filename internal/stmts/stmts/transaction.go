package stmts

import (
	"sqlfuse/internal/common"
)

// BeginTransactionGenerator is a StmtGenerator for BEGIN TRANSACTION statements.
type BeginTransactionGenerator struct{}

// Generate implements StmtGenerator for BEGIN TRANSACTION statements.
func (g *BeginTransactionGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenBeginTransaction(ctx.LCG), nil
}

// CanGenerate implements StmtGenerator. BEGIN TRANSACTION can always be generated.
func (g *BeginTransactionGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// CommitTransactionGenerator is a StmtGenerator for COMMIT TRANSACTION statements.
type CommitTransactionGenerator struct{}

// Generate implements StmtGenerator for COMMIT TRANSACTION statements.
func (g *CommitTransactionGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenCommitTransaction(ctx.LCG), nil
}

// CanGenerate implements StmtGenerator. COMMIT TRANSACTION can always be generated.
func (g *CommitTransactionGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// RollbackTransactionGenerator is a StmtGenerator for ROLLBACK TRANSACTION statements.
type RollbackTransactionGenerator struct{}

// Generate implements StmtGenerator for ROLLBACK TRANSACTION statements.
func (g *RollbackTransactionGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenRollbackTransaction(ctx.LCG), nil
}

// CanGenerate implements StmtGenerator. ROLLBACK TRANSACTION can always be generated.
func (g *RollbackTransactionGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// TransactionStmt represents transaction control statements (BEGIN, COMMIT, ROLLBACK).
// It embeds BaseStmt to avoid boilerplate method implementations.
type TransactionStmt struct {
	*BaseStmt
}

// GenBeginTransaction generates a BEGIN TRANSACTION statement.
// According to Turso COMPAT.md: Partial support (no transaction names).
func GenBeginTransaction(lcg *common.LCG) Stmt {
	lcg = EnsureLCG(lcg)

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
		BaseStmt: NewBaseStmt(sql, "transaction", GetDefaultFlavor()),
	}
}

// GenCommitTransaction generates a COMMIT TRANSACTION statement.
// According to Turso COMPAT.md: Partial support (no transaction names).
func GenCommitTransaction(lcg *common.LCG) Stmt {
	lcg = EnsureLCG(lcg)

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
		BaseStmt: NewBaseStmt(sql, "transaction", GetDefaultFlavor()),
	}
}

// GenRollbackTransaction generates a ROLLBACK TRANSACTION statement.
// According to Turso COMPAT.md: Yes (full support).
func GenRollbackTransaction(lcg *common.LCG) Stmt {
	lcg = EnsureLCG(lcg)

	// Turso supports ROLLBACK and ROLLBACK TRANSACTION but not named transactions
	variants := []string{
		"ROLLBACK;",
		"ROLLBACK TRANSACTION;",
	}

	sql := variants[lcg.Intn(len(variants))]
	return &TransactionStmt{
		BaseStmt: NewBaseStmt(sql, "transaction", GetDefaultFlavor()),
	}
}
