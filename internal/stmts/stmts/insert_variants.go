package stmts

import (
	"database/sql"
	"sqlsmith-go/internal/common"
)

// InsertVariantGenerator generates different variants of INSERT statements.
// This implements the Strategy pattern to handle different INSERT types.
type InsertVariantGenerator struct {
	variant StmtType
}

// Generate implements StmtGenerator for INSERT statement variants.
func (g *InsertVariantGenerator) Generate(ctx *GenContext) (Stmt, error) {
	switch g.variant {
	case StmtInsert:
		return genInsertSingleWithFlavor(ctx.DB, ctx.LCG, ctx.Flavor)
	case StmtInsertMultiple:
		return genInsertMultipleWithFlavor(ctx.DB, ctx.LCG, ctx.Flavor)
	case StmtInsertBulk:
		return genInsertBulkWithFlavor(ctx.DB, ctx.LCG, ctx.Flavor)
	case StmtInsertOrReplace:
		return genInsertOrWithFlavor(ctx.DB, ctx.LCG, "REPLACE", ctx.Flavor)
	case StmtInsertOrIgnore:
		return genInsertOrWithFlavor(ctx.DB, ctx.LCG, "IGNORE", ctx.Flavor)
	case StmtInsertOrAbort:
		return genInsertOrWithFlavor(ctx.DB, ctx.LCG, "ABORT", ctx.Flavor)
	case StmtInsertOrRollback:
		return genInsertOrWithFlavor(ctx.DB, ctx.LCG, "ROLLBACK", ctx.Flavor)
	case StmtInsertOrFail:
		return genInsertOrWithFlavor(ctx.DB, ctx.LCG, "FAIL", ctx.Flavor)
	default:
		return genInsertSingleWithFlavor(ctx.DB, ctx.LCG, ctx.Flavor)
	}
}

// CanGenerate implements StmtGenerator. INSERT requires tables to exist.
func (g *InsertVariantGenerator) CanGenerate(ctx *GenContext) bool {
	return hasTables(ctx.DB)
}

// Helper function for INSERT OR variants with flavor support
func genInsertOrWithFlavor(db *sql.DB, lcg *common.LCG, conflictAction string, flavor FlavorConfig) (Stmt, error) {
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
func genInsertMultipleWithFlavor(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	return GenInsertMultiple(db, lcg)
}

func genInsertBulkWithFlavor(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	return GenInsertBulk(db, lcg)
}
