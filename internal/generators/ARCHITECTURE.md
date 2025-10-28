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
   - Add database-specific configuration (e.g., flavor config)
   - Implement `Name()` and `SupportedStmts()` methods
   - Define default weights for their SQL flavor
   - Placed directly in the `internal/generators/` directory (not in subdirectories)
   - Each generator is self-contained in a single file including its flavor configuration

## Benefits

1. **Code Reuse**: Common generator logic is centralized in `BaseGenerator`
2. **Maintainability**: Changes to common logic only need to be made in one place
3. **Extensibility**: New generators can be added by embedding `BaseGenerator`
4. **Flexibility**: Each generator can override or extend base behavior
5. **Type Safety**: Go's embedding provides compile-time type checking

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

1. Create a new file in `internal/generators/` (e.g., `mydb.go`)
2. Include both the generator implementation and flavor configuration in the same file
3. Define a generator struct that embeds `*BaseGenerator`
4. Implement `DefaultXXXStmtWeights()` for your database flavor
5. Implement `Name()` and `SupportedStmts()` methods
6. Implement any database-specific initialization in `NewXXXGenerator()`
7. Add flavor-specific configuration if needed

Example:
```go
package generators

// Flavor configuration
type MyDBFlavorConfig struct{}

func (m *MyDBFlavorConfig) Name() string {
    return "mydb"
}

func (m *MyDBFlavorConfig) SupportsFeature(feature string) bool {
    // Implementation
}

func NewMyDBFlavorConfig() stmts.FlavorConfig {
    return &MyDBFlavorConfig{}
}

// Generator implementation
type MyDBGenerator struct {
    *BaseGenerator
    flavorConfig stmts.FlavorConfig
}

func NewMyDBGenerator(seed uint64) *MyDBGenerator {
    g := &MyDBGenerator{
        BaseGenerator: NewBaseGenerator(seed),
        flavorConfig:  NewMyDBFlavorConfig(),
    }
    g.SetWeights(DefaultMyDBStmtWeights())
    g.initGenMap()
    return g
}

func (g *MyDBGenerator) Name() string {
    return "mydb"
}

func (g *MyDBGenerator) SupportedStmts() map[string]uint64 {
    // Return supported statements
}
```
