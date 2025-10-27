package stmts

import (
	"fmt"
	"sqlsmith-go/internal/common"
)

// PragmaGenerator is a StmtGenerator for PRAGMA statements.
type PragmaGenerator struct{}

// Generate implements StmtGenerator for PRAGMA statements.
func (g *PragmaGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return genPragmaInternal(ctx.LCG), nil
}

// CanGenerate implements StmtGenerator. PRAGMA statements can always be generated.
func (g *PragmaGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// GenPragma generates a PRAGMA statement; values are chosen via the LCG.
// This function is kept for backward compatibility with existing code.
func GenPragma(lcg *common.LCG) Stmt {
	return genPragmaInternal(lcg)
}

// genPragmaInternal is the internal implementation used by both old and new interfaces.
func genPragmaInternal(lcg *common.LCG) Stmt {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}
	pragmas := []string{
		"application_id",
		"cache_size",
		"encoding",
		"freelist_count",
		"integrity_check",
		"journal_mode",
		"legacy_file_format",
		"max_page_count",
		"page_count",
		"page_size",
		"pragma_list",
		"query_only",
		"schema_version",
		"synchronous",
		"table_info",
		"user_version",
	}

	p := pragmas[lcg.Intn(len(pragmas))]
	var sql string
	switch p {
	case "application_id", "cache_size", "max_page_count", "page_size", "schema_version", "user_version":
		num := 1 + lcg.Intn(10000)
		sql = fmt.Sprintf("PRAGMA %s = %d;", p, num)
	case "encoding":
		encodings := []string{"'UTF-8'", "'UTF-16'", "'UTF-16le'", "'UTF-16be'"}
		v := encodings[lcg.Intn(len(encodings))]
		sql = fmt.Sprintf("PRAGMA %s = %s;", p, v)
	case "journal_mode":
		sql = "PRAGMA journal_mode = WAL;"
	case "synchronous":
		values := []string{"OFF", "FULL"}
		v := values[lcg.Intn(len(values))]
		sql = fmt.Sprintf("PRAGMA synchronous = %s;", v)
	case "query_only":
		values := []string{"ON", "OFF"}
		v := values[lcg.Intn(len(values))]
		sql = fmt.Sprintf("PRAGMA query_only = %s;", v)
	case "legacy_file_format":
		values := []string{"ON", "OFF"}
		v := values[lcg.Intn(len(values))]
		sql = fmt.Sprintf("PRAGMA legacy_file_format = %s;", v)
	case "integrity_check", "freelist_count", "page_count", "pragma_list", "table_info":
		sql = fmt.Sprintf("PRAGMA %s;", p)
	default:
		sql = fmt.Sprintf("PRAGMA %s;", p)
	}
	return &PragmaStmt{sql: sql}
}
