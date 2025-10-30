package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/stmts/helper"
)

// CreateTriggerGenerator is a StmtGenerator for CREATE TRIGGER statements.
type CreateTriggerGenerator struct{}

// Generate implements StmtGenerator for CREATE TRIGGER statements.
func (g *CreateTriggerGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenCreateTrigger(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. CREATE TRIGGER can always be generated.
func (g *CreateTriggerGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// DropTriggerGenerator is a StmtGenerator for DROP TRIGGER statements.
type DropTriggerGenerator struct{}

// Generate implements StmtGenerator for DROP TRIGGER statements.
func (g *DropTriggerGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenDropTrigger(ctx.LCG)
}

// CanGenerate implements StmtGenerator. DROP TRIGGER can always be generated.
func (g *DropTriggerGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// CreateTriggerStmt represents a CREATE TRIGGER statement.
// It embeds BaseStmt to avoid boilerplate method implementations.
type CreateTriggerStmt struct {
	*BaseStmt
}

// DropTriggerStmt represents a DROP TRIGGER statement.
// It embeds BaseStmt to avoid boilerplate method implementations.
type DropTriggerStmt struct {
	*BaseStmt
}

// GenCreateTrigger generates a CREATE TRIGGER statement.
// According to SQLite documentation: https://sqlite.org/lang_createtrigger.html
// Turso COMPAT.md: Partial support (some limitations on trigger actions).
func GenCreateTrigger(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	// Get available tables (only if db is not nil)
	var tables []helper.TableInfo
	var err error
	if db != nil {
		tables, err = helper.GetAllTablesAndCols(db, "sqlite")
	}
	
	if db == nil || err != nil || len(tables) == 0 {
		// Fallback to a simple trigger without schema
		return genCreateTriggerFallback(lcg, flavor), nil
	}

	tbl := tables[lcg.Intn(len(tables))]
	if len(tbl.Cols) == 0 {
		return genCreateTriggerFallback(lcg, flavor), nil
	}

	// Generate trigger name
	triggerName := fmt.Sprintf("trg_%s_%d", tbl.Name, lcg.Uint64()%100000)

	// Choose timing: BEFORE, AFTER, or INSTEAD OF (for views)
	timings := []string{"BEFORE", "AFTER"}
	timing := timings[lcg.Intn(len(timings))]

	// Choose event: INSERT, UPDATE, DELETE, or UPDATE OF columns
	events := []string{"INSERT", "UPDATE", "DELETE"}
	event := events[lcg.Intn(len(events))]

	// For UPDATE, sometimes add OF column_list
	if event == "UPDATE" && lcg.Intn(3) == 0 && len(tbl.Cols) > 0 {
		col := tbl.Cols[lcg.Intn(len(tbl.Cols))]
		event = fmt.Sprintf("UPDATE OF \"%s\"", col.Name)
	}

	// Optional WHEN clause (30% chance)
	whenClause := ""
	if lcg.Intn(10) < 3 && len(tbl.Cols) > 0 {
		col := tbl.Cols[lcg.Intn(len(tbl.Cols))]
		// Use NEW or OLD references depending on the event
		ref := "NEW"
		if event == "DELETE" {
			ref = "OLD"
		}
		whenClause = fmt.Sprintf(" WHEN %s.\"%s\" IS NOT NULL", ref, col.Name)
	}

	// Generate trigger action (simple SELECT for safety)
	action := "SELECT 1"

	// Build the trigger
	sql := fmt.Sprintf("CREATE TRIGGER IF NOT EXISTS \"%s\" %s %s ON \"%s\"%s BEGIN %s; END;",
		triggerName, timing, event, tbl.Name, whenClause, action)

	return &CreateTriggerStmt{
		BaseStmt: NewBaseStmt(sql, "create_trigger", flavor),
	}, nil
}

// genCreateTriggerFallback generates a simple CREATE TRIGGER without schema information
func genCreateTriggerFallback(lcg *common.LCG, flavor FlavorConfig) *CreateTriggerStmt {
	triggerName := fmt.Sprintf("trg_%d", lcg.Uint64()%100000)
	tblName := fmt.Sprintf("tbl_%d", lcg.Uint64()%1000000)

	timings := []string{"BEFORE", "AFTER"}
	timing := timings[lcg.Intn(len(timings))]

	events := []string{"INSERT", "UPDATE", "DELETE"}
	event := events[lcg.Intn(len(events))]

	sql := fmt.Sprintf("CREATE TRIGGER IF NOT EXISTS \"%s\" %s %s ON \"%s\" BEGIN SELECT 1; END;",
		triggerName, timing, event, tblName)

	return &CreateTriggerStmt{
		BaseStmt: NewBaseStmt(sql, "create_trigger", flavor),
	}
}

// GenDropTrigger generates a DROP TRIGGER statement.
// According to SQLite documentation: https://sqlite.org/lang_droptrigger.html
// Turso COMPAT.md: Yes (full support).
func GenDropTrigger(lcg *common.LCG) (Stmt, error) {
	lcg = ensureLCG(lcg)

	// Generate trigger name (matching naming scheme from GenCreateTrigger)
	triggerName := fmt.Sprintf("trg_%d", lcg.Uint64()%100000)

	// Always use IF EXISTS to avoid errors
	sql := fmt.Sprintf("DROP TRIGGER IF EXISTS \"%s\";", triggerName)

	return &DropTriggerStmt{
		BaseStmt: NewBaseStmt(sql, "drop_trigger", GetDefaultFlavor()),
	}, nil
}
