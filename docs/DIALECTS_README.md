# Database Dialects

This directory contains database-specific flavor configurations that implement the `stmts.FlavorConfig` interface.

## Purpose

Dialect configurations define which SQL features are supported by a specific database implementation. This allows the SQL generator to produce only valid SQL for the target database.

## Structure

Each database flavor has its own file:
- `turso.go` - Turso LibSQL flavor configuration

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

See individual dialect files for complete lists of supported features.
