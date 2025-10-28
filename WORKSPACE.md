# Go Workspace Structure

This project uses Go workspaces to split dependencies across different executors and components.

## Workspace Layout

The workspace is defined in `go.work` and consists of 5 modules:

```
sqlsmith-go/
├── go.work                          # Workspace definition
├── internal/                        # Shared internal packages
│   ├── go.mod
│   └── go.sum
├── cmd/
│   ├── executors/
│   │   ├── turso_embedded/          # Turso LibSQL executor
│   │   │   ├── go.mod
│   │   │   ├── go.sum
│   │   │   └── main.go
│   │   ├── go_sqlite3_embedded/     # go-sqlite3 executor
│   │   │   ├── go.mod
│   │   │   ├── go.sum
│   │   │   └── main.go
│   │   └── chai_embedded/           # Chai SQL executor
│   │       ├── go.mod
│   │       ├── go.sum
│   │       └── main.go
│   └── server/                      # HTTP API server
│       ├── go.mod
│       ├── go.sum
│       ├── main.go
│       └── job_ctrl.go
```

## Module Dependencies

### internal module (`sqlsmith-go/internal`)
Shared code used by all executors and server. Contains:
- `common/` - Logger, LCG random number generator
- `executors/` - Shared executor logic
- `generators/` - SQL generation framework and Turso generator
- `stmts/` - SQL statement builders and parsers

**Direct dependencies:**
- `github.com/antlr4-go/antlr/v4` - Parser runtime
- `github.com/libsql/sqlite-antlr4-parser` - SQLite parser
- `github.com/mattn/go-isatty` - Terminal detection
- `github.com/rs/zerolog` - Logging
- `github.com/spf13/cobra` - CLI framework

**Test dependencies:**
- `github.com/tursodatabase/turso-go` - Used in test files only

### turso_embedded executor (`sqlsmith-go/cmd/executors/turso_embedded`)
Fuzzer for Turso LibSQL database.

**Direct dependencies:**
- `github.com/tursodatabase/turso-go` - Turso database driver
- `github.com/spf13/cobra` - CLI framework
- `sqlsmith-go/internal` - Shared code (via replace directive)

**Binary includes:** ONLY turso-go (verified with `go version -m`)

### go_sqlite3_embedded executor (`sqlsmith-go/cmd/executors/go_sqlite3_embedded`)
Fuzzer for SQLite via go-sqlite3.

**Direct dependencies:**
- `github.com/mattn/go-sqlite3` - SQLite database driver
- `github.com/spf13/cobra` - CLI framework
- `sqlsmith-go/internal` - Shared code (via replace directive)

**Binary includes:** ONLY go-sqlite3 (verified with `go version -m`)

### chai_embedded executor (`sqlsmith-go/cmd/executors/chai_embedded`)
Fuzzer for Chai SQL database (driver currently commented out).

**Direct dependencies:**
- `github.com/spf13/cobra` - CLI framework
- `sqlsmith-go/internal` - Shared code (via replace directive)

### server (`sqlsmith-go/cmd/server`)
HTTP API server for managing fuzzing jobs.

**Direct dependencies:**
- `github.com/labstack/echo/v4` - HTTP framework
- `sqlsmith-go/internal` - Shared code (via replace directive)

## Benefits of Workspace Structure

1. **Dependency Isolation**: Each executor only includes its required database driver
   - turso_embedded: 155MB (includes turso-go)
   - go_sqlite3_embedded: 8.4MB (includes go-sqlite3)
   - server: 11MB (no database drivers)

2. **Code Reuse**: Shared internal packages are in one place, referenced via replace directives

3. **Independent Versioning**: Each module can update its dependencies independently

4. **Build Optimization**: Building one executor doesn't require dependencies for others

5. **Clear Boundaries**: Module structure makes it clear which code depends on what

## Building

### Quick Start with Makefile (Recommended for faster builds):
The Makefile provides parallel compilation, optimized caching, and incremental builds.

```bash
# Fast development build (with caching)
make dev-build

# Production build (optimized, smaller binaries)
make release-build

# Run tests with coverage
make test

# See all available targets
make help
```

See [BUILD_OPTIMIZATION.md](BUILD_OPTIMIZATION.md) for detailed performance benchmarks, optimization strategies, and troubleshooting.

### Using build.sh (Container-based):
```bash
export SQLSMITH_GO_CONTAINER_TYPE=test
bash build.sh
```

This builds:
- `output/turso_embedded_executor`
- `output/go_sqlite3_embedded_executor`
- `output/server`

### Build individual modules:
```bash
# Build turso executor
cd cmd/executors/turso_embedded && go build -trimpath -o ../../../output/turso_embedded_executor .

# Build go-sqlite3 executor
cd cmd/executors/go_sqlite3_embedded && go build -trimpath -o ../../../output/go_sqlite3_embedded_executor .

# Build server
cd cmd/server && go build -trimpath -o ../../output/server .
```

## Testing

### Quick test with Makefile:
```bash
# Run all tests with coverage
make test

# Verbose test output
make test-verbose
```

### Manual testing:
Tests are located in the `internal` module:
```bash
cd internal && go test ./...
```

Or use the build script:
```bash
export SQLSMITH_GO_CONTAINER_TYPE=test
bash build.sh --test
```

## Adding a New Executor

1. Create new directory: `cmd/executors/my_executor/`
2. Create `go.mod`:
   ```go
   module sqlsmith-go/cmd/executors/my_executor
   
   go 1.24.9
   
   require (
       github.com/my/database-driver vX.Y.Z
       github.com/spf13/cobra v1.10.1
       sqlsmith-go/internal v0.0.0
   )
   
   replace sqlsmith-go/internal => ../../../internal
   ```
3. Create `main.go` (see existing executors for examples)
4. Add to `go.work`:
   ```
   use ./cmd/executors/my_executor
   ```
5. Add to `build.sh`:
   ```bash
   (cd cmd/executors/my_executor && go_build ../../../output/my_executor .)
   ```

## Troubleshooting

### "module not found" errors
Run `go mod tidy` in the affected module directory.

### Changes in internal not reflected
The workspace uses replace directives, so changes in `internal/` are immediately visible to all modules.

### go.work.sum conflicts
This file is in `.gitignore` and is auto-generated. Each developer's workspace may have different versions.
