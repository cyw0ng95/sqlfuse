package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
)

// ExplainStmt represents an EXPLAIN statement.
type ExplainStmt struct {
	sql string
}

func (s *ExplainStmt) SQL() string  { return s.sql }
func (s *ExplainStmt) Type() string { return "explain" }

// GenExplain generates an EXPLAIN or EXPLAIN QUERY PLAN statement
// wrapping another SQL statement.
// According to Turso COMPAT.md: Yes (full support).
func GenExplain(db *sql.DB, lcg *common.LCG) (Stmt, error) {
	if lcg == nil {
		lcg = common.NewLCG(1)
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
	return &ExplainStmt{sql: sql}, nil
}

// GenExplainQueryPlan generates an EXPLAIN QUERY PLAN statement specifically.
func GenExplainQueryPlan(db *sql.DB, lcg *common.LCG) (Stmt, error) {
	if lcg == nil {
		lcg = common.NewLCG(1)
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
	return &ExplainStmt{sql: sql}, nil
}
