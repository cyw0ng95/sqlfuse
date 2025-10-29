package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
)

// CompoundSelectGenerator is a StmtGenerator for compound SELECT statements (UNION, INTERSECT, EXCEPT).
type CompoundSelectGenerator struct {
	variant StmtType
}

// Generate implements StmtGenerator for compound SELECT statements.
func (g *CompoundSelectGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenCompoundSelect(ctx.DB, ctx.LCG, g.variant, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. Compound SELECT can always be generated.
func (g *CompoundSelectGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// CompoundSelectStmt represents a compound SELECT statement.
// It embeds BaseStmt to avoid boilerplate method implementations.
type CompoundSelectStmt struct {
	*BaseStmt
}

// GenCompoundSelect generates a compound SELECT statement (UNION, INTERSECT, or EXCEPT).
// According to SQLite documentation: https://sqlite.org/lang_select.html#compound_select_statements
// Turso COMPAT.md: Yes (full support for UNION, INTERSECT, EXCEPT).
func GenCompoundSelect(db *sql.DB, lcg *common.LCG, variant StmtType, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	// Determine the compound operator based on variant
	var operator string
	var stmtType string

	switch variant {
	case StmtSelectUnion:
		// UNION or UNION ALL
		if lcg.Intn(2) == 0 {
			operator = "UNION"
		} else {
			operator = "UNION ALL"
		}
		stmtType = "select_union"
	case StmtSelectIntersect:
		operator = "INTERSECT"
		stmtType = "select_intersect"
	case StmtSelectExcept:
		operator = "EXCEPT"
		stmtType = "select_except"
	default:
		// Default to UNION
		operator = "UNION"
		stmtType = "select_union"
	}

	// Generate two SELECT statements to combine
	// Use simple SELECTs to avoid complexity and ensure column compatibility
	var sql string
	
	// If no db, use simple fallback
	if db == nil {
		sql = fmt.Sprintf("SELECT 1, 'a' %s SELECT 2, 'b';", operator)
	} else {
		select1, err1 := GenSelect(db, lcg)
		select2, err2 := GenSelect(db, lcg)

		if err1 != nil || err2 != nil {
			// Fallback to simple constant selects that are guaranteed to be compatible
			sql = fmt.Sprintf("SELECT 1, 'a' %s SELECT 2, 'b';", operator)
		} else {
			// Remove trailing semicolons from the individual SELECTs
			sql1 := select1.SQL()
			sql2 := select2.SQL()

			if len(sql1) > 0 && sql1[len(sql1)-1] == ';' {
				sql1 = sql1[:len(sql1)-1]
			}
			if len(sql2) > 0 && sql2[len(sql2)-1] == ';' {
				sql2 = sql2[:len(sql2)-1]
			}

			// Wrap each SELECT in parentheses for clarity (optional but good practice)
			sql = fmt.Sprintf("(%s) %s (%s);", sql1, operator, sql2)
		}
	}

	return &CompoundSelectStmt{
		BaseStmt: NewBaseStmt(sql, stmtType, flavor),
	}, nil
}

// GenSelectUnion generates a UNION or UNION ALL statement.
func GenSelectUnion(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	return GenCompoundSelect(db, lcg, StmtSelectUnion, flavor)
}

// GenSelectIntersect generates an INTERSECT statement.
func GenSelectIntersect(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	return GenCompoundSelect(db, lcg, StmtSelectIntersect, flavor)
}

// GenSelectExcept generates an EXCEPT statement.
func GenSelectExcept(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	return GenCompoundSelect(db, lcg, StmtSelectExcept, flavor)
}
