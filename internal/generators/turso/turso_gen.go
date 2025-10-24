package turso

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/generators/turso/stmts"
)

// Generator uses an LCG to drive generation directions and produce SQL snippets.
type Generator struct {
	lcg   *common.LCG
	first bool // first generation is forced into pragma
}

// NewGenerator creates a generator seeded with the provided seed.
func NewGenerator(seed uint64) *Generator {
	return &Generator{
		lcg:   common.NewLCG(seed),
		first: true,
	}
}

// Direction picks a direction to drive generation.
// On the very first call this is 100% "pragma". Afterwards it uses the LCG
// to pick between pragma, ddl, and dml.
func (g *Generator) Direction() string {
	if g.first {
		g.first = false
		return "pragma"
	}
	r := g.lcg.Intn(100)
	switch {
	case r < 40:
		return "insert"
	case r < 55:
		return "select_basic"
	case r < 70:
		return "select_where"
	case r < 77:
		return "select_like"
	case r < 82:
		return "select_limit"
	case r < 86:
		return "select_order"
	case r < 89:
		return "select_group"
	case r < 91:
		return "select_having"
	case r < 93:
		return "create_table"
	case r < 94:
		return "drop_table"
	case r < 100:
		return "alter_table"
	default:
		return "pragma"
	}
}

// Generate produces a single SQL statement according to the chosen direction.
// If db is provided, can generate SELECTs using schema.
func (g *Generator) GenerateWithDB(db *sql.DB) string {
	dir := g.Direction()
	switch dir {
	case "pragma":
		return stmts.GenPragma(g.lcg).SQL()
	case "insert":
		stmt, err := stmts.GenInsert(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating INSERT:", err)
			return "INSERT INTO sqlite_master DEFAULT VALUES;" // fallback
		}
		return stmt.SQL()
	case "select_basic":
		stmt, err := stmts.GenSelect(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT:", err)
			return "SELECT 1" // fallback
		}
		return stmt.SQL()
	case "select_where":
		stmtW, err := stmts.GenSelectWhere(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT WHERE:", err)
			return "SELECT 1"
		}
		return stmtW.SQL()
	case "select_like":
		stmtL, err := stmts.GenSelectWhereLike(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT LIKE:", err)
			return "SELECT 1"
		}
		return stmtL.SQL()
	case "select_limit":
		stmtLim, err := stmts.GenSelectLimit(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT LIMIT:", err)
			return "SELECT 1"
		}
		return stmtLim.SQL()
	case "select_order":
		stmtOrd, err := stmts.GenSelectOrderBy(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT ORDER BY:", err)
			return "SELECT 1"
		}
		return stmtOrd.SQL()
	case "select_group":
		stmtG, err := stmts.GenSelectGroupBy(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT GROUP BY:", err)
			return "SELECT 1"
		}
		return stmtG.SQL()
	case "select_having":
		stmtH, err := stmts.GenSelectHaving(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT HAVING:", err)
			return "SELECT 1"
		}
		return stmtH.SQL()
	case "create_table":
		stmt, err := stmts.GenCreateTable(g.lcg)
		if err != nil {
			fmt.Println("Error generating CREATE TABLE:", err)
			return "CREATE TABLE IF NOT EXISTS fallback (id INTEGER);" // fallback
		}
		return stmt.SQL()
	case "drop_table":
		stmt, err := stmts.GenDropTable(g.lcg)
		if err != nil {
			fmt.Println("Error generating DROP TABLE:", err)
			return "DROP TABLE IF EXISTS fallback;"
		}
		return stmt.SQL()
	case "alter_table":
		stmt, err := stmts.GenAlterTable(g.lcg)
		if err != nil {
			fmt.Println("Error generating ALTER TABLE:", err)
			return "ALTER TABLE fallback RENAME TO fallback2;"
		}
		return stmt.SQL()
	}

	return "SELECT 1" // placeholder for other directions
}

// TokensUsed returns the number of tokens used by the underlying LCG.
func (g *Generator) TokensUsed() uint64 {
	return g.lcg.TokensUsed()
}
