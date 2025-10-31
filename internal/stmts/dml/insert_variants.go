package dml

import (
	"sqlfuse/internal/stmts/stmts"
	"database/sql"
	"sqlfuse/internal/common"
)

// InsertVariantGenerator generates different variants of INSERT statements.
// This implements the Strategy pattern to handle different INSERT types.
type InsertVariantGenerator struct {
	variant stmts.StmtType
}

// Generate implements stmts.StmtGenerator for INSERT statement variants.
func (g *InsertVariantGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	switch g.variant {
	case stmts.StmtInsert:
		return genInsertSingleWithFlavor(ctx.DB, ctx.LCG, ctx.Flavor)
	case stmts.StmtInsertMultiple:
		return genInsertMultipleWithFlavor(ctx.DB, ctx.LCG, ctx.Flavor)
	case stmts.StmtInsertBulk:
		return genInsertBulkWithFlavor(ctx.DB, ctx.LCG, ctx.Flavor)
	case stmts.StmtInsertOrReplace:
		return genInsertOrWithFlavor(ctx.DB, ctx.LCG, "REPLACE", ctx.Flavor)
	case stmts.StmtInsertOrIgnore:
		return genInsertOrWithFlavor(ctx.DB, ctx.LCG, "IGNORE", ctx.Flavor)
	case stmts.StmtInsertOrAbort:
		return genInsertOrWithFlavor(ctx.DB, ctx.LCG, "ABORT", ctx.Flavor)
	case stmts.StmtInsertOrRollback:
		return genInsertOrWithFlavor(ctx.DB, ctx.LCG, "ROLLBACK", ctx.Flavor)
	case stmts.StmtInsertOrFail:
		return genInsertOrWithFlavor(ctx.DB, ctx.LCG, "FAIL", ctx.Flavor)
	default:
		return genInsertSingleWithFlavor(ctx.DB, ctx.LCG, ctx.Flavor)
	}
}

// CanGenerate implements stmts.StmtGenerator. INSERT requires tables to exist.
func (g *InsertVariantGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	return stmts.HasTables(ctx.DB)
}

// Helper function for INSERT OR variants with flavor support
func genInsertOrWithFlavor(db *sql.DB, lcg *common.LCG, conflictAction string, flavor stmts.FlavorConfig) (stmts.Stmt, error) {
	// Get the existing function based on conflict action
	switch conflictAction {
	case "REPLACE":
		return GenInsertOrReplace(db, lcg)
	case "IGNORE":
		return GenInsertOrIgnore(db, lcg)
	case "ABORT":
		return GenInsertOrAbort(db, lcg)
	case "ROLLBACK":
		return GenInsertOrRollback(db, lcg)
	case "FAIL":
		return GenInsertOrFail(db, lcg)
	default:
		return genInsertSingleWithFlavor(db, lcg, flavor)
	}
}

// Helper functions with flavor support
func genInsertMultipleWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (stmts.Stmt, error) {
	return GenInsertMultiple(db, lcg)
}

func genInsertBulkWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (stmts.Stmt, error) {
	return GenInsertBulk(db, lcg)
}
