# SQL Function Generation Support

This module implements SQL function generation for testing Turso/LibSQL compatibility with SQLite functions as specified in the [Turso COMPAT.md](https://github.com/tursodatabase/turso/blob/main/COMPAT.md#sql-functions).

## Supported Function Categories

### Scalar Functions

The following scalar functions are supported for SQL statement generation:

| Function | Variants | Status | Notes |
|----------|----------|--------|-------|
| abs(X) | 1 | ✅ | Absolute value |
| char(X1,X2,...) | Variable args | ✅ | Character from codepoints |
| coalesce(X,Y,...) | Variable args | ✅ | First non-NULL value |
| concat(X,...) | Variable args | ✅ | String concatenation |
| concat_ws(SEP,X,...) | Variable args | ✅ | Concatenate with separator |
| hex(X) | 1 | ✅ | Hexadecimal representation |
| ifnull(X,Y) | 1 | ✅ | NULL replacement |
| iif(X,Y,Z) | 1 | ✅ | If-then-else |
| instr(X,Y) | 1 | ✅ | String position |
| length(X) | 1 | ✅ | String length |
| like(X,Y) | 2 | ✅ | Pattern matching |
| lower(X) | 1 | ✅ | Lowercase conversion |
| upper(X) | 1 | ✅ | Uppercase conversion |
| ltrim(X) | 2 | ✅ | Left trim |
| rtrim(X) | 2 | ✅ | Right trim |
| trim(X) | 2 | ✅ | Trim whitespace |
| max(X,Y,...) | Variable args | ✅ | Maximum value |
| min(X,Y,...) | Variable args | ✅ | Minimum value |
| nullif(X,Y) | 1 | ✅ | NULL if equal |
| octet_length(X) | 1 | ✅ | Byte length |
| quote(X) | 1 | ✅ | SQL quoting |
| random() | 1 | ✅ | Random number |
| randomblob(N) | 1 | ✅ | Random blob |
| replace(X,Y,Z) | 1 | ✅ | String replacement |
| round(X) | 2 | ✅ | Rounding |
| sign(X) | 1 | ✅ | Sign of number |
| soundex(X) | 1 | ✅ | Soundex encoding |
| substr(X,Y) | 2 | ✅ | Substring |
| substring(X,Y) | 2 | ✅ | Substring (alias) |
| typeof(X) | 1 | ✅ | Type name |
| unhex(X) | 2 | ✅ | Hex decoding |
| unicode(X) | 1 | ✅ | Unicode codepoint |
| zeroblob(N) | 1 | ✅ | Zero-filled blob |

### Mathematical Functions

All mathematical functions from the Turso COMPAT.md are supported:

| Function | Status | Notes |
|----------|--------|-------|
| acos(X) | ✅ | Arc cosine |
| acosh(X) | ✅ | Hyperbolic arc cosine |
| asin(X) | ✅ | Arc sine |
| asinh(X) | ✅ | Hyperbolic arc sine |
| atan(X) | ✅ | Arc tangent |
| atan2(Y,X) | ✅ | Two-argument arc tangent |
| atanh(X) | ✅ | Hyperbolic arc tangent |
| ceil(X) | ✅ | Ceiling |
| ceiling(X) | ✅ | Ceiling (alias) |
| cos(X) | ✅ | Cosine |
| cosh(X) | ✅ | Hyperbolic cosine |
| degrees(X) | ✅ | Radians to degrees |
| exp(X) | ✅ | Exponential |
| floor(X) | ✅ | Floor |
| ln(X) | ✅ | Natural logarithm |
| log(X) | ✅ | Natural logarithm |
| log(B,X) | ✅ | Logarithm base B |
| log10(X) | ✅ | Base-10 logarithm |
| log2(X) | ✅ | Base-2 logarithm |
| mod(X,Y) | ✅ | Modulo |
| pi() | ✅ | Pi constant |
| pow(X,Y) | ✅ | Power |
| power(X,Y) | ✅ | Power (alias) |
| radians(X) | ✅ | Degrees to radians |
| sin(X) | ✅ | Sine |
| sinh(X) | ✅ | Hyperbolic sine |
| sqrt(X) | ✅ | Square root |
| tan(X) | ✅ | Tangent |
| tanh(X) | ✅ | Hyperbolic tangent |
| trunc(X) | ✅ | Truncate |

### Aggregate Functions

All aggregate functions are supported:

| Function | Variants | Status | Notes |
|----------|----------|--------|-------|
| avg(X) | 1 | ✅ | Average |
| count() | 2 | ✅ | Count rows |
| count(*) | 2 | ✅ | Count all rows |
| group_concat(X) | 2 | ✅ | Concatenate strings |
| string_agg(X,Y) | 1 | ✅ | String aggregation |
| max(X) | 1 | ✅ | Maximum |
| min(X) | 1 | ✅ | Minimum |
| sum(X) | 1 | ✅ | Sum |
| total(X) | 1 | ✅ | Total (always numeric) |

### Date and Time Functions

Core date/time functions are supported:

| Function | Status | Notes |
|----------|--------|-------|
| date() | ✅ | Date extraction |
| time() | ✅ | Time extraction |
| datetime() | ✅ | DateTime formatting |
| julianday() | ✅ | Julian day number |
| unixepoch() | ✅ | Unix timestamp |
| strftime() | ✅ | Format date/time |
| timediff() | ✅ | Time difference |

### JSON Functions

Comprehensive JSON function support:

| Function | Status | Notes |
|----------|--------|-------|
| json(json) | ✅ | Validate and minify JSON |
| jsonb(json) | ✅ | Binary JSON |
| json_array(...) | ✅ | Create JSON array |
| jsonb_array(...) | ✅ | Create binary JSON array |
| json_array_length(json) | ✅ | Array length |
| json_extract(json,path) | ✅ | Extract value |
| jsonb_extract(json,path) | ✅ | Extract from binary JSON |
| json_insert(json,path,value) | ✅ | Insert value |
| json_object(label,value,...) | ✅ | Create JSON object |
| jsonb_object(label,value,...) | ✅ | Create binary JSON object |
| json_patch(json1,json2) | ✅ | Apply JSON patch |
| json_pretty(json) | ✅ | Pretty print JSON |
| json_remove(json,path) | ✅ | Remove value |
| json_replace(json,path,value) | ✅ | Replace value |
| json_set(json,path,value) | ✅ | Set value |
| json_type(json) | ✅ | Get value type |
| json_valid(json) | ✅ | Validate JSON |
| json_quote(value) | ✅ | Quote value as JSON |

## Usage

### Basic Usage

Each function category has a dedicated generator function:

```go
import (
    "database/sql"
    "sqlsmith-go/internal/common"
    "sqlsmith-go/internal/generators/turso/stmts"
)

// Generate SELECT with scalar functions
lcg := common.NewLCG(42)
stmt, err := stmts.GenSelectWithScalarFunction(db, lcg)
sql := stmt.SQL()  // e.g., "SELECT abs(-42) FROM users LIMIT 5;"

// Generate SELECT with mathematical functions
stmt, err = stmts.GenSelectWithMathFunction(db, lcg)
sql = stmt.SQL()  // e.g., "SELECT sin(0.5) FROM products LIMIT 3;"

// Generate SELECT with aggregate functions
stmt, err = stmts.GenSelectWithAggregateFunction(db, lcg)
sql = stmt.SQL()  // e.g., "SELECT avg(age) FROM users;"

// Generate SELECT with date/time functions
stmt, err = stmts.GenSelectWithDateTimeFunction(db, lcg)
sql = stmt.SQL()  // e.g., "SELECT date('now');"

// Generate SELECT with JSON functions
stmt, err = stmts.GenSelectWithJSONFunction(db, lcg)
sql = stmt.SQL()  // e.g., "SELECT json_extract('{\"a\":1}', '$.a');"
```

### Testing

All generated SQL is:
1. **Syntax validated** using the SQLite ANTLR4 parser
2. **Execution tested** against Turso/LibSQL embedded database
3. **Deterministic** - same seed produces same SQL

Run tests with:
```bash
go test ./internal/generators/turso/stmts/... -v
```

### Function Randomization

Functions are selected randomly based on the provided LCG (Linear Congruential Generator):
- Each function has equal probability of being selected
- Arguments are randomized based on function signature
- Column references are randomly selected from available tables
- Literal values are generated with appropriate types

## Implementation Details

### Architecture

The implementation follows the existing pattern in the codebase:

```
internal/generators/turso/stmts/
├── select_functions.go       # Function generators
└── select_functions_test.go  # Comprehensive tests
```

### Helper Functions

Key helper functions used:
- `findNumericColumn()` - Locate numeric columns for math functions
- `findTextColumn()` - Locate text columns for string functions
- `isNumericType()` - Type checking for columns
- `isTextType()` - Type checking for text columns
- `joinStrings()` - String concatenation utility

### Test Coverage

Tests include:
- **Syntax validation** for all functions
- **Execution tests** with real database
- **Determinism tests** to ensure reproducibility
- **Coverage tests** to verify function variety
- **Individual function tests** for each supported function

## Turso Compatibility

This implementation targets the functions listed in [Turso COMPAT.md](https://github.com/tursodatabase/turso/blob/main/COMPAT.md#sql-functions). All functions marked as "Yes" in the compatibility document are supported.

### Excluded Functions

Functions marked as "No" or "Partial" with significant limitations in Turso COMPAT.md are not included:
- `format(FORMAT,...)` - Not supported
- `load_extension(X,Y)` - Not supported
- `sqlite_compileoption_*` - Not supported
- Some date/time modifiers with partial support

## Future Enhancements

Potential areas for expansion:
- [ ] Window functions (when Turso support improves)
- [ ] JSON table-valued functions (json_each, json_tree)
- [ ] Aggregate function filters (FILTER clause)
- [ ] Function nesting and composition
- [ ] User-defined function support
- [ ] Extended date/time modifier combinations

## References

- [Turso COMPAT.md](https://github.com/tursodatabase/turso/blob/main/COMPAT.md)
- [SQLite Core Functions](https://www.sqlite.org/lang_corefunc.html)
- [SQLite Math Functions](https://www.sqlite.org/lang_mathfunc.html)
- [SQLite Aggregate Functions](https://www.sqlite.org/lang_aggfunc.html)
- [SQLite Date/Time Functions](https://www.sqlite.org/lang_datefunc.html)
- [SQLite JSON Functions](https://www.sqlite.org/json1.html)
