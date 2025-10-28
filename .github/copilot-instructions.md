# SQLsmith-Go AI Coding Instructions

## Project Overview
This is a Go implementation of SQLsmith - a SQL query generator/fuzzer for testing database systems. The project focuses on Turso/LibSQL database integration and SQL query execution testing.

## Architecture & Key Components

### Database Integration
- **Primary database**: Turso LibSQL (SQLite-compatible) via `github.com/tursodatabase/go-libsql`
- **SQL parsing**: Uses ANTLR4-based SQLite parser (`github.com/libsql/sqlite-antlr4-parser`)
- **Connection pattern**: Standard `database/sql` interface with `libsql` driver

### Project Structure
```
cmd/executors/          # Database executor implementations
└── turso_embedded.go   # Turso LibSQL executor (main entry point)
```

## Development Patterns

### Database Connection Pattern
Always use this connection pattern for LibSQL:
```go
dbName := "file:./local.db"
db, err := sql.Open("libsql", dbName)
// Always defer db.Close() and handle connection errors with os.Exit(1)
```

### Query Execution Pattern
Follow the established pattern in `cmd/executors/turso_embedded.go`:
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
- Build: `go build ./cmd/executors/turso_embedded.go`
- Run: `./turso_embedded` (creates local.db automatically)
- The binary name `turso_embedded` is gitignored
- When forming git patch, donot include any binary built

### Container Development
Use the provided Containerfile for consistent environment:
```bash
podman build -t sqlsmith-go .
# Containerfile uses Fedora 43 with Go, git, and make
```

### Dependencies
- Go 1.24.9+ required
- ANTLR4 Go runtime for SQL parsing
- Turso LibSQL driver (latest from main branch)

## File Naming & Organization

### Executors
- Place database-specific executors in `cmd/executors/`
- Name pattern: `{database}_embedded.go` for embedded databases
- Each executor should be a standalone `main` package

### Database Files
- Local SQLite files use `.db` extension
- Default local database: `local.db` (gitignored)
- Test databases should follow `*.db` pattern (all gitignored)

## Key Integration Points

### SQL Generation (Future)
- Will integrate with ANTLR4 SQLite parser for query generation
- Parser dependency already included: `github.com/libsql/sqlite-antlr4-parser`

### Multi-Database Support (Future)
- Executor pattern allows easy addition of new database backends
- Each database gets its own executor in `cmd/executors/`

## Testing Considerations
- Test against local SQLite files (gitignored)
- Focus on SQL compatibility between different LibSQL versions
- Consider fuzzing with generated SQL queries once generator is implemented
