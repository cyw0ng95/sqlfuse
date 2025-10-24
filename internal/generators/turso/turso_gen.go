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
	case r < 0:
		return "pragma"
	case r < 80:
		return "insert"
	case r < 100:
		return "select"
	}

	return "pragma" // fallback
}

// Generate produces a single SQL statement according to the chosen direction.
// If db is provided, can generate SELECTs using schema.
func (g *Generator) GenerateWithDB(db *sql.DB) string {
	dir := g.Direction()
	fmt.Println(dir)
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
	case "select":
		stmt, err := stmts.GenSelect(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT:", err)
			return "SELECT 1" // fallback
		}
		return stmt.SQL()
	}

	return "SELECT 1" // placeholder for other directions
}

// TokensUsed returns the number of tokens used by the underlying LCG.
func (g *Generator) TokensUsed() uint64 {
	return g.lcg.TokensUsed()
}
