package turso

import (
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
	case r < 60:
		return "pragma"
	case r < 85:
		return "ddl"
	default:
		return "dml"
	}
}

// Generate produces a single SQL statement according to the chosen direction.
func (g *Generator) Generate() string {
	switch g.Direction() {
	case "pragma":
		return stmts.GenPragma(g.lcg).SQL()
	case "ddl":
		return g.genDDL()
	default:
		return g.genDML()
	}
}

// genDDL creates a simple CREATE TABLE statement driven by the LCG.
func (g *Generator) genDDL() string {
	id := g.lcg.Intn(100000)
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS t_%d (id INTEGER PRIMARY KEY, v TEXT);", id)
}

// genDML creates a simple INSERT statement driven by the LCG.
func (g *Generator) genDML() string {
	id := g.lcg.Intn(100000)
	val := g.lcg.Uint64()
	return fmt.Sprintf("INSERT INTO t_%d (v) VALUES ('x%016x');", id, val)
}
