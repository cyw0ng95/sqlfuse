# Design Pattern Improvements - Internal Package Simplification

## Overview

This document describes the design pattern improvements applied to the `internal/stmts/stmts` package to simplify implementation, reduce code duplication, and improve maintainability.

## Patterns Applied

### 1. Template Method Pattern - BaseStmt

**Problem**: Every statement type had to implement the same three methods:
```go
func (s *XyzStmt) SQL() string          { return s.sql }
func (s *XyzStmt) Type() string         { return "xyz" }
func (s *XyzStmt) Flavor() FlavorConfig { return s.flavor }
```

This resulted in ~10 lines of repetitive boilerplate per statement type.

**Solution**: Created `BaseStmt` struct that provides a reusable implementation:

```go
type BaseStmt struct {
	sql    string
	typ    string
	flavor FlavorConfig
}

func NewBaseStmt(sql, typ string, flavor FlavorConfig) *BaseStmt {
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}
	return &BaseStmt{sql: sql, typ: typ, flavor: flavor}
}

func (s *BaseStmt) SQL() string          { return s.sql }
func (s *BaseStmt) Type() string         { return s.typ }
func (s *BaseStmt) Flavor() FlavorConfig { return s.flavor }
```

**Usage**: Statement types now embed `BaseStmt`:

```go
// Before
type DeleteStmt struct {
	sql    string
	flavor FlavorConfig
}
func (s *DeleteStmt) SQL() string          { return s.sql }
func (s *DeleteStmt) Type() string         { return "delete" }
func (s *DeleteStmt) Flavor() FlavorConfig { return s.flavor }

// After
type DeleteStmt struct {
	*BaseStmt
}

// Creation
return &DeleteStmt{
	BaseStmt: NewBaseStmt(sql, "delete", flavor),
}
```

**Benefits**:
- Eliminates 7-10 lines of boilerplate per statement type
- Ensures consistent implementation across all statement types
- Easier to add new fields to all statements if needed
- Single source of truth for statement interface implementation

**Files Affected**: 9 statement types refactored
- `delete.go` - DeleteStmt
- `drop_table.go` - DropTableStmt
- `alter_table.go` - AlterTableStmt
- `create_table.go` - CreateTableStmt
- `create_view.go` - CreateViewStmt
- `drop_view.go` - DropViewStmt
- `attach.go` - AttachStmt, DetachStmt
- `explain.go` - ExplainStmt
- `transaction.go` - TransactionStmt

### 2. Null Object Pattern - Configuration Helpers

**Problem**: Every generator function started with nil checks:
```go
func genSomething(lcg *common.LCG, flavor FlavorConfig) {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}
	// ... actual logic
}
```

This resulted in ~6 lines of repetitive nil checking per function.

**Solution**: Created helper functions that encapsulate the nil checking:

```go
// ensureLCG returns the provided LCG or creates a default one if nil.
func ensureLCG(lcg *common.LCG) *common.LCG {
	if lcg == nil {
		return common.NewLCG(1)
	}
	return lcg
}

// ensureFlavor returns the provided flavor or creates a default one if nil.
func ensureFlavor(flavor FlavorConfig) FlavorConfig {
	if flavor == nil {
		return GetDefaultFlavor()
	}
	return flavor
}

// GeneratorConfig holds common configuration for statement generators.
type GeneratorConfig struct {
	LCG    *common.LCG
	Flavor FlavorConfig
}

func NewGeneratorConfig(lcg *common.LCG, flavor FlavorConfig) *GeneratorConfig {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}
	return &GeneratorConfig{LCG: lcg, Flavor: flavor}
}
```

**Usage**:

```go
// Before
func genSomething(lcg *common.LCG, flavor FlavorConfig) {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}
	// actual logic
}

// After
func genSomething(lcg *common.LCG, flavor FlavorConfig) {
	lcg = ensureLCG(lcg)
	flavor = ensureFlavor(flavor)
	// actual logic
}
```

**Benefits**:
- Reduces boilerplate from 6 lines to 2 lines per function
- Centralizes default value logic
- Makes code more readable
- Easier to change default initialization behavior
- Provides optional `GeneratorConfig` for grouping related parameters

**Files Affected**: Applied to 12+ generator functions across:
- `delete.go`
- `drop_table.go`
- `alter_table.go`
- `create_table.go`
- `create_view.go`
- `drop_view.go`
- `attach.go`
- `explain.go`
- `transaction.go`

## Code Metrics

### Lines of Code Reduction

| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| Boilerplate per statement type | ~10 lines | ~0 lines | 100% reduction |
| Nil checks per function | ~6 lines | ~2 lines | 67% reduction |
| Total lines reduced | - | ~150+ lines | Across refactored files |

### Maintainability Improvements

1. **Consistency**: All statement types now use the same pattern for implementing the `Stmt` interface
2. **DRY Principle**: No duplication of method implementations or nil checks
3. **Single Responsibility**: Each statement type focuses only on its SQL generation logic
4. **Easier Evolution**: Adding new fields to `BaseStmt` automatically propagates to all statement types

## Backward Compatibility

✅ **100% backward compatible**

All existing APIs continue to work:
- Legacy `Gen*` functions still work exactly as before
- All statement types still implement the `Stmt` interface
- No breaking changes to function signatures
- All tests pass without modification (except test data initialization)

## New Files

1. **`base_stmt.go`** (40 lines)
   - Template Method implementation for statement interface
   - `BaseStmt` struct with common fields
   - `NewBaseStmt` constructor with automatic defaults

2. **`generator_helpers.go`** (48 lines)
   - Null Object pattern helpers
   - `ensureLCG()` and `ensureFlavor()` functions
   - `GeneratorConfig` struct for grouped configuration

## Testing

All existing tests continue to pass:
- ✅ 70+ tests in stmts package
- ✅ Integration tests in other packages
- ✅ Example tests demonstrating usage
- ✅ No test failures or regressions

Updated test data initialization to use new `BaseStmt`:
- Test structs now use `NewBaseStmt` for initialization
- Validates that embedded struct pattern works correctly

## Future Enhancements

These improvements lay the groundwork for additional patterns:

1. **Builder Pattern Enhancement**: Could extend `BaseStmt` with builder methods for common modifications
2. **Decorator Pattern**: Could wrap `BaseStmt` to add features like logging, validation, or transformation
3. **Prototype Pattern**: Could add `Clone()` methods to `BaseStmt` for statement copying
4. **Composite Pattern**: Could use `BaseStmt` as base for complex multi-statement constructs

## Migration Guide

### For Statement Type Authors

When creating a new statement type:

```go
// 1. Define the struct with embedded BaseStmt
type MyNewStmt struct {
	*BaseStmt
}

// 2. Create a generator function
func genMyNewStmt(lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	flavor = ensureFlavor(flavor)
	
	// Generate SQL
	sql := fmt.Sprintf("MY SQL STATEMENT")
	
	// Return with BaseStmt
	return &MyNewStmt{
		BaseStmt: NewBaseStmt(sql, "my_new", flavor),
	}, nil
}
```

No need to implement `SQL()`, `Type()`, or `Flavor()` methods - they're inherited from `BaseStmt`.

### For Generator Function Authors

When writing generator functions:

```go
// Replace this:
func genSomething(lcg *common.LCG, flavor FlavorConfig) {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}
	// ...
}

// With this:
func genSomething(lcg *common.LCG, flavor FlavorConfig) {
	lcg = ensureLCG(lcg)
	flavor = ensureFlavor(flavor)
	// ...
}
```

## Conclusion

These design pattern improvements significantly simplify the internal package implementation:

- **Reduced boilerplate**: ~150+ lines of repetitive code eliminated
- **Improved maintainability**: Single source of truth for common functionality
- **Better consistency**: All statement types follow the same pattern
- **Enhanced readability**: Less clutter, clearer intent
- **100% backward compatible**: No breaking changes

The improvements align with SOLID principles and make the codebase easier to understand, modify, and extend.
