package stmts

import (
	"sqlsmith-go/internal/common"
)

// TransactionStmt represents transaction control statements (BEGIN, COMMIT, ROLLBACK).
// It embeds BaseStmt to avoid boilerplate method implementations.
type TransactionStmt struct {
	*BaseStmt
}

// GenBeginTransaction generates a BEGIN TRANSACTION statement.
// According to Turso COMPAT.md: Partial support (no transaction names).
func GenBeginTransaction(lcg *common.LCG) Stmt {
	lcg = ensureLCG(lcg)

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
	lcg = ensureLCG(lcg)

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
	lcg = ensureLCG(lcg)

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
