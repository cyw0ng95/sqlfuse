# SQL Flavor Design Pattern

## Overview

This document describes the design pattern used to support different SQL flavors (dialects) in sqlsmith-go. Different SQL databases have different feature sets and compatibility constraints. This pattern allows statement generators to adapt their behavior based on the target SQL flavor.

## Problem Statement

Different SQL flavors have different capabilities:
- **Turso LibSQL**: Does not support window functions, recursive CTEs, EXISTS/IN subqueries
- **Standard SQLite**: Supports most features including window functions and recursive CTEs  
- **PostgreSQL**: Has its own set of features and syntax variations
- **MySQL**: Different syntax for certain operations

Without a design pattern to handle these differences, generators would either:
1. Generate SQL that fails on some databases
2. Generate only the lowest common denominator SQL
3. Require separate generator implementations for each database

## Solution: FlavorConfig Interface

We use the **Strategy Pattern** to encapsulate SQL flavor-specific behavior.

### Core Components

#### 1. FlavorConfig Interface

```go
type FlavorConfig interface {
    Name() string
    SupportsFeature(feature string) bool
    ValidateSQL(sql string) error
}
```

This interface defines three key methods:
- `Name()`: Returns the flavor identifier (e.g., "turso", "sqlite", "postgres")
- `SupportsFeature(feature)`: Checks if a specific SQL feature is supported
- `ValidateSQL(sql)`: Optionally validates SQL syntax (can return nil if not implemented)

#### 2. GenContext Enhancement

The `GenContext` struct now includes a `Flavor` field:

```go
type GenContext struct {
    DB       *sql.DB
    LCG      *common.LCG
    Depth    int
    MaxDepth int
    Flavor   FlavorConfig // SQL dialect configuration
}
```

Helper method for easy feature checking:

```go
func (ctx *GenContext) SupportsFeature(feature string) bool {
    if ctx.Flavor != nil {
        return ctx.Flavor.SupportsFeature(feature)
    }
    return true
}
```

#### 3. Flavor-Specific Implementations

Each SQL flavor implements the `FlavorConfig` interface. For example, `TursoFlavorConfig`:

```go
type TursoFlavorConfig struct{}

func (t *TursoFlavorConfig) Name() string {
    return "turso"
}

func (t *TursoFlavorConfig) SupportsFeature(feature string) bool {
    unsupported := map[string]bool{
        "window_functions": false,
        "cte_recursive":    false,
        "exists_subquery":  false,
        // ... more constraints
    }
    
    if supported, exists := unsupported[feature]; exists {
        return supported
    }
    return true // Default to supported
}
```

## Usage Examples

### Creating Flavor-Aware Generators

#### Example 1: Turso Generator

```go
func NewGenerator(seed uint64) *Generator {
    g := &Generator{
        lcg:          common.NewLCG(seed),
        flavorConfig: NewTursoFlavorConfig(), // Use Turso-specific flavor
    }
    return g
}

func (g *Generator) createGenContext(db *sql.DB) *stmts.GenContext {
    return stmts.NewGenContextWithFlavor(db, g.lcg, g.maxRecursionDepth, g.flavorConfig)
}
```

#### Example 2: Flavor-Aware Statement Generation

```go
func genSelectWithWindowFunctionInternal(ctx *GenContext) (SelectStmt, error) {
    // Check if window functions are supported
    if !ctx.SupportsFeature("window_functions") {
        // Generate alternative SQL for unsupported flavors
        return SelectStmt{sql: "SELECT ROWID AS row_num FROM table;"}, nil
    }
    
    // Generate window function SQL for supported flavors
    return SelectStmt{sql: "SELECT ROW_NUMBER() OVER (ORDER BY id) FROM table;"}, nil
}
```

### Standardized Feature Names

To ensure consistency, we use standardized feature names:

| Feature Name | Description | Example Unsupported In |
|--------------|-------------|------------------------|
| `window_functions` | OVER clause and window functions | Turso LibSQL |
| `cte_recursive` | RECURSIVE keyword in WITH clause | Turso LibSQL |
| `cte_materialized` | MATERIALIZED keyword in WITH | Turso LibSQL |
| `exists_subquery` | EXISTS (subquery) expressions | Turso LibSQL |
| `in_subquery` | IN (subquery) expressions | Turso LibSQL |
| `filter_clause` | Aggregate FILTER (WHERE ...) | Turso LibSQL |
| `schema_qualified` | schema.table.column syntax | Turso LibSQL |
| `regexp` | REGEXP operator | Turso LibSQL |
| `match` | MATCH operator | Turso LibSQL |

### Backward Compatibility

The pattern maintains full backward compatibility with existing code:

```go
// Old way - still works, uses default SQLite flavor
stmt, err := stmts.GenSelectWithRecursiveCTE(db, lcg)

// New way - flavor-aware
ctx := stmts.NewGenContextWithFlavor(db, lcg, 2, tursoFlavor)
stmt, err := genSelectWithRecursiveCTEInternal(ctx)
```

## Implementation Guidelines

### For Generator Authors

1. **Check Feature Support**: Always check `ctx.SupportsFeature(feature)` before generating flavor-specific SQL
2. **Provide Alternatives**: When a feature isn't supported, provide a functionally similar alternative
3. **Document Features**: Document which features your generator uses
4. **Test Both Paths**: Test both supported and unsupported code paths

Example:

```go
func genComplexQuery(ctx *GenContext) (Stmt, error) {
    if ctx.SupportsFeature("window_functions") {
        // Generate with window functions
        return genWithWindowFunctions(ctx)
    }
    // Generate alternative using subqueries
    return genWithSubqueries(ctx)
}
```

### For Flavor Implementers

1. **Be Explicit**: Clearly list all unsupported features
2. **Document Constraints**: Reference official compatibility documentation
3. **Default to Permissive**: Unknown features should default to supported
4. **Keep Updated**: Update as the database evolves

Example implementation pattern:

```go
type MyDBFlavorConfig struct{}

func (m *MyDBFlavorConfig) SupportsFeature(feature string) bool {
    // Explicit unsupported features
    unsupported := map[string]bool{
        "feature1": false,
        "feature2": false,
    }
    
    if supported, exists := unsupported[feature]; exists {
        return supported
    }
    
    // Default: assume supported
    return true
}
```

## Testing

### Testing Flavor Behavior

```go
func TestFlavorSpecificBehavior(t *testing.T) {
    // Test with restrictive flavor
    restrictiveFlavor := &MockFlavorConfig{
        supportedFeatures: map[string]bool{
            "window_functions": false,
        },
    }
    ctx := NewGenContextWithFlavor(nil, lcg, 2, restrictiveFlavor)
    
    stmt, err := genSelectWithWindowFunctionInternal(ctx)
    if err != nil {
        t.Fatal(err)
    }
    
    // Verify no OVER clause in generated SQL
    if strings.Contains(stmt.SQL(), "OVER (") {
        t.Error("Should not generate OVER clause for restrictive flavor")
    }
}
```

### Mock Flavor for Testing

```go
type MockFlavorConfig struct {
    name              string
    supportedFeatures map[string]bool
}

func (m *MockFlavorConfig) Name() string {
    return m.name
}

func (m *MockFlavorConfig) SupportsFeature(feature string) bool {
    if m.supportedFeatures == nil {
        return true
    }
    supported, exists := m.supportedFeatures[feature]
    if !exists {
        return true
    }
    return supported
}
```

## Benefits

1. **Maintainability**: Single codebase for all SQL flavors
2. **Correctness**: Generators automatically adapt to flavor constraints
3. **Extensibility**: Easy to add new flavors or features
4. **Testing**: Can test flavor-specific behavior in isolation
5. **Documentation**: Self-documenting through feature names

## Future Extensions

### Potential Enhancements

1. **Syntax Variations**: Add methods for flavor-specific syntax (e.g., LIMIT vs TOP)
   ```go
   LimitSyntax(n int) string
   ```

2. **Type Mapping**: Map generic types to flavor-specific types
   ```go
   MapType(genericType string) string
   ```

3. **Function Aliasing**: Map function names across flavors
   ```go
   GetFunctionName(canonical string) string
   ```

4. **Version Support**: Check feature support by version
   ```go
   SupportsFeatureInVersion(feature string, version string) bool
   ```

## References

- Turso LibSQL Compatibility: `internal/generators/turso/TURSO_COMPAT.md`
- Statement Generation: `internal/generators/sqlite/stmts/README.md`
- GenContext Documentation: `internal/generators/sqlite/stmts/context.go`
