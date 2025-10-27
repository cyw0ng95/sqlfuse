# Statement Generation Structure

This document describes the improved statement generation structure for SQLsmith-go's SQLite-compatible generators (Turso, ChaiSQL, go-sqlite3, etc.).

## Overview

The statement generation system has been refactored to provide:

1. **Unified Interface**: All generators implement the `StmtGenerator` interface
2. **Generator Registry**: Dynamic registration and lookup of generators
3. **Builder Pattern**: Fluent API for composing complex statements
4. **Context-Aware Generation**: `GenContext` provides recursion control and random number generation

## Core Components

### StmtGenerator Interface

All statement generators implement this interface:

```go
type StmtGenerator interface {
    // Generate creates a new statement using the provided context
    Generate(ctx *GenContext) (Stmt, error)
    
    // CanGenerate returns true if this generator can produce a valid statement
    // given the current context (e.g., requires tables to exist)
    CanGenerate(ctx *GenContext) bool
}
```

### GenContext

Provides context for statement generation:

```go
type GenContext struct {
    DB       *sql.DB      // Database connection for schema queries
    LCG      *common.LCG  // Random number generator
    Depth    int          // Current recursion depth
    MaxDepth int          // Maximum allowed recursion depth
}
```

Usage:
```go
ctx := stmts.NewGenContext(db, lcg, 2) // max depth = 2
stmt, err := generator.Generate(ctx)
```

### Generator Registry

Manage generators dynamically:

```go
// Get default registry with all standard generators
reg := stmts.DefaultRegistry()

// Register custom generator
reg.Register("my_custom", &MyCustomGenerator{})

// Get generator
gen := reg.Get("pragma")

// Check if exists
if reg.Has("pragma") {
    // ...
}

// List all generators
names := reg.Names()
```

### Convenience Function

Generate statements by name:

```go
ctx := stmts.NewGenContext(db, lcg, 2)
stmt, err := stmts.GenerateStmt(ctx, "pragma")
```

## Standard Generators

All standard generators are pre-registered in the default registry:

| Generator Name | Generator Type | Requires Tables | Description |
|---------------|----------------|-----------------|-------------|
| `pragma` | `PragmaGenerator` | No | PRAGMA statements |
| `insert` | `InsertGenerator` | Yes | Single row INSERT |
| `select` | `SelectGenerator` | No* | Basic SELECT |
| `update` | `UpdateGenerator` | No** | UPDATE statements |
| `delete` | `DeleteGenerator` | No** | DELETE statements |
| `create_table` | `CreateTableGenerator` | No | CREATE TABLE |
| `drop_table` | `DropTableGenerator` | No | DROP TABLE |
| `alter_table` | `AlterTableGenerator` | No | ALTER TABLE |
| `create_view` | `CreateViewGenerator` | No | CREATE VIEW |
| `drop_view` | `DropViewGenerator` | No | DROP VIEW |

*SELECT can generate without tables (returns `SELECT 1`)
**UPDATE/DELETE generate synthetic table names, don't require existing tables

## Builder Pattern

### SelectBuilder

Fluent interface for building SELECT statements:

```go
ctx := stmts.NewGenContext(db, lcg, 2)

// Basic SELECT
stmt, err := stmts.NewSelectBuilder(ctx).
    Select("id", "name", "email").
    From("users").
    Where("active = 1").
    Where("age > 18").  // Multiple WHERE conditions are AND-ed
    OrderBy("name ASC").
    Limit(10).
    Build()

// With GROUP BY and HAVING
stmt, err := stmts.NewSelectBuilder(ctx).
    Select("department", "COUNT(*)").
    From("employees").
    GroupBy("department").
    Having("COUNT(*) > 5").
    Build()

// With subquery
stmt, err := stmts.NewSelectBuilder(ctx).
    Select("*").
    FromSubquery("SELECT id, name FROM users WHERE active = 1", "active_users").
    Limit(10).
    Build()

// Random column selection from table
table := helper.TableInfo{Name: "users", Cols: [...]}
stmt, err := stmts.NewSelectBuilder(ctx).
    SelectColumns(table, 3).  // Select 3 random columns
    FromTable(table).
    Limit(10).
    Build()
```

### InsertBuilder

Fluent interface for building INSERT statements:

```go
// Single row
stmt, err := stmts.NewInsertBuilder(ctx).
    Into("users").
    Columns("name", "email").
    Values("'Alice'", "'alice@example.com'").
    Build()

// Multiple rows
stmt, err := stmts.NewInsertBuilder(ctx).
    Into("users").
    Columns("name", "email").
    Values("'Alice'", "'alice@example.com'").
    Values("'Bob'", "'bob@example.com'").
    Build()

// With ON CONFLICT
stmt, err := stmts.NewInsertBuilder(ctx).
    Into("users").
    Columns("email", "name").
    Values("'alice@example.com'", "'Alice'").
    OnConflict("ON CONFLICT(email) DO UPDATE SET name=excluded.name").
    Build()

// DEFAULT VALUES
stmt, err := stmts.NewInsertBuilder(ctx).
    Into("users").
    Build()  // Generates: INSERT INTO "users" DEFAULT VALUES;
```

### RandomSelectBuilder

Generate random SELECT statements using schema:

```go
ctx := stmts.NewGenContext(db, lcg, 2)
builder := stmts.NewRandomSelectBuilder(ctx, db)
stmt, err := builder.Build()

// Generates random SELECT with:
// - Random table from schema
// - 1-3 random columns
// - Optional WHERE clause
// - Optional ORDER BY
// - Random LIMIT 1-50
```

## Creating Custom Generators

### Using GeneratorFunc

For simple generators without preconditions:

```go
myGen := stmts.GeneratorFunc(func(ctx *stmts.GenContext) (stmts.Stmt, error) {
    sql := fmt.Sprintf("SELECT %d;", ctx.Intn(100))
    return &stmts.SelectStmt{sql: sql}, nil
})

// Register it
reg := stmts.DefaultRegistry()
reg.Register("my_simple", myGen)
```

### Using StmtGeneratorWithCheck

For generators with custom preconditions:

```go
myGen := &stmts.StmtGeneratorWithCheck{
    GenerateFn: func(ctx *stmts.GenContext) (stmts.Stmt, error) {
        // Your generation logic
        return &stmts.SelectStmt{sql: "SELECT 1;"}, nil
    },
    CanGenerateFn: func(ctx *stmts.GenContext) bool {
        // Custom precondition check
        return ctx.Depth == 0  // Only generate at depth 0
    },
}

reg.Register("my_conditional", myGen)
```

### Implementing StmtGenerator

For full control:

```go
type MyCustomGenerator struct {
    // Optional fields
}

func (g *MyCustomGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
    // Use ctx.LCG for randomness
    // Use ctx.DB for schema queries
    // Use ctx.Depth/MaxDepth for recursion control
    
    sql := fmt.Sprintf("SELECT %d;", ctx.Intn(100))
    return &stmts.SelectStmt{sql: sql}, nil
}

func (g *MyCustomGenerator) CanGenerate(ctx *stmts.GenContext) bool {
    // Return false if prerequisites aren't met
    return true
}

// Register
reg := stmts.DefaultRegistry()
reg.Register("my_custom", &MyCustomGenerator{})
```

## Backward Compatibility

All existing `Gen*()` functions remain available for backward compatibility:

```go
// Old way (still works)
stmt, err := stmts.GenPragma(lcg)
stmt, err := stmts.GenSelect(db, lcg)
stmt, err := stmts.GenInsert(db, lcg)

// New way (recommended)
ctx := stmts.NewGenContext(db, lcg, 2)
stmt, err := (&stmts.PragmaGenerator{}).Generate(ctx)
stmt, err := (&stmts.SelectGenerator{}).Generate(ctx)
stmt, err := (&stmts.InsertGenerator{}).Generate(ctx)

// Or via registry
stmt, err := stmts.GenerateStmt(ctx, "pragma")
```

## Best Practices

1. **Use GenContext**: Always create a `GenContext` for generation instead of passing raw LCG
2. **Check CanGenerate**: Before calling `Generate()`, check `CanGenerate()` to avoid errors
3. **Use Registry**: Use the registry for dynamic generator management
4. **Use Builders**: Use builders for complex statement composition
5. **Control Recursion**: Set appropriate `MaxDepth` in `GenContext` to prevent infinite recursion
6. **Validate SQL**: Use the ANTLR parser to validate generated SQL syntax

## Examples

### Basic Usage

```go
package main

import (
    "database/sql"
    "fmt"
    "sqlsmith-go/internal/common"
    "sqlsmith-go/internal/generators/sqlite/stmts"
    _ "github.com/tursodatabase/turso-go"
)

func main() {
    // Open database
    db, _ := sql.Open("turso", "file:test.db")
    defer db.Close()
    
    // Create context
    lcg := common.NewLCG(42)
    ctx := stmts.NewGenContext(db, lcg, 2)
    
    // Generate using registry
    stmt, _ := stmts.GenerateStmt(ctx, "select")
    fmt.Println(stmt.SQL())
    
    // Generate using builder
    stmt, _ = stmts.NewSelectBuilder(ctx).
        Select("*").
        From("users").
        Limit(10).
        Build()
    fmt.Println(stmt.SQL())
}
```

### Advanced Usage with Custom Generator

```go
type ComplexQueryGenerator struct{}

func (g *ComplexQueryGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
    // Use builder for composition
    builder := stmts.NewSelectBuilder(ctx)
    
    // Get tables from schema
    tables, _ := helper.GetAllTablesAndCols(ctx.DB)
    if len(tables) == 0 {
        return &stmts.SelectStmt{sql: "SELECT 1;"}, nil
    }
    
    // Pick random table
    table := tables[ctx.Intn(len(tables))]
    
    // Build complex query
    builder.FromTable(table).
        SelectColumns(table, 3)
    
    // Maybe add subquery in FROM clause
    if ctx.CanRecurse() && ctx.Intn(2) == 0 {
        subCtx := ctx.Descend()
        subGen := stmts.NewRandomSelectBuilder(subCtx, subCtx.DB)
        subStmt, _ := subGen.Build()
        builder.FromSubquery(subStmt.SQL(), "subq")
    }
    
    return builder.Build()
}

func (g *ComplexQueryGenerator) CanGenerate(ctx *stmts.GenContext) bool {
    return stmts.hasTables(ctx.DB)
}
```

## Migration Guide

If you're using the old generation interface, here's how to migrate:

### Before

```go
lcg := common.NewLCG(42)
stmt, err := stmts.GenSelect(db, lcg)
```

### After

```go
lcg := common.NewLCG(42)
ctx := stmts.NewGenContext(db, lcg, 2)
stmt, err := (&stmts.SelectGenerator{}).Generate(ctx)

// Or using registry
stmt, err := stmts.GenerateStmt(ctx, "select")

// Or using builder for more control
stmt, err := stmts.NewSelectBuilder(ctx).
    Select("id", "name").
    From("users").
    Limit(10).
    Build()
```

The old functions still work but the new interface provides:
- Better composability
- Recursion control
- Type safety
- More flexibility

## Testing

The new structure includes comprehensive tests:

- `stmt_test.go` - Tests for individual generators
- `registry_test.go` - Tests for registry functionality
- `builder_test.go` - Tests for builder pattern
- `recursive_test.go` - Tests for recursive generation

Run all tests:
```bash
go test ./internal/generators/sqlite/stmts/... -v
```

## Package Location

This package was moved from `internal/generators/turso/stmts` to `internal/generators/sqlite/stmts` to enable sharing statement generation logic across multiple SQLite-compatible database executors:

- **Turso/LibSQL**: Uses this package via `internal/generators/turso`
- **ChaiSQL**: Uses this package via the turso generator wrapper
- **go-sqlite3**: Can use this package for fuzzing
- **Other SQLite-compatible databases**: Can easily integrate

This allows multiple executors to leverage the same SQL generation infrastructure without duplication.
