package stmts

import (
	"fmt"
	"sqlsmith-go/internal/common"
)

// PragmaGenerator is a StmtGenerator for PRAGMA statements.
type PragmaGenerator struct{}

// Generate implements StmtGenerator for PRAGMA statements.
func (g *PragmaGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return genPragmaWithFlavor(ctx.LCG, ctx.Flavor), nil
}

// CanGenerate implements StmtGenerator. PRAGMA can always be generated.
func (g *PragmaGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// GenPragma generates a PRAGMA statement compatible with Turso/LibSQL.
// Only includes PRAGMAs that are fully or partially supported according to
// https://github.com/tursodatabase/turso/blob/main/COMPAT.md#pragma
func GenPragma(lcg *common.LCG) Stmt {
	return genPragmaWithFlavor(lcg, GetDefaultFlavor())
}

// genPragmaWithFlavor generates a PRAGMA statement with flavor support.
func genPragmaWithFlavor(lcg *common.LCG, flavor FlavorConfig) Stmt {
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}
	// Only include PRAGMAs with "Yes" or "Partial" support in Turso
	pragmas := []string{
		"application_id",     // Yes
		"cache_size",         // Yes
		"database_list",      // Yes
		"encoding",           // Yes
		"freelist_count",     // Yes
		"integrity_check",    // Yes
		"journal_mode",       // Yes
		"legacy_file_format", // Yes
		"max_page_count",     // Yes
		"page_count",         // Yes
		"page_size",          // Yes
		"pragma_list",        // Yes
		"query_only",         // Yes
		"schema_version",     // Yes (write is noop in defensive mode)
		"synchronous",        // Partial (only OFF and FULL)
		"table_info",         // Yes
		"user_version",       // Yes
		"wal_checkpoint",     // Partial (without param)
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
	case "integrity_check", "freelist_count", "page_count", "pragma_list", "table_info", "database_list", "wal_checkpoint":
		sql = fmt.Sprintf("PRAGMA %s;", p)
	default:
		sql = fmt.Sprintf("PRAGMA %s;", p)
	}
	return &PragmaStmt{sql: sql, flavor: flavor}
}
