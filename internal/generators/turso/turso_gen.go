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
	case r < 50:
		return "insert"
	case r < 74:
		return "select_basic"
	case r < 86:
		return "select_where"
	case r < 92:
		return "select_like"
	case r < 96:
		return "create_table"
	case r < 98:
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
