# Database Dialects

This directory contains database-specific flavor configurations that implement the `stmts.FlavorConfig` interface.

## Purpose

Dialect configurations define which SQL features are supported by a specific database implementation. This allows the SQL generator to produce only valid SQL for the target database.

## Structure

Each database flavor has its own file:
- `turso.go` - Turso LibSQL flavor configuration
- `go_sqlite3.go` - go-sqlite3 (full SQLite3) flavor configuration
- `duckdb.go` - DuckDB flavor configuration

## Adding a New Dialect

To add a new database dialect:

1. Create a new file named after your database (e.g., `postgres.go` or `mysql.go`)
2. Implement the `stmts.FlavorConfig` interface:
   ```go
   type PostgresFlavorConfig struct{}

   func (p *PostgresFlavorConfig) Name() string {
       return "postgres"
   }

   func (p *PostgresFlavorConfig) SupportsFeature(feature string) bool {
       // Define which features are supported
       return true
   }

   func (p *PostgresFlavorConfig) ValidateSQL(sql string) error {
       // Optional: validate SQL syntax
       return nil
   }

   func NewPostgresFlavorConfig() stmts.FlavorConfig {
       return &PostgresFlavorConfig{}
   }
   ```

3. Create tests in `postgres_test.go` to verify feature support
4. Use the dialect in your generator (see `../turso.go` for an example)

## Feature Flags

Common feature flags include:
- `exists_subquery` - NOT EXISTS (subquery) expressions
- `in_subquery` - IN (subquery) expressions
- `window_functions` - OVER (...) window functions
- `cte_recursive` - RECURSIVE keyword in WITH clause
- `regexp` - REGEXP operator
- `match` - MATCH operator
- `sqlite_pragma` - SQLite PRAGMA statements
- `named_transactions` - Named BEGIN/COMMIT transactions

See individual dialect files for complete lists of supported features.

## Supported Dialects

### Turso LibSQL (`turso.go`)
- **Description**: Turso LibSQL with SQLite compatibility constraints
- **Key Restrictions**: No window functions, no recursive CTEs, limited PRAGMA support
- **Reference**: [Turso Compatibility](https://github.com/tursodatabase/turso/blob/main/COMPAT.md)

### go-sqlite3 (`go_sqlite3.go`)
- **Description**: Full SQLite3 implementation via go-sqlite3
- **Key Features**: All SQLite3 features supported, extensive PRAGMA support
- **Reference**: [SQLite Documentation](https://sqlite.org/pragma.html)

### DuckDB (`duckdb.go`)
- **Description**: DuckDB in-process OLAP database
- **Key Features**: Advanced analytical features, window functions, CTEs, PostgreSQL-style syntax
- **Key Differences**: Uses SET instead of PRAGMA, information_schema instead of sqlite_master
- **Reference**: [DuckDB SQL Introduction](https://duckdb.org/docs/stable/sql/introduction)
