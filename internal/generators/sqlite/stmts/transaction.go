package stmts

import (
	"sqlsmith-go/internal/common"
)

// TransactionStmt represents transaction control statements (BEGIN, COMMIT, ROLLBACK).
type TransactionStmt struct {
	sql string
}

func (s *TransactionStmt) SQL() string  { return s.sql }
func (s *TransactionStmt) Type() string { return "transaction" }

// GenBeginTransaction generates a BEGIN TRANSACTION statement.
// According to Turso COMPAT.md: Partial support (no transaction names).
func GenBeginTransaction(lcg *common.LCG) Stmt {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}

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
	return &TransactionStmt{sql: sql}
}

// GenCommitTransaction generates a COMMIT TRANSACTION statement.
// According to Turso COMPAT.md: Partial support (no transaction names).
func GenCommitTransaction(lcg *common.LCG) Stmt {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}

	// Turso supports COMMIT and COMMIT TRANSACTION but not named transactions
	// END TRANSACTION is an alias for COMMIT TRANSACTION
	variants := []string{
		"COMMIT;",
		"COMMIT TRANSACTION;",
		"END;",
		"END TRANSACTION;",
	}

	sql := variants[lcg.Intn(len(variants))]
	return &TransactionStmt{sql: sql}
}

// GenRollbackTransaction generates a ROLLBACK TRANSACTION statement.
// According to Turso COMPAT.md: Yes (full support).
func GenRollbackTransaction(lcg *common.LCG) Stmt {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}

	// Turso supports ROLLBACK and ROLLBACK TRANSACTION but not named transactions
	variants := []string{
		"ROLLBACK;",
		"ROLLBACK TRANSACTION;",
	}

	sql := variants[lcg.Intn(len(variants))]
	return &TransactionStmt{sql: sql}
}
