# Stmts Design Patterns

## Overview

The `internal/stmts/stmts` package has been refactored to use proper design patterns that improve code maintainability, reduce duplication, and make the codebase more extensible.

## Design Patterns Implemented

### 1. Factory Pattern

**File:** `factory.go`

The Factory pattern centralizes the creation of statement generators, eliminating code duplication and providing a consistent interface for generator instantiation.

**Key Components:**
- `StmtGeneratorFactory`: Main factory class for creating generators
- `CreateGenerator(stmtType StmtType)`: Creates appropriate generator based on statement type
- `GenerateStmt(db, stmtType)`: Convenience method that creates and executes a generator
- Custom error types: `UnsupportedStmtTypeError`, `CannotGenerateError`

**Benefits:**
- Single responsibility: Generator creation is centralized
- Extensibility: Easy to add new statement types
- Type safety: Compile-time checking for supported statement types
- Reduced boilerplate: No need to manually wire up generators

**Example Usage:**
```go
factory := NewStmtGeneratorFactory(lcg, maxDepth, flavorConfig)
stmt, err := factory.GenerateStmt(db, StmtSelectWhere)
if err != nil {
    // Handle error
}
sql := stmt.SQL()
```

### 2. Strategy Pattern

**Files:** `insert_variants.go`, `select_variants.go`

The Strategy pattern encapsulates different SQL generation strategies, allowing the algorithm to vary independently from clients that use it.

**Key Components:**
- `InsertVariantGenerator`: Handles different INSERT statement variants (OR REPLACE, OR IGNORE, etc.)
- `SelectVariantGenerator`: Handles different SELECT statement variants (WHERE, JOIN, CTE, etc.)

**Benefits:**
- Polymorphism: Different strategies can be swapped at runtime
- Cohesion: Related generation logic is grouped together
- Open/Closed Principle: Easy to add new variants without modifying existing code
- DRY: Eliminates repetitive conditional logic

**Example Usage:**
```go
// Create a generator for INSERT OR REPLACE
gen := &InsertVariantGenerator{variant: StmtInsertOrReplace}
stmt, err := gen.Generate(ctx)

// Create a generator for SELECT with CTE
gen := &SelectVariantGenerator{variant: StmtSelectCTE, maxDepth: 3}
stmt, err := gen.Generate(ctx)
```

### 3. Builder Pattern (Existing, Enhanced)

**File:** `builder.go`

The Builder pattern provides a fluent interface for constructing complex SQL statements step by step.

**Key Components:**
- `SelectBuilder`: Fluent interface for SELECT statements
- `InsertBuilder`: Fluent interface for INSERT statements
- `RandomSelectBuilder`: Builder that generates random SELECT statements

**Benefits:**
- Readability: Clear, chainable API
- Flexibility: Optional clauses can be added as needed
- Validation: Build-time validation of SQL structure
- Immutability: Builder doesn't mutate after Build()

**Example Usage:**
```go
stmt, err := NewSelectBuilder(ctx).
    FromTable(table).
    Select("col1", "col2").
    Where("col1 > 10").
    OrderBy("col2 DESC").
    Limit(100).
    Build()
```

### 4. Template Method Pattern (via Interface)

**File:** `stmt.go`

The StmtGenerator interface defines a template for all statement generators, ensuring consistency across different implementations.

**Key Components:**
- `StmtGenerator` interface: Defines `Generate()` and `CanGenerate()` methods
- `GeneratorFunc`: Function adapter for simple generators
- `StmtGeneratorWithCheck`: Adapter for generators with custom validation
- `GeneratorRegistry`: Registry pattern for dynamic generator lookup

**Benefits:**
- Consistency: All generators follow the same contract
- Testability: Easy to mock and test individual generators
- Composition: Can combine generators using the interface
- Documentation: Interface serves as living documentation

### 5. Registry Pattern (Existing, Enhanced)

**File:** `stmt.go`

The Registry pattern provides centralized management of statement generators.

**Key Components:**
- `GeneratorRegistry`: Thread-safe registry for generators
- `DefaultRegistry()`: Pre-populated with standard generators
- `Register()`, `Get()`, `Has()`, `Names()`: Registry operations

**Benefits:**
- Decoupling: Clients don't need to know about concrete generator types
- Dynamic registration: Can register custom generators at runtime
- Discovery: Can enumerate available generators
- Thread-safety: Safe for concurrent access

## Refactored Code

### gen_map.go Simplification

**Before:**
```go
// 340+ lines of repetitive lambda functions
m["insert"] = func(db *sql.DB) (string, error) {
    s, err := GenInsert(db, lcg)
    if err != nil {
        return "INSERT INTO sqlite_master DEFAULT VALUES;", err
    }
    return s.SQL(), nil
}
// ... repeated 40+ times with different statements
```

**After:**
```go
// ~70 lines using factory pattern
factory := NewStmtGeneratorFactory(lcg, maxRecursionDepth, flavorConfig)
createGenFunc := func(stmtType StmtType, fallback string) func(db *sql.DB) (string, error) {
    return func(db *sql.DB) (string, error) {
        stmt, err := factory.GenerateStmt(db, stmtType)
        if err != nil {
            return fallback, err
        }
        return stmt.SQL(), nil
    }
}

m["insert"] = createGenFunc(StmtInsert, "INSERT INTO sqlite_master DEFAULT VALUES;")
// ... simple one-line registrations
```

**Improvements:**
- 80% reduction in lines of code
- Eliminated code duplication
- Consistent error handling
- Easier to add new statement types
- Better testability

## Testing

All design patterns are thoroughly tested:

- **Factory Pattern Tests:** `factory_test.go`
  - Generator creation validation
  - Statement generation validation
  - Error handling
  - Context creation

- **Strategy Pattern Tests:** Included in `factory_test.go`
  - INSERT variant generation
  - SELECT variant generation
  - Variant-specific behavior

- **Existing Tests:** All existing tests continue to pass
  - Builder pattern tests in `builder_test.go`
  - Statement generation tests in `stmt_test.go`
  - Integration tests in `turso_compat_test.go`

## Migration Guide

### For Generator Users

No changes required! The refactoring maintains backward compatibility with existing code that uses `GenSelect()`, `GenInsert()`, etc.

### For New Code

Prefer using the factory pattern for new code:

```go
// Old way (still works)
stmt, err := GenSelect(db, lcg)

// New way (recommended)
factory := NewStmtGeneratorFactory(lcg, maxDepth, flavorConfig)
stmt, err := factory.GenerateStmt(db, StmtSelectBasic)
```

### For Adding New Statement Types

1. Add the statement type to `types.go` and `AllStmtTypes`
2. Implement the generator (or add to existing variant generator)
3. Update factory's `CreateGenerator()` method
4. Register in `gen_map.go` using `createGenFunc()`

## Future Enhancements

Potential improvements for future iterations:

1. **Command Pattern**: For SQL statement execution with undo/redo support
2. **Visitor Pattern**: For SQL AST traversal and transformation
3. **Chain of Responsibility**: For statement validation pipelines
4. **Decorator Pattern**: For statement enhancement (e.g., adding comments, hints)
5. **Composite Pattern**: For complex nested query generation

## References

- Gang of Four Design Patterns
- Effective Go: Composition and Interfaces
- Clean Code: SOLID Principles
