package other

import (
	"sqlfuse/internal/stmts/stmts"
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
)

// ExplainGenerator is a stmts.StmtGenerator for EXPLAIN statements.
type ExplainGenerator struct {
	variant stmts.StmtType
}

// Generate implements stmts.StmtGenerator for EXPLAIN statements.
func (g *ExplainGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	if g.variant == stmts.StmtExplainQueryPlan {
		return GenExplainQueryPlan(ctx.DB, ctx.LCG)
	}
	return GenExplain(ctx.DB, ctx.LCG)
}

// CanGenerate implements stmts.StmtGenerator. EXPLAIN can always be generated.
func (g *ExplainGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	return true
}

// ExplainStmt represents an EXPLAIN statement.
// It embeds stmts.BaseStmt to avoid boilerplate method implementations.
type ExplainStmt struct {
	*stmts.BaseStmt
}

// GenExplain generates an EXPLAIN or EXPLAIN QUERY PLAN statement
// wrapping another SQL statement.
// According to Turso COMPAT.md: Yes (full support).
func GenExplain(db *sql.DB, lcg *common.LCG) (stmts.Stmt, error) {
	lcg = stmts.EnsureLCG(lcg)

	// If no db is provided, generate a simple EXPLAIN for fallback
	if db == nil {
		sql := "EXPLAIN SELECT 1;"
		return &ExplainStmt{
			BaseStmt: stmts.NewBaseStmt(sql, "explain", stmts.GetDefaultFlavor()),
		}, nil
	}

	// Choose between EXPLAIN and EXPLAIN QUERY PLAN
	explainType := "EXPLAIN"
	if lcg.Intn(2) == 0 {
		explainType = "EXPLAIN QUERY PLAN"
	}

	// Generate a statement to explain
	// Choose from various statement types that EXPLAIN can wrap
	choice := lcg.Intn(5)
	var innerSQL string

	switch choice {
	case 0:
		// EXPLAIN SELECT
		innerStmt, err := GenSelect(db, lcg)
		if err != nil {
			innerSQL = "SELECT 1"
		} else {
			innerSQL = innerStmt.SQL()
		}
	case 1:
		// EXPLAIN UPDATE
		innerStmt, err := GenUpdate(db, lcg)
		if err != nil {
			innerSQL = "SELECT 1"
		} else {
			innerSQL = innerStmt.SQL()
		}
	case 2:
		// EXPLAIN DELETE
		innerStmt, err := GenDelete(db, lcg)
		if err != nil {
			innerSQL = "SELECT 1"
		} else {
			innerSQL = innerStmt.SQL()
		}
	case 3:
		// EXPLAIN INSERT
		innerStmt, err := GenInsert(db, lcg)
		if err != nil {
			innerSQL = "SELECT 1"
		} else {
			innerSQL = innerStmt.SQL()
		}
	default:
		// EXPLAIN a simple SELECT with WHERE
		innerStmt, err := GenSelect(db, lcg)
		if err != nil {
			innerSQL = "SELECT 1"
		} else {
			innerSQL = innerStmt.SQL()
		}
	}

	// Remove trailing semicolon if present, as we'll add it after EXPLAIN
	if len(innerSQL) > 0 && innerSQL[len(innerSQL)-1] == ';' {
		innerSQL = innerSQL[:len(innerSQL)-1]
	}

	sql := fmt.Sprintf("%s %s;", explainType, innerSQL)
	return &ExplainStmt{
		BaseStmt: stmts.NewBaseStmt(sql, "explain", stmts.GetDefaultFlavor()),
	}, nil
}

// GenExplainQueryPlan generates an EXPLAIN QUERY PLAN statement specifically.
func GenExplainQueryPlan(db *sql.DB, lcg *common.LCG) (stmts.Stmt, error) {
	lcg = stmts.EnsureLCG(lcg)

	// If no db is provided, generate a simple EXPLAIN for fallback
	if db == nil {
		sql := "EXPLAIN QUERY PLAN SELECT 1;"
		return &ExplainStmt{
			BaseStmt: stmts.NewBaseStmt(sql, "explain", stmts.GetDefaultFlavor()),
		}, nil
	}

	// Generate a SELECT statement to explain
	innerStmt, err := GenSelect(db, lcg)
	innerSQL := "SELECT 1"
	if err == nil {
		innerSQL = innerStmt.SQL()
	}

	if len(innerSQL) > 0 && innerSQL[len(innerSQL)-1] == ';' {
		innerSQL = innerSQL[:len(innerSQL)-1]
	}

	sql := fmt.Sprintf("EXPLAIN QUERY PLAN %s;", innerSQL)
	return &ExplainStmt{
		BaseStmt: stmts.NewBaseStmt(sql, "explain", stmts.GetDefaultFlavor()),
	}, nil
}
