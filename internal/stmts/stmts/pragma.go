package stmts

import (
	"fmt"
	"sqlsmith-go/internal/common"
)

// PragmaValueGenerator is a function that generates the value part of a PRAGMA SQL statement.
// It returns the complete SQL string for the pragma.
type PragmaValueGenerator func(pragma string, lcg *common.LCG, flavor FlavorConfig) string

// PragmaDefinition defines a PRAGMA statement including which flavors support it
// and how to generate its value.
type PragmaDefinition struct {
	Name             string               // PRAGMA name
	SupportedFlavors []string             // List of flavors that support this pragma (e.g., ["go-sqlite3", "turso"])
	GenerateValue    PragmaValueGenerator // Function to generate the complete SQL
}

// PragmaGenerator is a StmtGenerator for PRAGMA statements.
type PragmaGenerator struct{}

// pragmaDefinitions is a map of all pragma definitions by name.
// This separates the configuration from the generation logic.
var pragmaDefinitions = map[string]PragmaDefinition{
	// Integer value pragmas
	"application_id": {
		Name:             "application_id",
		SupportedFlavors: []string{"go-sqlite3", "turso", "sqlite"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			num := 1 + lcg.Intn(10000)
			return fmt.Sprintf("PRAGMA %s = %d;", pragma, num)
		},
	},
	"cache_size": {
		Name:             "cache_size",
		SupportedFlavors: []string{"go-sqlite3", "turso", "sqlite"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			num := 1 + lcg.Intn(10000)
			return fmt.Sprintf("PRAGMA %s = %d;", pragma, num)
		},
	},
	"max_page_count": {
		Name:             "max_page_count",
		SupportedFlavors: []string{"go-sqlite3", "turso", "sqlite"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			num := 1 + lcg.Intn(10000)
			return fmt.Sprintf("PRAGMA %s = %d;", pragma, num)
		},
	},
	"page_size": {
		Name:             "page_size",
		SupportedFlavors: []string{"go-sqlite3", "turso", "sqlite"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			num := 1 + lcg.Intn(10000)
			return fmt.Sprintf("PRAGMA %s = %d;", pragma, num)
		},
	},
	"schema_version": {
		Name:             "schema_version",
		SupportedFlavors: []string{"go-sqlite3", "turso", "sqlite"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			num := 1 + lcg.Intn(10000)
			return fmt.Sprintf("PRAGMA %s = %d;", pragma, num)
		},
	},
	"user_version": {
		Name:             "user_version",
		SupportedFlavors: []string{"go-sqlite3", "turso", "sqlite"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			num := 1 + lcg.Intn(10000)
			return fmt.Sprintf("PRAGMA %s = %d;", pragma, num)
		},
	},
	"busy_timeout": {
		Name:             "busy_timeout",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			timeout := lcg.Intn(60000) // 0-60 seconds in milliseconds
			return fmt.Sprintf("PRAGMA busy_timeout = %d;", timeout)
		},
	},
	"cache_spill": {
		Name:             "cache_spill",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			if lcg.Intn(2) == 0 {
				return fmt.Sprintf("PRAGMA cache_spill = %d;", 1+lcg.Intn(1000))
			}
			return "PRAGMA cache_spill = -1;" // Use default
		},
	},
	"journal_size_limit": {
		Name:             "journal_size_limit",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			var limit int
			if lcg.Intn(10) == 0 { // 10% chance to disable the limit
				limit = -1
			} else {
				limit = lcg.Intn(10000000) // A positive value
			}
			return fmt.Sprintf("PRAGMA journal_size_limit = %d;", limit)
		},
	},
	"mmap_size": {
		Name:             "mmap_size",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			size := lcg.Intn(1073741824) // 0 to 1GB
			return fmt.Sprintf("PRAGMA mmap_size = %d;", size)
		},
	},
	"soft_heap_limit": {
		Name:             "soft_heap_limit",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			limit := lcg.Intn(100000000) // 0 to 100MB
			return fmt.Sprintf("PRAGMA soft_heap_limit = %d;", limit)
		},
	},
	"threads": {
		Name:             "threads",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			threads := lcg.Intn(8) // 0-7 threads
			return fmt.Sprintf("PRAGMA threads = %d;", threads)
		},
	},
	"wal_autocheckpoint": {
		Name:             "wal_autocheckpoint",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			pages := 100 + lcg.Intn(10000)
			return fmt.Sprintf("PRAGMA wal_autocheckpoint = %d;", pages)
		},
	},
	// String/encoding pragmas
	"encoding": {
		Name:             "encoding",
		SupportedFlavors: []string{"go-sqlite3", "turso", "sqlite"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			encodings := []string{"'UTF-8'", "'UTF-16'", "'UTF-16le'", "'UTF-16be'"}
			v := encodings[lcg.Intn(len(encodings))]
			return fmt.Sprintf("PRAGMA encoding = %s;", v)
		},
	},
	// Boolean/mode pragmas
	"automatic_index": {
		Name:             "automatic_index",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			values := []string{"ON", "OFF"}
			v := values[lcg.Intn(len(values))]
			return fmt.Sprintf("PRAGMA %s = %s;", pragma, v)
		},
	},
	"case_sensitive_like": {
		Name:             "case_sensitive_like",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			values := []string{"ON", "OFF"}
			v := values[lcg.Intn(len(values))]
			return fmt.Sprintf("PRAGMA %s = %s;", pragma, v)
		},
	},
	"cell_size_check": {
		Name:             "cell_size_check",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			values := []string{"ON", "OFF"}
			v := values[lcg.Intn(len(values))]
			return fmt.Sprintf("PRAGMA %s = %s;", pragma, v)
		},
	},
	"checkpoint_fullfsync": {
		Name:             "checkpoint_fullfsync",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			values := []string{"ON", "OFF"}
			v := values[lcg.Intn(len(values))]
			return fmt.Sprintf("PRAGMA %s = %s;", pragma, v)
		},
	},
	"foreign_keys": {
		Name:             "foreign_keys",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			values := []string{"ON", "OFF"}
			v := values[lcg.Intn(len(values))]
			return fmt.Sprintf("PRAGMA %s = %s;", pragma, v)
		},
	},
	"fullfsync": {
		Name:             "fullfsync",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			values := []string{"ON", "OFF"}
			v := values[lcg.Intn(len(values))]
			return fmt.Sprintf("PRAGMA %s = %s;", pragma, v)
		},
	},
	"ignore_check_constraints": {
		Name:             "ignore_check_constraints",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			values := []string{"ON", "OFF"}
			v := values[lcg.Intn(len(values))]
			return fmt.Sprintf("PRAGMA %s = %s;", pragma, v)
		},
	},
	"legacy_alter_table": {
		Name:             "legacy_alter_table",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			values := []string{"ON", "OFF"}
			v := values[lcg.Intn(len(values))]
			return fmt.Sprintf("PRAGMA %s = %s;", pragma, v)
		},
	},
	"legacy_file_format": {
		Name:             "legacy_file_format",
		SupportedFlavors: []string{"go-sqlite3", "turso", "sqlite"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			values := []string{"ON", "OFF"}
			v := values[lcg.Intn(len(values))]
			return fmt.Sprintf("PRAGMA %s = %s;", pragma, v)
		},
	},
	"query_only": {
		Name:             "query_only",
		SupportedFlavors: []string{"go-sqlite3", "turso", "sqlite"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			values := []string{"ON", "OFF"}
			v := values[lcg.Intn(len(values))]
			return fmt.Sprintf("PRAGMA %s = %s;", pragma, v)
		},
	},
	"read_uncommitted": {
		Name:             "read_uncommitted",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			values := []string{"ON", "OFF"}
			v := values[lcg.Intn(len(values))]
			return fmt.Sprintf("PRAGMA %s = %s;", pragma, v)
		},
	},
	"recursive_triggers": {
		Name:             "recursive_triggers",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			values := []string{"ON", "OFF"}
			v := values[lcg.Intn(len(values))]
			return fmt.Sprintf("PRAGMA %s = %s;", pragma, v)
		},
	},
	"reverse_unordered_selects": {
		Name:             "reverse_unordered_selects",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			values := []string{"ON", "OFF"}
			v := values[lcg.Intn(len(values))]
			return fmt.Sprintf("PRAGMA %s = %s;", pragma, v)
		},
	},
	"secure_delete": {
		Name:             "secure_delete",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			values := []string{"ON", "OFF"}
			v := values[lcg.Intn(len(values))]
			return fmt.Sprintf("PRAGMA %s = %s;", pragma, v)
		},
	},
	"trusted_schema": {
		Name:             "trusted_schema",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			values := []string{"ON", "OFF"}
			v := values[lcg.Intn(len(values))]
			return fmt.Sprintf("PRAGMA %s = %s;", pragma, v)
		},
	},
	"writable_schema": {
		Name:             "writable_schema",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			values := []string{"ON", "OFF"}
			v := values[lcg.Intn(len(values))]
			return fmt.Sprintf("PRAGMA %s = %s;", pragma, v)
		},
	},
	// Journal mode
	"journal_mode": {
		Name:             "journal_mode",
		SupportedFlavors: []string{"go-sqlite3", "turso", "sqlite"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			if flavor.Name() != "go-sqlite3" {
				return "PRAGMA journal_mode = WAL;" // Turso/default only supports WAL
			}
			modes := []string{"DELETE", "TRUNCATE", "PERSIST", "MEMORY", "WAL", "OFF"}
			mode := modes[lcg.Intn(len(modes))]
			return fmt.Sprintf("PRAGMA journal_mode = %s;", mode)
		},
	},
	// Synchronous mode
	"synchronous": {
		Name:             "synchronous",
		SupportedFlavors: []string{"go-sqlite3", "turso", "sqlite"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			if flavor.Name() != "go-sqlite3" {
				values := []string{"OFF", "FULL"} // Turso/default only supports OFF and FULL
				v := values[lcg.Intn(len(values))]
				return fmt.Sprintf("PRAGMA synchronous = %s;", v)
			}
			// go-sqlite3 supports: OFF, NORMAL, FULL, EXTRA
			values := []string{"OFF", "NORMAL", "FULL", "EXTRA"}
			v := values[lcg.Intn(len(values))]
			return fmt.Sprintf("PRAGMA synchronous = %s;", v)
		},
	},
	// Auto-vacuum mode
	"auto_vacuum": {
		Name:             "auto_vacuum",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			modes := []string{"NONE", "FULL", "INCREMENTAL"}
			mode := modes[lcg.Intn(len(modes))]
			return fmt.Sprintf("PRAGMA auto_vacuum = %s;", mode)
		},
	},
	// Locking mode
	"locking_mode": {
		Name:             "locking_mode",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			modes := []string{"NORMAL", "EXCLUSIVE"}
			mode := modes[lcg.Intn(len(modes))]
			return fmt.Sprintf("PRAGMA locking_mode = %s;", mode)
		},
	},
	// Temp store mode
	"temp_store": {
		Name:             "temp_store",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			modes := []string{"DEFAULT", "FILE", "MEMORY"}
			mode := modes[lcg.Intn(len(modes))]
			return fmt.Sprintf("PRAGMA temp_store = %s;", mode)
		},
	},
	// Incremental vacuum
	"incremental_vacuum": {
		Name:             "incremental_vacuum",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			pages := lcg.Intn(1000)
			if pages == 0 {
				return "PRAGMA incremental_vacuum;"
			}
			return fmt.Sprintf("PRAGMA incremental_vacuum(%d);", pages)
		},
	},
	// WAL checkpoint
	"wal_checkpoint": {
		Name:             "wal_checkpoint",
		SupportedFlavors: []string{"go-sqlite3", "turso", "sqlite"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			if flavor.Name() != "go-sqlite3" {
				return "PRAGMA wal_checkpoint;" // Turso/default doesn't support modes
			}
			modes := []string{"PASSIVE", "FULL", "RESTART", "TRUNCATE"}
			mode := modes[lcg.Intn(len(modes))]
			return fmt.Sprintf("PRAGMA wal_checkpoint(%s);", mode)
		},
	},
	// Table info - needs a table name (use common table names)
	"table_info": {
		Name:             "table_info",
		SupportedFlavors: []string{"go-sqlite3", "turso", "sqlite"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			// Turso/default: PRAGMA table_info; (no table name)
			// go-sqlite3: PRAGMA table_info(table_name);
			if flavor.Name() != "go-sqlite3" {
				return "PRAGMA table_info;"
			}
			tables := []string{"t1", "t2", "users", "items"}
			table := tables[lcg.Intn(len(tables))]
			return fmt.Sprintf("PRAGMA table_info(%s);", table)
		},
	},
	// Read-only query pragmas
	"integrity_check": {
		Name:             "integrity_check",
		SupportedFlavors: []string{"go-sqlite3", "turso", "sqlite"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			return fmt.Sprintf("PRAGMA %s;", pragma)
		},
	},
	"freelist_count": {
		Name:             "freelist_count",
		SupportedFlavors: []string{"go-sqlite3", "turso", "sqlite"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			return fmt.Sprintf("PRAGMA %s;", pragma)
		},
	},
	"page_count": {
		Name:             "page_count",
		SupportedFlavors: []string{"go-sqlite3", "turso", "sqlite"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			return fmt.Sprintf("PRAGMA %s;", pragma)
		},
	},
	"pragma_list": {
		Name:             "pragma_list",
		SupportedFlavors: []string{"go-sqlite3", "turso", "sqlite"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			return fmt.Sprintf("PRAGMA %s;", pragma)
		},
	},
	"database_list": {
		Name:             "database_list",
		SupportedFlavors: []string{"go-sqlite3", "turso", "sqlite"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			return fmt.Sprintf("PRAGMA %s;", pragma)
		},
	},
	"quick_check": {
		Name:             "quick_check",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			return fmt.Sprintf("PRAGMA %s;", pragma)
		},
	},
	"collation_list": {
		Name:             "collation_list",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			return fmt.Sprintf("PRAGMA %s;", pragma)
		},
	},
	"compile_options": {
		Name:             "compile_options",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			return fmt.Sprintf("PRAGMA %s;", pragma)
		},
	},
	"parser_trace": {
		Name:             "parser_trace",
		SupportedFlavors: []string{"go-sqlite3"},
		GenerateValue: func(pragma string, lcg *common.LCG, flavor FlavorConfig) string {
			return fmt.Sprintf("PRAGMA %s;", pragma)
		},
	},
}

// getSupportedPragmas returns a list of pragma names supported by the given flavor.
// For unknown flavors, it returns the conservative set (sqlite/turso pragmas).
// For DuckDB, it returns an empty list since DuckDB uses SET instead of PRAGMA.
// The returned list is sorted to ensure deterministic behavior.
func getSupportedPragmas(flavor FlavorConfig) []string {
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	flavorName := flavor.Name()

	// DuckDB uses SET for configuration, not PRAGMA
	// Return empty list to avoid executing SQLite-style PRAGMA statements
	if flavorName == "duckdb" {
		return []string{}
	}

	var supported []string

	// Collect supported pragmas
	for name, def := range pragmaDefinitions {
		for _, supportedFlavor := range def.SupportedFlavors {
			if supportedFlavor == flavorName {
				supported = append(supported, name)
				break
			}
		}
	}

	// If no pragmas were found (unknown flavor), return conservative set
	if len(supported) == 0 {
		// Use sqlite/turso conservative set for unknown flavors
		for name, def := range pragmaDefinitions {
			for _, supportedFlavor := range def.SupportedFlavors {
				if supportedFlavor == "sqlite" {
					supported = append(supported, name)
					break
				}
			}
		}
	}

	// Sort to ensure deterministic ordering
	// This is important for LCG-based random selection
	sortPragmaNames(supported)

	return supported
}

// sortPragmaNames sorts pragma names in place to ensure deterministic behavior.
// Using a simple selection sort to avoid importing sort package.
func sortPragmaNames(names []string) {
	for i := 0; i < len(names); i++ {
		minIdx := i
		for j := i + 1; j < len(names); j++ {
			if names[j] < names[minIdx] {
				minIdx = j
			}
		}
		names[i], names[minIdx] = names[minIdx], names[i]
	}
}

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
//   - Turso/LibSQL: Only pragmas with "Yes" or "Partial" support
//     (https://github.com/tursodatabase/turso/blob/main/COMPAT.md#pragma)
//   - go-sqlite3: Full SQLite3 pragma support
//     (https://sqlite.org/pragma.html)
func GenPragma(lcg *common.LCG) Stmt {
	return genPragmaWithFlavor(lcg, GetDefaultFlavor())
}

// genPragmaWithFlavor generates a PRAGMA statement with flavor support.
// It generates different sets of pragmas based on the database flavor (Turso vs go-sqlite3).
// For flavors that don't support PRAGMA (like DuckDB), returns a no-op statement.
func genPragmaWithFlavor(lcg *common.LCG, flavor FlavorConfig) Stmt {
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	// Get pragmas supported by this flavor
	pragmas := getSupportedPragmas(flavor)

	// If no pragmas are supported (e.g., DuckDB), return a no-op comment
	if len(pragmas) == 0 {
		return &PragmaStmt{sql: "-- PRAGMA not supported for this flavor", flavor: flavor}
	}

	// Select a random pragma from the supported list
	p := pragmas[lcg.Intn(len(pragmas))]
	return generatePragmaSQL(p, lcg, flavor)
}

// generatePragmaSQL generates SQL for a specific pragma with appropriate values.
func generatePragmaSQL(pragma string, lcg *common.LCG, flavor FlavorConfig) *PragmaStmt {
	// Look up the pragma definition
	def, exists := pragmaDefinitions[pragma]
	if !exists {
		// Fallback for unknown pragmas
		return &PragmaStmt{sql: fmt.Sprintf("PRAGMA %s;", pragma), flavor: flavor}
	}

	// Use the definition's value generator
	sql := def.GenerateValue(pragma, lcg, flavor)
	return &PragmaStmt{sql: sql, flavor: flavor}
}
