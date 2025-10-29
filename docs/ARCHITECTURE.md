# Generator Architecture

## Overview

The generator architecture has been refactored to use a clean separation of concerns with a base implementation that provides common functionality and specific implementations for database flavors.

## Design Pattern

The architecture uses the **Composition Pattern** (specifically, embedding in Go) to allow generator implementations to inherit common functionality while maintaining flexibility.

### Key Components

1. **Generator Interface** (`generators.go`)
   - Primary interface for all generators
   - Combines SQL generation with metadata queries
   - Methods:
     - `GenerateWithDB(db *sql.DB) string` - Generates SQL statements
     - `TokensUsed() uint64` - Returns PRNG token usage
     - `Name() string` - Returns generator identifier
     - `SupportedStmts() map[string]uint64` - Returns supported statement types and weights

2. **BaseGenerator** (`base_generator.go`)
   - Provides common functionality for all generators
   - Handles:
     - LCG (Linear Congruential Generator) for randomness
     - Statement weights and selection
     - Recursion depth management
     - Direction selection (choosing which SQL statement type to generate)
   - Uses the **Template Method Pattern** for SQL generation

3. **Specific Generators** (e.g., `turso.go`)
   - Embed `BaseGenerator` for common functionality
   - Reference database-specific dialect configuration from `dialects/` directory
   - Implement `Name()` and `SupportedStmts()` methods
   - Define default weights for their SQL flavor
   - Placed directly in the `internal/generators/` directory (not in subdirectories)
   - Generator files contain only generator-specific code (weights, initialization)

4. **Dialect Configurations** (`dialects/` directory)
   - Database-specific `FlavorConfig` implementations
   - Each dialect in its own file (e.g., `turso.go`)
   - Defines feature support for specific database flavors
   - Kept separate from generator logic for better organization

## Benefits

1. **Code Reuse**: Common generator logic is centralized in `BaseGenerator`
2. **Maintainability**: Changes to common logic only need to be made in one place
3. **Extensibility**: New generators can be added by embedding `BaseGenerator`
4. **Flexibility**: Each generator can override or extend base behavior
5. **Type Safety**: Go's embedding provides compile-time type checking
6. **Separation of Concerns**: Dialect configurations are separated from generator implementations

## Integration of GeneratorInfo

Previously, generator metadata (name, supported statements) was separated into a `GeneratorInfo` interface and implementation. This has been integrated directly into the `Generator` interface, simplifying the architecture:

- **Before**: Separate `GeneratorInfo` interface and `tursoInfo` implementation
- **After**: `Generator` interface includes `Name()` and `SupportedStmts()` methods

This integration:
- Reduces code duplication
- Simplifies the API surface
- Makes generator capabilities self-documenting
- Eliminates the need for parallel type hierarchies

## Usage Example

```go
// Create a new generator
gen := generators.NewTursoGenerator(seed)

// Use interface methods
name := gen.Name()                    // "turso"
stmts := gen.SupportedStmts()        // map of statement types to weights
sql := gen.GenerateWithDB(db)        // Generate SQL
tokens := gen.TokensUsed()           // Get token usage

// Configure generator
gen.SetMaxRecursionDepth(3)
gen.SetWeight(stmts.StmtSelect, 500)
```

## Adding a New Generator

To add a new database-specific generator:

1. Create a new dialect configuration file in `internal/generators/dialects/` (e.g., `mydb.go`)
   - Define a `MyDBFlavorConfig` struct implementing `stmts.FlavorConfig`
   - Implement `Name()`, `SupportsFeature()`, and `ValidateSQL()` methods
   - Provide a `NewMyDBFlavorConfig()` constructor

2. Create a new generator file in `internal/generators/` (e.g., `mydb.go`)
   - Define a generator struct that embeds `*BaseGenerator`
   - Implement `DefaultMyDBStmtWeights()` for your database flavor
   - Implement `Name()` and `SupportedStmts()` methods
   - Implement `NewMyDBGenerator()` constructor

Example dialect configuration (`dialects/mydb.go`):
```go
package dialects

import "sqlsmith-go/internal/stmts/stmts"

type MyDBFlavorConfig struct{}

func (m *MyDBFlavorConfig) Name() string {
    return "mydb"
}

func (m *MyDBFlavorConfig) SupportsFeature(feature string) bool {
    // Implementation
    return true
}

func (m *MyDBFlavorConfig) ValidateSQL(sql string) error {
    return nil
}

func NewMyDBFlavorConfig() stmts.FlavorConfig {
    return &MyDBFlavorConfig{}
}
```

Example generator (`mydb.go`):
```go
package generators

import (
    "database/sql"
    "sqlsmith-go/internal/generators/dialects"
    "sqlsmith-go/internal/stmts/stmts"
)

type MyDBGenerator struct {
    *BaseGenerator
    flavorConfig stmts.FlavorConfig
}

func DefaultMyDBStmtWeights() map[stmts.StmtType]uint64 {
    // Define weights for your database flavor
    w := map[stmts.StmtType]uint64{}
    w[stmts.StmtSelectBasic] = 100
    // ... more weights
    return w
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

func (g *MyDBGenerator) initGenMap() {
    built := stmts.BuildGeneratorFuncs(g.GetLCG(), g.GetMaxRecursionDepth(), g.flavorConfig)
    m := make(map[stmts.StmtType]func(db *sql.DB) (string, error), len(built))
    for k, fn := range built {
        m[stmts.StmtType(k)] = fn
    }
    g.SetGenMap(m)
}

func (g *MyDBGenerator) Name() string {
    return "mydb"
}

func (g *MyDBGenerator) SupportedStmts() map[string]uint64 {
    w := DefaultMyDBStmtWeights()
    out := make(map[string]uint64, len(w))
    for k, v := range w {
        out[string(k)] = v
    }
    return out
}
```
