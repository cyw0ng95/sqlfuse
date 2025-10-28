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

// GenPragma generates a PRAGMA statement using the default flavor.
// The set of pragmas generated depends on the database flavor:
// - Turso/LibSQL: Only pragmas with "Yes" or "Partial" support
//   (https://github.com/tursodatabase/turso/blob/main/COMPAT.md#pragma)
// - go-sqlite3: Full SQLite3 pragma support
//   (https://sqlite.org/pragma.html)
func GenPragma(lcg *common.LCG) Stmt {
	return genPragmaWithFlavor(lcg, GetDefaultFlavor())
}

// genPragmaWithFlavor generates a PRAGMA statement with flavor support.
// It generates different sets of pragmas based on the database flavor (Turso vs go-sqlite3).
func genPragmaWithFlavor(lcg *common.LCG, flavor FlavorConfig) Stmt {
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	var pragmas []string
	
	// Choose pragma list based on flavor
	// For go-sqlite3, use full SQLite3 support
	// For Turso, or default/unknown flavors, use Turso-compatible subset
	if flavor.Name() == "go-sqlite3" {
		// Full SQLite3 pragma support for go-sqlite3
		// Based on https://sqlite.org/pragma.html
		pragmas = []string{
			// Commonly used pragmas
			"application_id",
			"auto_vacuum",
			"automatic_index",
			"busy_timeout",
			"cache_size",
			"cache_spill",
			"case_sensitive_like",
			"cell_size_check",
			"checkpoint_fullfsync",
			"collation_list",
			"compile_options",
			"database_list",
			"encoding",
			"foreign_keys",
			"freelist_count",
			"fullfsync",
			"ignore_check_constraints",
			"incremental_vacuum",
			"integrity_check",
			"journal_mode",
			"journal_size_limit",
			"legacy_alter_table",
			"legacy_file_format",
			"locking_mode",
			"max_page_count",
			"mmap_size",
			"page_count",
			"page_size",
			"parser_trace",
			"pragma_list",
			"query_only",
			"quick_check",
			"read_uncommitted",
			"recursive_triggers",
			"reverse_unordered_selects",
			"schema_version",
			"secure_delete",
			"soft_heap_limit",
			"synchronous",
			"table_info",
			"temp_store",
			"threads",
			"trusted_schema",
			"user_version",
			"wal_autocheckpoint",
			"wal_checkpoint",
			"writable_schema",
		}
	} else {
		// Turso/LibSQL or default: Only include PRAGMAs with "Yes" or "Partial" support
		// Based on https://github.com/tursodatabase/turso/blob/main/COMPAT.md#pragma
		// This is the conservative default for unknown flavors
		pragmas = []string{
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
	}

	p := pragmas[lcg.Intn(len(pragmas))]
	return generatePragmaSQL(p, lcg, flavor)
}

// generatePragmaSQL generates SQL for a specific pragma with appropriate values.
func generatePragmaSQL(pragma string, lcg *common.LCG, flavor FlavorConfig) *PragmaStmt {
	var sql string
	
	switch pragma {
	// Integer value pragmas
	case "application_id", "cache_size", "max_page_count", "page_size", "schema_version", "user_version":
		num := 1 + lcg.Intn(10000)
		sql = fmt.Sprintf("PRAGMA %s = %d;", pragma, num)
	
	case "busy_timeout":
		timeout := lcg.Intn(60000) // 0-60 seconds in milliseconds
		sql = fmt.Sprintf("PRAGMA busy_timeout = %d;", timeout)
	
	case "cache_spill":
		if lcg.Intn(2) == 0 {
			sql = fmt.Sprintf("PRAGMA cache_spill = %d;", 1+lcg.Intn(1000))
		} else {
			sql = "PRAGMA cache_spill = -1;" // Use default
		}
	
	case "journal_size_limit":
		var limit int
		if lcg.Intn(10) == 0 { // 10% chance to disable the limit
			limit = -1
		} else {
			limit = lcg.Intn(10000000) // A positive value
		}
		sql = fmt.Sprintf("PRAGMA journal_size_limit = %d;", limit)
	
	case "mmap_size":
		size := lcg.Intn(1073741824) // 0 to 1GB
		sql = fmt.Sprintf("PRAGMA mmap_size = %d;", size)
	
	case "soft_heap_limit":
		limit := lcg.Intn(100000000) // 0 to 100MB
		sql = fmt.Sprintf("PRAGMA soft_heap_limit = %d;", limit)
	
	case "threads":
		threads := lcg.Intn(8) // 0-7 threads
		sql = fmt.Sprintf("PRAGMA threads = %d;", threads)
	
	case "wal_autocheckpoint":
		pages := 100 + lcg.Intn(10000)
		sql = fmt.Sprintf("PRAGMA wal_autocheckpoint = %d;", pages)
	
	// String/encoding pragmas
	case "encoding":
		encodings := []string{"'UTF-8'", "'UTF-16'", "'UTF-16le'", "'UTF-16be'"}
		v := encodings[lcg.Intn(len(encodings))]
		sql = fmt.Sprintf("PRAGMA encoding = %s;", v)
	
	// Boolean/mode pragmas
	case "automatic_index", "case_sensitive_like", "cell_size_check", "checkpoint_fullfsync",
		"foreign_keys", "fullfsync", "ignore_check_constraints", "legacy_alter_table",
		"legacy_file_format", "query_only", "read_uncommitted", "recursive_triggers",
		"reverse_unordered_selects", "secure_delete", "trusted_schema", "writable_schema":
		values := []string{"ON", "OFF"}
		v := values[lcg.Intn(len(values))]
		sql = fmt.Sprintf("PRAGMA %s = %s;", pragma, v)
	
	// Journal mode
	case "journal_mode":
		if flavor.Name() != "go-sqlite3" {
			sql = "PRAGMA journal_mode = WAL;" // Turso/default only supports WAL
		} else {
			modes := []string{"DELETE", "TRUNCATE", "PERSIST", "MEMORY", "WAL", "OFF"}
			mode := modes[lcg.Intn(len(modes))]
			sql = fmt.Sprintf("PRAGMA journal_mode = %s;", mode)
		}
	
	// Synchronous mode
	case "synchronous":
		if flavor.Name() != "go-sqlite3" {
			values := []string{"OFF", "FULL"} // Turso/default only supports OFF and FULL
			v := values[lcg.Intn(len(values))]
			sql = fmt.Sprintf("PRAGMA synchronous = %s;", v)
		} else {
			// go-sqlite3 supports: OFF, NORMAL, FULL, EXTRA
			values := []string{"OFF", "NORMAL", "FULL", "EXTRA"}
			v := values[lcg.Intn(len(values))]
			sql = fmt.Sprintf("PRAGMA synchronous = %s;", v)
		}
	
	// Auto-vacuum mode
	case "auto_vacuum":
		modes := []string{"NONE", "FULL", "INCREMENTAL"}
		mode := modes[lcg.Intn(len(modes))]
		sql = fmt.Sprintf("PRAGMA auto_vacuum = %s;", mode)
	
	// Locking mode
	case "locking_mode":
		modes := []string{"NORMAL", "EXCLUSIVE"}
		mode := modes[lcg.Intn(len(modes))]
		sql = fmt.Sprintf("PRAGMA locking_mode = %s;", mode)
	
	// Temp store mode
	case "temp_store":
		modes := []string{"DEFAULT", "FILE", "MEMORY"}
		mode := modes[lcg.Intn(len(modes))]
		sql = fmt.Sprintf("PRAGMA temp_store = %s;", mode)
	
	// Incremental vacuum
	case "incremental_vacuum":
		pages := lcg.Intn(1000)
		if pages == 0 {
			sql = "PRAGMA incremental_vacuum;"
		} else {
			sql = fmt.Sprintf("PRAGMA incremental_vacuum(%d);", pages)
		}
	
	// WAL checkpoint
	case "wal_checkpoint":
		if flavor.Name() != "go-sqlite3" {
			sql = "PRAGMA wal_checkpoint;" // Turso/default doesn't support modes
		} else {
			modes := []string{"PASSIVE", "FULL", "RESTART", "TRUNCATE"}
			mode := modes[lcg.Intn(len(modes))]
			sql = fmt.Sprintf("PRAGMA wal_checkpoint(%s);", mode)
		}
	
	// Table info - needs a table name (use common table names)
	case "table_info":
		// Turso/default: PRAGMA table_info; (no table name)
		// go-sqlite3: PRAGMA table_info(table_name);
		if flavor.Name() != "go-sqlite3" {
			sql = "PRAGMA table_info;"
		} else {
			tables := []string{"t1", "t2", "users", "items"}
			table := tables[lcg.Intn(len(tables))]
			sql = fmt.Sprintf("PRAGMA table_info(%s);", table)
		}
	
	// Read-only query pragmas
	case "integrity_check", "freelist_count", "page_count", "pragma_list", "database_list",
		"quick_check", "collation_list", "compile_options", "parser_trace":
		sql = fmt.Sprintf("PRAGMA %s;", pragma)
	
	default:
		sql = fmt.Sprintf("PRAGMA %s;", pragma)
	}
	
	return &PragmaStmt{sql: sql, flavor: flavor}
}
