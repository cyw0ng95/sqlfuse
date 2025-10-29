# SQLsmith-Go

A high-performance SQL query generator and fuzzer for testing SQLite-compatible database systems. SQLsmith-Go generates syntactically valid, semantically interesting SQL statements to discover bugs, edge cases, and performance issues in database implementations.

## Overview

SQLsmith-Go is a Go implementation of the [SQLsmith](https://github.com/anse1/sqlsmith) approach to database testing through randomized query generation. Unlike traditional fuzzing that generates random bytes, SQLsmith-Go produces valid SQL statements that exercise diverse database features while respecting the constraints and capabilities of different SQLite flavors.

### Key Features

- **Multi-Flavor Support**: Generates SQL compatible with different database implementations (Turso LibSQL, go-sqlite3, DuckDB, Chai SQL)
- **Intelligent Generation**: Uses schema awareness to produce meaningful queries with valid table/column references
- **Comprehensive Coverage**: Supports diverse SQL features including CTEs, window functions, subqueries, and complex expressions
- **Flavor-Aware**: Adapts generated SQL to match the capabilities and constraints of the target database
- **Parallel Execution**: Multi-worker architecture for high-throughput fuzzing
- **Web Interface**: Vue.js-based frontend for monitoring and controlling fuzzing jobs
- **Modular Architecture**: Clean separation between generators, executors, and statement builders

## Architecture

### Component Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        SQLsmith-Go                           │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │  Executors   │───▶│  Generators  │───▶│  Statement   │  │
│  │              │    │              │    │  Builders    │  │
│  │ • Turso      │    │ • Base       │    │              │  │
│  │ • go-sqlite3 │    │ • Turso      │    │ • SELECT     │  │
│  │ • DuckDB     │    │ • go-sqlite3 │    │ • INSERT     │  │
│  │ • Chai       │    │ • DuckDB     │    │ • UPDATE     │  │
│  │ • HTTP API   │    │              │    │ • DELETE     │  │
│  └──────────────┘    └──────────────┘    │ • PRAGMA     │  │
│                                           │ • CREATE     │  │
│  │   Dialects   │    │   Frontend   │    │ • ...        │  │
│  │              │    │              │    └──────────────┘  │
│  │ • Feature    │    │ • Vue.js     │                      │
│  │   Detection  │    │ • Vuetify    │                      │
│  │ • SQL        │    │ • Job Ctrl   │                      │
│  │   Validation │    │              │                      │
│  └──────────────┘    └──────────────┘                      │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Design Patterns

**1. Flavor-Based Polymorphism**

The architecture uses dialect configurations (`FlavorConfig`) to adapt SQL generation to specific database capabilities:

```go
type FlavorConfig interface {
    Name() string
    SupportsFeature(feature string) bool
    ValidateSQL(sql string) error
}
```

Each database flavor (Turso, go-sqlite3) has its own configuration defining supported features, allowing the generator to produce only compatible SQL.

**2. Composition-Based Generators**

Generators use Go's embedding pattern to share common functionality while maintaining flavor-specific customization:

```go
type TursoGenerator struct {
    *BaseGenerator
    flavorConfig FlavorConfig
}
```

The `BaseGenerator` handles LCG-based randomness, statement selection, and recursion depth, while specific generators define statement weights and capabilities.

**3. Workspace-Based Dependency Isolation**

The project uses Go workspaces to ensure each executor includes only its required database driver:

- **turso_embedded**: 155MB (includes turso-go)
- **go_sqlite3_embedded**: 8.6MB (includes go-sqlite3 + CGo SQLite)
- **server**: 11MB (no database drivers)

This keeps binaries focused and reduces deployment size.

## Quick Start

### Prerequisites

- Go 1.24.9 or later
- C compiler (for go-sqlite3 CGo bindings)
- Optional: Node.js 24+ and pnpm (for web frontend)

### Building

```bash
# Set required environment variable
export SQLSMITH_GO_CONTAINER_TYPE=test

# Build all executors and server
bash build.sh

# Outputs:
# - output/turso_embedded_executor
# - output/go_sqlite3_embedded_executor
# - output/server
```

### Running Executors

**Turso LibSQL Executor:**

```bash
# In-memory fuzzing with default schema
./output/turso_embedded_executor --seed 42 --queries 100 --workers 4

# File-based database with custom schema
./output/turso_embedded_executor \
  --dsn "file:./test.db" \
  --init-sql "./assets/turso/init.sql" \
  --queries 1000 \
  --verbose
```

**go-sqlite3 Executor:**

```bash
# Full SQLite3 fuzzing with verbose output
./output/go_sqlite3_embedded_executor \
  --dsn "./sqlite3.db" \
  --seed 12345 \
  --queries 500 \
  --workers 8 \
  --verbose
```

**DuckDB Executor:**

```bash
# In-memory DuckDB fuzzing with analytical queries
./output/duckdb_embedded_executor \
  --dsn "" \
  --init-sql "./assets/duckdb/init.sql" \
  --queries 1000 \
  --workers 4 \
  --verbose

# File-based DuckDB database
./output/duckdb_embedded_executor \
  --dsn "test.duckdb" \
  --seed 42 \
  --queries 500
```

**HTTP Server:**

```bash
# Start the API server
./output/server

# Server listens on :8080 by default
# Configure via config/server.json
```

### Common Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--dsn` | Database connection string | `:memory:` |
| `--seed` | Random seed (0 = random) | `0` |
| `--queries` | Queries per worker | `10` |
| `--workers` | Concurrent workers | `1` |
| `--verbose` | Show executed SQL | `false` |
| `--init-sql` | Schema initialization file | `[flavor-specific]` |

## SQL Generation Strategy

### Randomness & Reproducibility

SQLsmith-Go uses a Linear Congruential Generator (LCG) for deterministic randomness:

- **Seeded**: All generation is reproducible given the same seed
- **Token-Based**: Tracks PRNG consumption for profiling
- **Stateless**: Each query generation is independent

### Statement Selection

Statements are selected based on weighted probabilities defined per flavor:

```go
// Example weights for Turso flavor
weights := map[StmtType]uint64{
    StmtSelectBasic:     100,
    StmtSelectJoin:      50,
    StmtSelectSubquery:  30,
    StmtInsert:          20,
    StmtUpdate:          10,
    StmtPragma:          5,
}
```

Higher weights = more frequent generation, allowing targeted stress testing of specific features.

### Schema-Aware Generation

When a database connection is provided, generators query the schema to produce realistic queries:

```sql
-- Generator queries schema
SELECT name FROM sqlite_master WHERE type='table';
PRAGMA table_info(users);

-- Then generates queries like
SELECT u.name, a.balance 
FROM users u 
INNER JOIN accounts a ON u.id = a.user_id 
WHERE a.balance > 100.5;
```

### Flavor-Specific Features

**Turso LibSQL** (18 PRAGMAs, conservative features):
```sql
PRAGMA journal_mode = WAL;        -- Only WAL supported
PRAGMA synchronous = FULL;         -- Only OFF/FULL
PRAGMA table_info;                 -- No table name parameter
```

**go-sqlite3** (47 PRAGMAs, full SQLite3):
```sql
PRAGMA journal_mode = DELETE;      -- All modes available
PRAGMA synchronous = NORMAL;       -- All modes available
PRAGMA table_info(users);          -- Parameterized
PRAGMA foreign_keys = ON;          -- Extended pragmas
PRAGMA auto_vacuum = INCREMENTAL;  -- Storage management
```

## Project Structure

```
sqlsmith-go/
├── cmd/
│   ├── executors/
│   │   ├── turso_embedded/      # Turso LibSQL executor
│   │   ├── go_sqlite3_embedded/ # go-sqlite3 executor
│   │   ├── duckdb_embedded/     # DuckDB executor
│   │   └── chai_embedded/       # Chai SQL executor
│   └── server/                  # HTTP API server
│
├── internal/
│   ├── common/                  # Logger, LCG utilities
│   ├── executors/               # Executor framework
│   ├── generators/              # SQL generators
│   │   ├── dialects/           # Flavor configurations
│   │   ├── base_generator.go  # Common generator logic
│   │   ├── turso.go           # Turso-specific generator
│   │   ├── go_sqlite3.go      # go-sqlite3 generator
│   │   └── duckdb.go          # DuckDB generator
│   └── stmts/
│       ├── stmts/              # Statement builders
│       └── types/              # SQL type generators
│
├── assets/                      # Database schemas & configs
│   ├── turso/init.sql
│   ├── go_sqlite3/init.sql
│   └── duckdb/init.sql
│
├── config/                      # Server & executor configs
├── docs/                        # Architecture documentation
├── view/                        # Vue.js web frontend
└── go.work                      # Go workspace definition
```

## Development

### Running Tests

```bash
# Run all tests with coverage
export SQLSMITH_GO_CONTAINER_TYPE=test
bash build.sh --test

# Coverage report: .cache/coverage.html
```

### Adding a New Database Flavor

1. **Create Flavor Configuration** (`internal/generators/dialects/mydb.go`):

```go
package dialects

type MyDBFlavorConfig struct{}

func (m *MyDBFlavorConfig) Name() string {
    return "mydb"
}

func (m *MyDBFlavorConfig) SupportsFeature(feature string) bool {
    switch feature {
    case "window_functions":
        return true
    case "regexp":
        return false
    default:
        return true
    }
}

func (m *MyDBFlavorConfig) ValidateSQL(sql string) error {
    return nil
}

func NewMyDBFlavorConfig() stmts.FlavorConfig {
    return &MyDBFlavorConfig{}
}
```

2. **Create Generator** (`internal/generators/mydb.go`):

```go
package generators

type MyDBGenerator struct {
    *BaseGenerator
    flavorConfig stmts.FlavorConfig
}

func NewMyDBGenerator(seed uint64) *MyDBGenerator {
    g := &MyDBGenerator{
        BaseGenerator: NewBaseGenerator(seed),
        flavorConfig:  dialects.NewMyDBFlavorConfig(),
    }
    g.SetWeights(DefaultMyDBStmtWeights())
    g.initGenMap()
    return g
}

func (g *MyDBGenerator) Name() string {
    return "mydb"
}
```

3. **Create Executor** (`cmd/executors/mydb_embedded/main.go`):

```go
package main

import (
    "database/sql"
    "sqlsmith-go/internal/executors"
    "sqlsmith-go/internal/generators"
    _ "github.com/mydb/driver"
)

func main() {
    // Standard executor pattern...
    executors.Run(
        "MyDB Executor",
        &flags,
        func(dsn string) (*sql.DB, error) {
            return sql.Open("mydb", dsn)
        },
        func(seed uint64) generators.Generator {
            return generators.NewMyDBGenerator(seed)
        },
        printSchema,
    )
}
```

4. **Add to Workspace** (`go.work`):

```go
use ./cmd/executors/mydb_embedded
```

### Code Organization Principles

- **Executors**: Database-specific; handle connection, execution, error reporting
- **Generators**: Flavor-specific; define weights and capabilities
- **Dialects**: Feature detection; define what SQL is valid for a flavor
- **Statement Builders**: Database-agnostic; generate SQL based on dialect constraints

## Use Cases

### 1. Bug Discovery

Find crashes, assertion failures, and incorrect results:

```bash
# High-volume fuzzing to find edge cases
./output/go_sqlite3_embedded_executor \
  --queries 10000 \
  --workers 16 \
  --seed 0
```

### 2. Regression Testing

Ensure database changes don't break existing behavior:

```bash
# Deterministic fuzzing with fixed seed
./output/turso_embedded_executor \
  --seed 42 \
  --queries 1000 \
  --verbose > baseline.log

# After database update, compare outputs
./output/turso_embedded_executor \
  --seed 42 \
  --queries 1000 \
  --verbose > updated.log

diff baseline.log updated.log
```

### 3. Performance Profiling

Identify slow queries and optimization opportunities:

```bash
# Generate workload with verbose output
./output/go_sqlite3_embedded_executor \
  --queries 5000 \
  --verbose | tee workload.sql

# Analyze with EXPLAIN QUERY PLAN
sqlite3 test.db < analyze.sql
```

### 4. Compatibility Testing

Verify SQL compatibility across flavors:

```bash
# Test Turso-compatible subset
./output/turso_embedded_executor --seed 100 --queries 1000

# Test full SQLite3 features
./output/go_sqlite3_embedded_executor --seed 100 --queries 1000
```

## API Server

The HTTP server provides a REST API for managing fuzzing jobs:

### Endpoints

- `GET /api/health` - Health check
- `GET /api/executors` - List available executors
- `POST /api/jobs` - Start a new fuzzing job
- `GET /api/jobs/:id` - Get job status
- `GET /api/jobs/:id/output` - Stream job output
- `DELETE /api/jobs/:id` - Cancel a job

### Example Usage

```bash
# Start the server
./output/server

# Create a fuzzing job
curl -X POST http://localhost:8080/api/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "executor": "turso",
    "seed": 42,
    "queries": 1000,
    "workers": 4
  }'

# Check job status
curl http://localhost:8080/api/jobs/1

# Stream output
curl http://localhost:8080/api/jobs/1/output
```

## Web Frontend

A Vue.js-based interface for interactive fuzzing across multiple database flavors:

### Features

- **Multi-Flavor Support**: Select from different database executors (Turso, go-sqlite3, Chai) with visual flavor indicators
- **Job Management**: Start, stop, and monitor fuzzing jobs with real-time status updates
- **Live Output**: View stdout/stderr output from running jobs
- **Executor Selection**: Dropdown shows both executor name and flavor (e.g., "turso_embedded (turso)")
- **Configuration**: Adjust seeds, query counts, worker counts via command-line arguments
- **Job History**: Track job IDs, status, and execution times

### Quick Start

1. **Start the backend server**:
   ```bash
   ./server  # or ./output/server
   ```

2. **Start the frontend dev server**:
   ```bash
   cd view
   pnpm install
   pnpm run dev
   ```

3. **Access the UI**: Navigate to `http://localhost:3000` (dev) or `http://localhost:5173` (depending on Vite version)

### Using Different Flavors

The UI displays available executors with their associated flavors:

![Executor Selection](https://github.com/user-attachments/assets/f35d7a04-a271-4195-95ed-af2ca1f973de)

When an executor is selected, its flavor is shown in parentheses:

![Selected Executor](https://github.com/user-attachments/assets/f42cb3b1-6179-49a8-8ea4-9af5407b3937)

**Example workflow**:
1. Select `turso_embedded (turso)` from the dropdown
2. Enter arguments: `--workers 4 --queries 100 --init-sql /path/to/schema.sql`
3. Click "Start Job" to begin fuzzing
4. Note the job ID and use Status/Info buttons to monitor progress

For detailed usage instructions, see [docs/WEB_UI_USAGE.md](docs/WEB_UI_USAGE.md).

### Development

```bash
cd view
pnpm install
pnpm run dev

# Frontend: http://localhost:3000
# Backend API: http://localhost:8080
```

### Production Build

```bash
cd view
pnpm run build

# Output: view/dist (served by server at /)
```

## Configuration

### Server Configuration (`config/server.json`)

```json
{
  "port": "8080",
  "executors_config_path": "./config/executors.json",
  "server_name": "sqlsmith-go server",
  "server_version": "0.1",
  "job": {
    "max_output_bytes": 65536,
    "persist_path": "./jobs"
  }
}
```

### Executors Configuration (`config/executors.json`)

```json
[
  {
    "executor": "turso_embedded",
    "path": "./output/turso_embedded",
    "flavor": "turso"
  },
  {
    "executor": "go_sqlite3_embedded",
    "path": "./output/go_sqlite3_embedded",
    "flavor": "go-sqlite3"
  },
  {
    "executor": "chai_embedded",
    "path": "./output/chai_embedded",
    "flavor": "chai"
  }
]
```

The `flavor` field is optional but recommended - it's displayed in the web UI to help users identify which database flavor each executor targets.

## Documentation

Comprehensive documentation is available in the `docs/` directory:

- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)**: Generator architecture and design patterns
- **[WORKSPACE.md](docs/WORKSPACE.md)**: Go workspace structure and module layout
- **[PRAGMA_SUPPORT.md](docs/PRAGMA_SUPPORT.md)**: Flavor-specific PRAGMA generation
- **[DIALECTS_README.md](docs/DIALECTS_README.md)**: Database dialect system
- **[DESIGN_PATTERNS.md](docs/DESIGN_PATTERNS.md)**: Code organization patterns
- **[DESIGN_IMPROVEMENTS.md](docs/DESIGN_IMPROVEMENTS.md)**: Planned enhancements
- **[WEB_UI_USAGE.md](docs/WEB_UI_USAGE.md)**: Web interface usage guide

## Performance Characteristics

### Throughput

On a typical workstation (8-core, 16GB RAM):

- **Single Worker**: ~500-1000 queries/second
- **8 Workers**: ~3000-5000 queries/second
- **Bottleneck**: Database execution (not generation)

### Memory Usage

- **Executor Process**: ~50-100MB per worker
- **Generator State**: ~1-5MB (schema cache + LCG state)
- **Database**: Depends on schema size and operations

### Scalability

Horizontal scaling via multiple executor instances:

```bash
# Terminal 1
./output/turso_embedded_executor --seed 1 --workers 8

# Terminal 2
./output/turso_embedded_executor --seed 1000 --workers 8

# Terminal 3
./output/turso_embedded_executor --seed 2000 --workers 8
```

## Contributing

Contributions are welcome! Areas of interest:

1. **New Database Flavors**: Add support for PostgreSQL, MySQL, etc.
2. **Statement Generators**: Implement more SQL features (WINDOW, PARTITION, etc.)
3. **Mutation Strategies**: Guided fuzzing based on code coverage
4. **Performance**: Optimize generation speed and memory usage
5. **Validation**: Enhanced SQL correctness checking

## License

[Add license information here]

## Related Projects

- **[SQLsmith](https://github.com/anse1/sqlsmith)**: Original C++ implementation for PostgreSQL
- **[SQLancer](https://github.com/sqlancer/sqlancer)**: Java-based database testing with logical validation
- **[go-fuzz](https://github.com/dvyukov/go-fuzz)**: Coverage-guided fuzzing for Go programs

## Acknowledgments

- Original SQLsmith approach by Andreas Seltenreich
- SQLite project for comprehensive SQL implementation
- Turso team for LibSQL compatibility documentation
- Go community for excellent database/sql abstraction
