# SQLsmith-Go AI Coding Instructions

## Project Overview
This is a Go implementation of SQLsmith - a SQL query generator/fuzzer for testing database systems. The project supports multiple SQLite-compatible database flavors with different feature sets and compatibility constraints.

## Architecture & Key Components

### Database Flavors
The project supports multiple database flavors, each with different SQL feature sets:
- **Turso LibSQL**: SQLite-compatible with some restrictions (see [COMPAT.md](https://github.com/tursodatabase/turso/blob/main/COMPAT.md))
- **go-sqlite3**: Full SQLite3 support via `github.com/mattn/go-sqlite3`
- **Default/Unknown**: Conservative mode (Turso-compatible subset for safety)

### Flavor Configuration
- Flavor configs are in `internal/generators/dialects/` directory
- Each flavor implements the `stmts.FlavorConfig` interface:
  - `Name() string`: Returns flavor identifier ("turso", "go-sqlite3", etc.)
  - `SupportsFeature(feature string) bool`: Feature capability checking
  - `ValidateSQL(sql string) error`: Optional SQL validation
- Statement generators use flavor configs to generate compatible SQL
- **Key principle**: When generating SQL statements or pragmas, always check the flavor to ensure compatibility

### Database Integration
- **Turso LibSQL**: SQLite-compatible via `github.com/tursodatabase/go-libsql`
- **go-sqlite3**: Full SQLite3 via `github.com/mattn/go-sqlite3`
- **SQL parsing**: Uses ANTLR4-based SQLite parser (`github.com/libsql/sqlite-antlr4-parser`)
- **Connection pattern**: Standard `database/sql` interface with flavor-specific drivers

### Project Structure
```
cmd/executors/               # Database executor implementations
├── turso_embedded/         # Turso LibSQL executor
├── go_sqlite3_embedded/    # go-sqlite3 executor
└── chai_embedded/          # Chai SQL executor
internal/
├── generators/             # SQL statement generators
│   ├── dialects/          # Flavor-specific configurations
│   │   ├── turso.go       # Turso LibSQL flavor config
│   │   └── go_sqlite3.go  # go-sqlite3 flavor config
│   ├── go_sqlite3.go      # go-sqlite3 generator
│   └── turso.go           # Turso generator (base)
└── stmts/stmts/           # Statement generation logic
    ├── pragma.go          # Flavor-aware PRAGMA generation
    └── ...                # Other statement types
```

## SQL Statement Generation with Flavors

### Flavor-Aware Statement Generation
When generating SQL statements (especially PRAGMA, DDL, or advanced SQL features):

1. **Check the flavor first**: Use `flavor.Name()` to identify the target database
2. **Use flavor feature checks**: Call `flavor.SupportsFeature(feature)` for conditional generation
3. **Generate appropriate SQL**: Adjust syntax, features, and parameters based on flavor

### PRAGMA Statement Generation
PRAGMA generation is flavor-aware (see `internal/stmts/stmts/pragma.go`):

- **Turso/Default flavor** (18 pragmas):
  - Limited to Turso-compatible pragmas ([reference](https://github.com/tursodatabase/turso/blob/main/COMPAT.md#pragma))
  - `journal_mode`: Only WAL supported
  - `synchronous`: Only OFF and FULL
  - `table_info`: No table name parameter
  - `wal_checkpoint`: No checkpoint mode parameter

- **go-sqlite3 flavor** (47 pragmas):
  - Full SQLite3 pragma support ([reference](https://sqlite.org/pragma.html))
  - All journal modes: DELETE, TRUNCATE, PERSIST, MEMORY, WAL, OFF
  - All synchronous modes: OFF, NORMAL, FULL, EXTRA
  - Extended pragmas: `auto_vacuum`, `foreign_keys`, `busy_timeout`, `mmap_size`, `threads`, etc.
  - Parameterized pragmas: `table_info(table_name)`, `wal_checkpoint(mode)`, etc.

### Adding Flavor Support to Statements
When implementing new statement generators:
```go
func genStatementWithFlavor(lcg *common.LCG, flavor FlavorConfig) Stmt {
    if flavor.Name() == "go-sqlite3" {
        // Generate with full SQLite3 features
    } else {
        // Generate conservative/Turso-compatible SQL
    }
}
```

## Development Patterns

### Database Connection Pattern
Always use the appropriate connection pattern for each database flavor:

**Turso LibSQL:**
```go
dbName := "file:./local.db"
db, err := sql.Open("libsql", dbName)
// Always defer db.Close() and handle connection errors with os.Exit(1)
```

**go-sqlite3:**
```go
dbName := "./local.db"  // or "file:./local.db"
db, err := sql.Open("sqlite3", dbName)
// Always defer db.Close() and handle connection errors with os.Exit(1)
```

### Query Execution Pattern
Follow the established pattern across all executors:
- Use `db.Query()` for SELECT statements
- Always defer `rows.Close()`
- Handle `rows.Err()` after iteration
- Print errors to `os.Stderr` and exit with status 1 for fatal errors

### Error Handling Convention
- Database connection errors: `fmt.Fprintf(os.Stderr, ...)` + `os.Exit(1)`
- Query execution errors: Print to stderr and exit
- Row scanning errors: Print to stdout and return (non-fatal)

## Build & Development Workflow

### Local Development
Each executor can be built independently:
- **Turso**: `go build ./cmd/executors/turso_embedded`
- **go-sqlite3**: `go build ./cmd/executors/go_sqlite3_embedded`
- Binary names are gitignored
- When forming git patches, do not include any binaries built

### Container Development
Use the provided Containerfile for consistent environment:
```bash
podman build -t sqlsmith-go .
# Containerfile uses Fedora 43 with Go, git, and make
```

### Dependencies
- Go 1.24.9+ required
- ANTLR4 Go runtime for SQL parsing
- Database drivers:
  - Turso LibSQL: `github.com/tursodatabase/turso-go` (latest from main branch)
  - go-sqlite3: `github.com/mattn/go-sqlite3`

## File Naming & Organization

### Executors
- Place database-specific executors in `cmd/executors/{flavor}_embedded/`
- Name pattern: `{flavor}_embedded` for embedded databases (e.g., `turso_embedded`, `go_sqlite3_embedded`)
- Each executor directory contains a `main.go` file with the standalone `main` package

### Flavor Configurations
- Place flavor configs in `internal/generators/dialects/`
- Name pattern: `{flavor}.go` (e.g., `turso.go`, `go_sqlite3.go`)
- Include tests: `{flavor}_test.go`
- Update README.md in dialects directory when adding new flavors

### Database Files
- Local SQLite files use `.db` extension
- Default local database: `local.db` (gitignored)
- Test databases should follow `*.db` pattern (all gitignored)

### Documentation Files
- All project documentation files should be placed in the `docs/` directory at the repository root
- This includes:
  - Project-level documentation (e.g., PRAGMA_SUPPORT.md, WORKSPACE.md, REFACTORING_SUMMARY.md)
  - Architecture and design documentation (e.g., ARCHITECTURE.md, DESIGN_PATTERNS.md)
  - Component-specific documentation that describes cross-cutting concerns
- Exception: READMEs that are part of specific component scaffolding (e.g., view/src/*/README.md) remain in their respective directories
- When creating new documentation, always place it in the `docs/` directory
- When referencing documentation in code or other docs, use paths relative to repository root (e.g., `docs/ARCHITECTURE.md`)
- **Do not write detailed documentation** - keep documentation concise and focused on essential information only

## Key Integration Points

### SQL Generation
- SQL generation is flavor-aware and adapts to database capabilities
- Uses ANTLR4 SQLite parser for validation (`github.com/libsql/sqlite-antlr4-parser`)
- Statement generators check flavor configs before generating SQL features
- Example: Window functions, CTEs, REGEXP - enabled for go-sqlite3, disabled for Turso

### Multi-Database Support
- The flavor system enables support for multiple SQLite-compatible databases
- Each database flavor has:
  1. Executor in `cmd/executors/{flavor}_embedded/`
  2. Generator in `internal/generators/{flavor}.go`
  3. Flavor config in `internal/generators/dialects/{flavor}.go`
- To add a new database flavor:
  1. Create flavor config implementing `stmts.FlavorConfig`
  2. Create generator extending `BaseGenerator`
  3. Create executor using the generator
  4. Add tests for flavor-specific behavior

## Testing Considerations
- Test against local SQLite files (gitignored)
- Test flavor-specific SQL generation:
  - Turso: Focus on compatibility constraints
  - go-sqlite3: Test full SQLite3 feature coverage
- Write flavor-aware tests that verify correct SQL generation per flavor
- Consider fuzzing with generated SQL queries for each flavor
