package stmts

import (
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
)

// DescribeGenerator is a StmtGenerator for DESCRIBE statements.
type DescribeGenerator struct{}

// Generate implements StmtGenerator for DESCRIBE statements.
func (g *DescribeGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenDescribe(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. DESCRIBE can always be generated.
func (g *DescribeGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// DescribeStmt represents a DESCRIBE statement.
type DescribeStmt struct {
	*BaseStmt
}

// GenDescribe generates a DESCRIBE statement.
// DESCRIBE shows information about a table, view, or query.
// Reference: https://duckdb.org/docs/stable/sql/statements/describe
func GenDescribe(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	// DESCRIBE can be used on tables or queries
	choice := lcg.Intn(3)
	var sql string

	if choice == 0 {
		// DESCRIBE a SELECT query
		sql = "DESCRIBE SELECT 1 AS id, 'test' AS name;"
	} else if choice == 1 {
		// DESCRIBE a table (using a generic table name)
		tableName := fmt.Sprintf("table_%d", lcg.Uint64()%100)
		sql = fmt.Sprintf("DESCRIBE \"%s\";", tableName)
	} else {
		// DESCRIBE with a more complex query
		sql = "DESCRIBE SELECT * FROM (VALUES (1, 'a'), (2, 'b')) AS t(id, name);"
	}

	return &DescribeStmt{
		BaseStmt: NewBaseStmt(sql, "describe", flavor),
	}, nil
}

// ShowGenerator is a StmtGenerator for SHOW statements.
type ShowGenerator struct{}

// Generate implements StmtGenerator for SHOW statements.
func (g *ShowGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenShow(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. SHOW can always be generated.
func (g *ShowGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// ShowStmt represents a SHOW statement.
type ShowStmt struct {
	*BaseStmt
}

// GenShow generates a SHOW statement.
// SHOW displays various database metadata.
// Reference: https://duckdb.org/docs/stable/sql/statements/show
func GenShow(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	// Various SHOW commands
	commands := []string{
		"SHOW TABLES;",
		"SHOW ALL TABLES;",
		"SHOW DATABASES;",
		"SHOW;", // Shows all settings
	}

	sql := commands[lcg.Intn(len(commands))]

	return &ShowStmt{
		BaseStmt: NewBaseStmt(sql, "show", flavor),
	}, nil
}

// SummarizeGenerator is a StmtGenerator for SUMMARIZE statements.
type SummarizeGenerator struct{}

// Generate implements StmtGenerator for SUMMARIZE statements.
func (g *SummarizeGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenSummarize(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. SUMMARIZE can always be generated.
func (g *SummarizeGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// SummarizeStmt represents a SUMMARIZE statement.
type SummarizeStmt struct {
	*BaseStmt
}

// GenSummarize generates a SUMMARIZE statement.
// SUMMARIZE provides a quick statistical summary of data.
// Reference: https://duckdb.org/docs/stable/sql/statements/summarize
func GenSummarize(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	// SUMMARIZE can be used on tables or queries
	choice := lcg.Intn(3)
	var sql string

	if choice == 0 {
		// SUMMARIZE a SELECT query
		sql = "SUMMARIZE SELECT 1 AS id, 2.5 AS value, 'test' AS name;"
	} else if choice == 1 {
		// SUMMARIZE a table (using a generic table name)
		tableName := fmt.Sprintf("table_%d", lcg.Uint64()%100)
		sql = fmt.Sprintf("SUMMARIZE \"%s\";", tableName)
	} else {
		// SUMMARIZE with a VALUES clause
		sql = "SUMMARIZE SELECT * FROM (VALUES (1, 10.5), (2, 20.5), (3, 30.5)) AS t(id, value);"
	}

	return &SummarizeStmt{
		BaseStmt: NewBaseStmt(sql, "summarize", flavor),
	}, nil
}
