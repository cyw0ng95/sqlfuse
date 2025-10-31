# Statement Package Refactoring Status

## Overview
Refactoring of `internal/stmts/stmts` package into organized sub-packages (DDL, DML, Other) to improve code readability and maintainability.

## Completed Work

### 1. Package Structure Created
- ✅ `internal/stmts/ddl` - Data Definition Language statements (9 files)
- ✅ `internal/stmts/dml` - Data Manipulation Language statements (31 files)  
- ✅ `internal/stmts/other` - Utility and control statements (17 files)

### 2. Files Organized by Category

**DDL Package (9 files):**
- alter_table.go
- create_table.go
- create_view.go
- drop_table.go
- drop_view.go
- index.go
- schema.go
- trigger.go
- virtual_table.go

**DML Package (31 files):**
- delete.go, insert.go, insert_variants.go, update.go
- select.go and 26 select_*.go variants
- Full SELECT statement support with all join types, aggregates, CTEs, etc.

**OTHER Package (17 files):**
- Transaction control: transaction.go, savepoint.go
- Database utilities: pragma.go, analyze.go, vacuum.go, reindex.go
- Database management: attach.go, copy.go, explain.go
- DuckDB-specific: duckdb_advanced.go, duckdb_extensions.go
- Advanced: macro.go, pivot.go, sequence.go, set.go
- Expression generation: expr.go, expr_recursive.go

### 3. Helper Functions Exported
Updated `internal/stmts/stmts` package to export:
- `EnsureLCG()` - LCG validation with backward compatibility
- `EnsureFlavor()` - Flavor configuration validation  
- `HasTables()` - Database table existence check
- `QuoteIdent()` - SQL identifier quoting

### 4. Import Updates
- All new packages import `sqlfuse/internal/stmts/stmts` for base types
- Type references updated to use `stmts.` prefix where appropriate
- Struct literals fixed to use correct embedded field syntax

## Build Status

- **DDL Package**: ✅ **Compiles Successfully**
- **DML Package**: ⚠️ Pending (ExprGenerator dependency)
- **OTHER Package**: ⚠️ Pending

## Remaining Work

### 1. Resolve Expression Generator Dependencies
**Issue**: `NewExprGenerator` is defined in `expr_recursive.go` (now in OTHER package) but is heavily used by DML package for SELECT statement generation.

**Solutions** (choose one):
- **Option A**: Move expr*.go files to DML package (expressions primarily used for SELECT)
- **Option B**: Keep expr in OTHER and export `NewExprGenerator` for DML to import
- **Option C**: Create separate `internal/stmts/expr` package that both can import

**Recommended**: Option A - Move to DML since expressions are primarily used in SELECT statements.

### 2. Update Factory and Registry
Update `internal/stmts/stmts/factory.go` to:
- Import new sub-packages (ddl, dml, other)
- Return generators from appropriate sub-packages
- Example: `case StmtCreateTable: return &ddl.CreateTableGenerator{}`

### 3. Update Generator Map
Update `internal/stmts/stmts/gen_map.go` (BuildGeneratorFuncs) to:
- Use generators from new packages
- Delegate to sub-package generators instead of local ones

### 4. Update External Imports
Files outside stmts that import specific generators need updates:
- `internal/generators/base_generator.go`
- `internal/generators/turso.go`
- `internal/generators/go_sqlite3.go`
- `internal/generators/duckdb.go`
- `internal/generators/dialects/*.go`
- `examples/impedance_example.go`

### 5. Testing
- Run full test suite: `go test ./internal/stmts/...`
- Verify backward compatibility
- Check all executors still build and run

### 6. Cleanup
After all packages work and tests pass:
- Remove duplicate files from `internal/stmts/stmts` that were moved
- Keep only core infrastructure files in stmts/stmts
- Update documentation

## Core Files Remaining in internal/stmts/stmts

These files provide base types and infrastructure:
- `stmt.go` - Base interfaces (Stmt, StmtGenerator, etc.)
- `context.go` - GenContext type
- `factory.go` - Statement factory
- `gen_map.go` - Generator function map
- `base_stmt.go` - BaseStmt implementation
- `types.go` / `type.go` - StmtType enums
- `builder.go` - SQL builder utilities
- `metadata.go` - Statement metadata
- `generator_helpers.go` - Helper functions
- `misc.go` - Miscellaneous utilities

## Testing Strategy

1. **Unit Tests**: Verify each sub-package compiles and basic functions work
2. **Integration Tests**: Run existing test suite to ensure no regressions
3. **Build Tests**: Verify all executors (turso, go-sqlite3, duckdb) still build
4. **Functional Tests**: Run sample SQL generation to verify correctness

## Migration Path for External Code

Code importing from `sqlfuse/internal/stmts/stmts` will need updates:

**Before:**
```go
import "sqlfuse/internal/stmts/stmts"

gen := &stmts.CreateTableGenerator{}
```

**After (if using new packages directly):**
```go
import "sqlfuse/internal/stmts/ddl"
import "sqlfuse/internal/stmts/stmts"

gen := &ddl.CreateTableGenerator{}  // Generator from sub-package
ctx := stmts.NewGenContext(...)     // Context from base package
```

**Or (if using factory - recommended for compatibility):**
```go
import "sqlfuse/internal/stmts/stmts"

factory := stmts.NewStmtGeneratorFactory(...)
gen := factory.CreateGenerator(stmts.StmtCreateTable)  // Factory handles delegation
```

## Benefits of This Refactoring

1. **Improved Organization**: Files logically grouped by SQL statement category
2. **Better Readability**: Smaller, focused packages easier to navigate
3. **Clear Separation**: DDL, DML, and utilities clearly separated
4. **Maintainability**: Changes to one category isolated from others
5. **Documentation**: Package structure self-documents statement categories

## Next Steps

1. Resolve ExprGenerator dependency (move expr files to DML)
2. Update factory.go to delegate to new packages
3. Update gen_map.go generator functions
4. Fix remaining compile errors in DML and OTHER packages
5. Run test suite and fix any failures
6. Update external code imports
7. Remove duplicate files from original stmts/stmts
8. Update main README with new package structure
