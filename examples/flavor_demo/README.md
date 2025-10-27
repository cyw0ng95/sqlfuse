# SQL Flavor Design Pattern - Example

This example demonstrates how the SQL flavor design pattern allows sqlsmith-go to generate appropriate SQL for different database flavors.

## What This Demonstrates

The flavor pattern solves the problem of different SQL databases having different feature sets:

- **Turso LibSQL**: Does not support window functions, recursive CTEs, or certain subquery patterns
- **Standard SQLite**: Supports all standard SQLite features including window functions and recursive CTEs
- **Future databases**: Can easily add PostgreSQL, MySQL, etc. with their own constraints

## Running the Example

```bash
go run examples/flavor_demo/main.go
```

## Expected Output

The demo shows:

1. **Default SQLite Flavor**: Supports all features including window functions
2. **Turso Flavor**: Shows which features are NOT supported (window functions, recursive CTEs, etc.)
3. **Recursive CTE Example**: Demonstrates how SQLite generates RECURSIVE CTEs
4. **Feature Matrix**: Side-by-side comparison of supported features

## Key Concepts

### FlavorConfig Interface

```go
type FlavorConfig interface {
    Name() string
    SupportsFeature(feature string) bool
    ValidateSQL(sql string) error
}
```

### Creating Flavor-Aware Context

```go
// Default SQLite (supports everything)
ctx := stmts.NewGenContext(db, lcg, 2)

// Turso-specific flavor
tursoFlavor := turso.NewTursoFlavorConfig()
ctx := stmts.NewGenContextWithFlavor(db, lcg, 2, tursoFlavor)
```

### Checking Feature Support

```go
if ctx.SupportsFeature("window_functions") {
    // Generate window function SQL
} else {
    // Generate alternative SQL (e.g., using ROWID)
}
```

## Benefits

1. **Single Codebase**: One implementation supports multiple databases
2. **Automatic Adaptation**: Generators automatically produce compatible SQL
3. **Extensibility**: Easy to add new database flavors
4. **Safety**: Prevents generation of unsupported SQL that would fail
5. **Backward Compatibility**: Existing code continues to work

## Learn More

- See `internal/generators/sqlite/stmts/FLAVOR_PATTERN.md` for detailed documentation
- See `internal/generators/turso/TURSO_COMPAT.md` for Turso-specific constraints
- See tests in `internal/generators/sqlite/stmts/flavor_test.go` for usage examples
