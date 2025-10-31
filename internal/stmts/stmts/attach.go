package stmts

import (
	"fmt"
	"sqlfuse/internal/common"
)

// AttachGenerator is a StmtGenerator for ATTACH DATABASE statements.
type AttachGenerator struct{}

// Generate implements StmtGenerator for ATTACH DATABASE statements.
func (g *AttachGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenAttachDatabase(ctx.LCG), nil
}

// CanGenerate implements StmtGenerator. ATTACH DATABASE can always be generated.
func (g *AttachGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// DetachGenerator is a StmtGenerator for DETACH DATABASE statements.
type DetachGenerator struct{}

// Generate implements StmtGenerator for DETACH DATABASE statements.
func (g *DetachGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenDetachDatabase(ctx.LCG), nil
}

// CanGenerate implements StmtGenerator. DETACH DATABASE can always be generated.
func (g *DetachGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// AttachStmt represents an ATTACH DATABASE statement.
// It embeds BaseStmt to avoid boilerplate method implementations.
type AttachStmt struct {
	*BaseStmt
}

// DetachStmt represents a DETACH DATABASE statement.
// It embeds BaseStmt to avoid boilerplate method implementations.
type DetachStmt struct {
	*BaseStmt
}

// GenAttachDatabase generates an ATTACH DATABASE statement.
// According to Turso COMPAT.md: Partial support (only for reads, modifications will fail).
func GenAttachDatabase(lcg *common.LCG) Stmt {
	lcg = ensureLCG(lcg)

	// Generate a database file path and alias
	// Use temporary in-memory or file-based databases
	dbChoice := lcg.Intn(3)
	var dbPath string

	switch dbChoice {
	case 0:
		// In-memory database
		dbPath = ":memory:"
	case 1:
		// Empty database (will be created)
		dbPath = ""
	default:
		// File-based database with unique name
		dbPath = fmt.Sprintf("attached_%d.db", lcg.Uint64()%100000)
	}

	// Generate database alias
	alias := fmt.Sprintf("db_%d", lcg.Uint64()%1000)

	// ATTACH DATABASE 'path' AS alias
	sql := fmt.Sprintf("ATTACH DATABASE '%s' AS \"%s\";", escapeSingleQuote(dbPath), alias)
	return &AttachStmt{
		BaseStmt: NewBaseStmt(sql, "attach", GetDefaultFlavor()),
	}
}

// GenDetachDatabase generates a DETACH DATABASE statement.
// According to Turso COMPAT.md: Yes (full support).
func GenDetachDatabase(lcg *common.LCG) Stmt {
	lcg = ensureLCG(lcg)

	// Generate database alias matching the naming scheme from GenAttachDatabase
	alias := fmt.Sprintf("db_%d", lcg.Uint64()%1000)

	// DETACH DATABASE can omit the DATABASE keyword
	variants := []string{
		fmt.Sprintf("DETACH DATABASE \"%s\";", alias),
		fmt.Sprintf("DETACH \"%s\";", alias),
	}

	sql := variants[lcg.Intn(len(variants))]
	return &DetachStmt{
		BaseStmt: NewBaseStmt(sql, "detach", GetDefaultFlavor()),
	}
}

// escapeSingleQuote escapes single quotes in strings for SQL literals
func escapeSingleQuote(s string) string {
	result := ""
	for _, c := range s {
		if c == '\'' {
			result += "''"
		} else {
			result += string(c)
		}
	}
	return result
}
