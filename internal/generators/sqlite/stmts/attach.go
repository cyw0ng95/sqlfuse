package stmts

import (
	"fmt"
	"sqlsmith-go/internal/common"
)

// AttachStmt represents an ATTACH DATABASE statement.
type AttachStmt struct {
	sql string
}

func (s *AttachStmt) SQL() string  { return s.sql }
func (s *AttachStmt) Type() string { return "attach" }

// DetachStmt represents a DETACH DATABASE statement.
type DetachStmt struct {
	sql string
}

func (s *DetachStmt) SQL() string  { return s.sql }
func (s *DetachStmt) Type() string { return "detach" }

// GenAttachDatabase generates an ATTACH DATABASE statement.
// According to Turso COMPAT.md: Partial support (only for reads, modifications will fail).
func GenAttachDatabase(lcg *common.LCG) Stmt {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}

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
	return &AttachStmt{sql: sql}
}

// GenDetachDatabase generates a DETACH DATABASE statement.
// According to Turso COMPAT.md: Yes (full support).
func GenDetachDatabase(lcg *common.LCG) Stmt {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}

	// Generate database alias matching the naming scheme from GenAttachDatabase
	alias := fmt.Sprintf("db_%d", lcg.Uint64()%1000)

	// DETACH DATABASE can omit the DATABASE keyword
	variants := []string{
		fmt.Sprintf("DETACH DATABASE \"%s\";", alias),
		fmt.Sprintf("DETACH \"%s\";", alias),
	}

	sql := variants[lcg.Intn(len(variants))]
	return &DetachStmt{sql: sql}
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
