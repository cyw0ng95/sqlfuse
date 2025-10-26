# Expression Generator Design and Coverage

This document describes the expression generation logic added to cover [Turso's expression compatibility](https://github.com/tursodatabase/turso/blob/main/COMPAT.md#expressions).

## Overview

The expression generator module (`expr.go`) provides centralized expression generation capabilities for various SQL expression types supported by Turso/LibSQL. This design allows expressions to be reused across multiple statement types (SELECT, WHERE, ORDER BY, etc.), following the optimization goal mentioned in the problem statement.

## Supported Expressions

### ✅ Implemented Expressions

The following expressions from the Turso COMPAT.md are implemented:

1. **CAST (expr AS type)** - `GenCastExpr()`
   - Casts column values to INTEGER, TEXT, REAL, or BLOB types
   - Example: `CAST("value" AS INTEGER)`

2. **(NOT) BETWEEN ... AND ...** - `GenBetweenExpr()`
   - Generates BETWEEN expressions with numeric columns
   - Supports both BETWEEN and NOT BETWEEN variants
   - Example: `"count" BETWEEN 10 AND 100`

3. **(NOT) GLOB** - `GenGlobExpr()`
   - Pattern matching with GLOB operator
   - Uses text columns when available
   - Supports wildcards: `*`, `?`, `[a-z]`
   - Example: `"name" GLOB 'A*'`

4. **COLLATE** - `GenCollateExpr()`
   - Uses default SQLite collations (BINARY, NOCASE, RTRIM)
   - Custom collations are not supported by Turso
   - Example: `"name" COLLATE NOCASE`

5. **Unary operators** - `GenUnaryExpr()`
   - Supports: `+`, `-`, `~`, `NOT`
   - Example: `-"value"`, `NOT ("count" = 10)`

6. **Binary operators** - `GenBinaryExpr()`
   - Arithmetic: `+`, `-`, `*`, `/`, `&`, `|`, `<<`, `>>`
   - Comparison: `=`, `!=`, `<`, `<=`, `>`, `>=`, `<>`
   - Logical: `AND`, `OR`
   - **Excludes unsupported operators**: `%`, `!<`, `!>` (as per Turso COMPAT.md)
   - Example: `"count" + "value"`, `"id" = 100`

7. **Parenthesized expressions** - `GenParenthesizedExpr()`
   - Wraps expressions in parentheses
   - Example: `("count" = 10)`

8. **IS (NOT) NULL** - `GenIsNullExpr()`
   - Tests for NULL values
   - Example: `"value" IS NOT NULL`

9. **CASE WHEN THEN ELSE END** - `genCaseExpr()` (existing)
   - Already implemented in `expr_recursive.go`
   - Example: `CASE WHEN "id" = 1 THEN 'one' ELSE 'other' END`

10. **(NOT) LIKE** - Already covered in existing code
    - Example: `"name" LIKE '%test%'`

### ❌ Not Implemented (Unsupported by Turso)

The following expressions are marked as "No" or have limitations in Turso COMPAT.md:

1. **IS (NOT) DISTINCT FROM** - Listed as "Yes" in COMPAT.md but not recognized by SQLite ANTLR parser
   - Implementation exists (`GenIsDistinctFromExpr()`) but is not used
   - Tests are skipped

2. **(NOT) IN (subquery)** - Status: No in COMPAT.md
3. **(NOT) EXISTS (subquery)** - Status: No in COMPAT.md
4. **(NOT) REGEXP** - Status: No in COMPAT.md
5. **(NOT) MATCH** - Status: No in COMPAT.md
6. **RAISE** - Status: No in COMPAT.md
7. **agg() FILTER (WHERE ...)** - Status: No (incorrectly ignored)
8. **... OVER (...)** - Status: No (incorrectly ignored)

## Design Decisions

### Centralized Expression Generator

The `ExprGenerator` struct provides a centralized location for all expression generation logic:

```go
type ExprGenerator struct {
    ctx *GenContext
}
```

This design:
- ✅ Avoids code duplication across statement generators
- ✅ Ensures consistent expression generation
- ✅ Makes it easy to add new expression types
- ✅ Allows expressions to be used in SELECT lists, WHERE clauses, ORDER BY, etc.

### Context-Based Generation

The `GenContext` provides:
- Database connection for introspection
- Random number generator (LCG) for deterministic generation
- Recursion depth tracking to prevent infinite recursion

### Random Expression Generation

`GenRandomExpr()` selects a random expression type, providing variety in generated SQL statements. This is useful for fuzzing and testing.

## Statement Generators Using Expressions

New statement generators were created to exercise the expression generators:

1. **GenSelectWithExpressions** - Various expressions in SELECT list
2. **GenSelectWhereCast** - CAST in WHERE clause
3. **GenSelectWhereBetween** - BETWEEN in WHERE clause
4. **GenSelectWhereGlob** - GLOB in WHERE clause
5. **GenSelectWithCollate** - COLLATE in ORDER BY
6. **GenSelectWithUnaryOp** - Unary operators in SELECT list
7. **GenSelectWithBinaryOp** - Binary operators in SELECT list

## Integration with Existing Code

The expression generators integrate seamlessly with existing code:

- **ExprGenerator** reuses the existing `GenContext` from `context.go`
- **Helper functions** like `isNumericType()` and `quoteIdent()` are reused
- **Type system** from `types` package is used for value generation
- **ANTLR parser** is used for SQL validation in tests

## Testing

Comprehensive unit tests were added:

- **expr_test.go** - Tests for each expression generator
  - Validates SQL syntax using ANTLR parser
  - Executes generated SQL against in-memory database
  - Tests determinism (same seed → same output)

- **select_expressions_test.go** - Tests for statement generators using expressions
  - Validates syntax and execution
  - Covers all new statement generator functions

All tests pass successfully, ensuring the generated SQL is valid and executable.

## Future Enhancements

Potential improvements:

1. Add support for IS DISTINCT FROM if Turso/LibSQL adds parser support
2. Generate more complex nested expressions
3. Add expressions to UPDATE SET clauses
4. Add expressions to INSERT VALUES
5. Generate computed columns in CREATE TABLE statements

## References

- [Turso Compatibility Documentation](https://github.com/tursodatabase/turso/blob/main/COMPAT.md#expressions)
- [SQLite Expression Syntax](https://www.sqlite.org/lang_expr.html)
