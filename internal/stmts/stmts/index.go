package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/stmts/helper"
	"strings"
)

// CreateIndexGenerator is a StmtGenerator for CREATE INDEX statements.
type CreateIndexGenerator struct{}

// Generate implements StmtGenerator for CREATE INDEX statements.
func (g *CreateIndexGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenCreateIndex(ctx.DB, ctx.LCG)
}

// CanGenerate implements StmtGenerator. CREATE INDEX can always be generated.
func (g *CreateIndexGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// DropIndexGenerator is a StmtGenerator for DROP INDEX statements.
type DropIndexGenerator struct{}

// Generate implements StmtGenerator for DROP INDEX statements.
func (g *DropIndexGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenDropIndex(ctx.LCG)
}

// CanGenerate implements StmtGenerator. DROP INDEX can always be generated.
func (g *DropIndexGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// CreateIndexStmt represents a CREATE INDEX statement.
type CreateIndexStmt struct {
	sql    string
	flavor FlavorConfig
}

func (s *CreateIndexStmt) SQL() string          { return s.sql }
func (s *CreateIndexStmt) Type() string         { return "create_index" }
func (s *CreateIndexStmt) Flavor() FlavorConfig { return s.flavor }

// DropIndexStmt represents a DROP INDEX statement.
type DropIndexStmt struct {
	sql    string
	flavor FlavorConfig
}

func (s *DropIndexStmt) SQL() string          { return s.sql }
func (s *DropIndexStmt) Type() string         { return "drop_index" }
func (s *DropIndexStmt) Flavor() FlavorConfig { return s.flavor }

// GenCreateIndex generates a CREATE INDEX statement.
// According to Turso COMPAT.md: Partial support (only for columns, not arbitrary expressions).
func GenCreateIndex(db *sql.DB, lcg *common.LCG) (Stmt, error) {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}

	// Try to get actual tables from schema (only if db is not nil)
	var tables []helper.TableInfo
	var err error
	if db != nil {
		tables, err = helper.GetAllTablesAndCols(db, "sqlite")
	}
	
	if db == nil || err != nil || len(tables) == 0 {
		// Fallback to simple index without schema
		return genCreateIndexFallback(lcg), nil
	}

	rnd := lcg.Intn
	tbl := tables[rnd(len(tables))]

	if len(tbl.Cols) == 0 {
		return genCreateIndexFallback(lcg), nil
	}

	// Generate index name
	indexName := fmt.Sprintf("idx_%s_%d", tbl.Name, lcg.Uint64()%100000)

	// Choose 1..min(3, len(cols)) columns for the index
	maxCols := len(tbl.Cols)
	if maxCols > 3 {
		maxCols = 3
	}
	nCols := 1 + rnd(maxCols)
	if nCols > len(tbl.Cols) {
		nCols = len(tbl.Cols)
	}

	// Select columns for index (avoiding duplicates)
	selected := make(map[int]struct{}, nCols)
	indexCols := make([]string, 0, nCols)
	for len(indexCols) < nCols {
		idx := rnd(len(tbl.Cols))
		if _, ok := selected[idx]; ok {
			continue
		}
		selected[idx] = struct{}{}

		col := tbl.Cols[idx]
		// Turso supports column references only, not arbitrary expressions
		// Optionally add ASC/DESC and COLLATE
		colSpec := quoteIdent(col.Name)

		// 30% chance to add sort order
		if rnd(10) < 3 {
			if rnd(2) == 0 {
				colSpec += " ASC"
			} else {
				colSpec += " DESC"
			}
		}

		indexCols = append(indexCols, colSpec)
	}

	// Decide on UNIQUE and IF NOT EXISTS modifiers
	unique := ""
	if rnd(3) == 0 { // 33% chance of UNIQUE
		unique = "UNIQUE "
	}

	ifNotExists := ""
	if rnd(2) == 0 { // 50% chance of IF NOT EXISTS
		ifNotExists = "IF NOT EXISTS "
	}

	sql := fmt.Sprintf("CREATE %sINDEX %s%s ON %s (%s);",
		unique, ifNotExists, quoteIdent(indexName), quoteIdent(tbl.Name), strings.Join(indexCols, ", "))
	return &CreateIndexStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}

// genCreateIndexFallback generates a simple CREATE INDEX without schema information
func genCreateIndexFallback(lcg *common.LCG) *CreateIndexStmt {
	tbl := fmt.Sprintf("tbl_%d", lcg.Uint64()%1000000)
	idx := fmt.Sprintf("idx_%d", lcg.Uint64()%1000000)
	col := fmt.Sprintf("col%d", 1+lcg.Intn(6))

	unique := ""
	if lcg.Intn(3) == 0 {
		unique = "UNIQUE "
	}

	sql := fmt.Sprintf("CREATE %sINDEX IF NOT EXISTS \"%s\" ON \"%s\" (\"%s\");", unique, idx, tbl, col)
	return &CreateIndexStmt{sql: sql, flavor: GetDefaultFlavor()}
}

// GenDropIndex generates a DROP INDEX statement.
// According to Turso COMPAT.md: Partial support (disabled by default).
func GenDropIndex(lcg *common.LCG) (Stmt, error) {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}

	// Generate index name (matching naming scheme from GenCreateIndex)
	indexName := fmt.Sprintf("idx_%d", lcg.Uint64()%100000)

	// Always use IF EXISTS to avoid errors
	sql := fmt.Sprintf("DROP INDEX IF EXISTS \"%s\";", indexName)
	return &DropIndexStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}
