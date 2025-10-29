# Keywords Package

This package provides comprehensive SQLite keyword support for the sqlsmith-go project.

## Purpose

The keywords package ensures that SQL identifiers (table names, column names, etc.) are properly quoted when they conflict with SQLite reserved keywords. This prevents SQL syntax errors when randomly generated names happen to match keywords like SELECT, FROM, WHERE, etc.

## Features

- **147 SQLite keywords**: Complete list from [sqlite.org/lang_keywords.html](https://sqlite.org/lang_keywords.html)
- **Case-insensitive detection**: Works with uppercase, lowercase, or mixed-case identifiers
- **Turso compatibility checking**: Identifies keywords not supported by Turso LibSQL
- **Smart quoting**: Two quoting strategies (always quote vs. quote when needed)
- **Zero dependencies**: Pure Go implementation with no external dependencies

## API

### Keyword Detection

```go
// Check if a string is any SQLite keyword
IsKeyword(name string) bool

// Check if a string is a reserved keyword
IsReserved(name string) bool

// Check if a keyword is supported by Turso LibSQL
IsTursoSupported(name string) bool

// Get detailed keyword information
GetKeyword(name string) *Keyword

// Get all keywords
AllKeywords() []Keyword
```

### Identifier Quoting

```go
// Always quote (safest, recommended for SQL generation)
QuoteIdentifier(name string) string

// Quote only when necessary (for optimization or readability)
QuoteIdentifierIfNeeded(name string) string
```

## Usage

### Basic Example

```go
import "sqlsmith-go/internal/stmts/keywords"

// Check if identifier is a keyword
if keywords.IsKeyword("SELECT") {
    fmt.Println("SELECT is a keyword")  // Will print
}

// Quote an identifier
tableName := "table"  // This is a keyword
quoted := keywords.QuoteIdentifier(tableName)  // Returns: "table"

// Use in SQL
sql := fmt.Sprintf("CREATE TABLE %s (id INTEGER)", quoted)
// Produces: CREATE TABLE "table" (id INTEGER)
```

### Integration with Statement Generation

The package is automatically used by the `quoteIdent` function:

```go
// In internal/stmts/stmts/insert.go
func quoteIdent(s string) string {
    return keywords.QuoteIdentifier(s)  // Now keyword-aware
}

// All statement generation automatically benefits
sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", 
    quoteIdent("select"),   // Properly quoted: "select"
    quoteIdent("from"),     // Properly quoted: "from"
    values)
```

### Turso Compatibility

```go
// Check before using advanced SQL features
if keywords.IsTursoSupported("WINDOW") {
    // Generate window function SQL
} else {
    // Use alternative approach for Turso
}
```

## Keyword Classification

### All Keywords are Reserved

For safety and simplicity, all 147 keywords are classified as `Reserved`. This conservative approach ensures:

1. No ambiguity about when to quote
2. Consistent behavior across all SQL dialects
3. Protection against future SQLite changes

### Turso Unsupported Keywords

These keywords are flagged as unsupported by Turso LibSQL:

- `FILTER` - Aggregate FILTER clause not supported
- `MATCH` - MATCH operator not supported  
- `MATERIALIZED` - MATERIALIZED CTEs not supported
- `OVER` - Window functions not supported
- `PARTITION` - Partitioning in window functions not supported
- `RAISE` - RAISE() function not supported
- `RECURSIVE` - RECURSIVE CTEs not supported
- `REGEXP` - REGEXP operator not supported
- `WINDOW` - Window function syntax not supported

## Identifier Quoting Rules

### When to Use QuoteIdentifier (Always Quote)

Use for:
- SQL statement generation (recommended)
- Maximum safety
- Unknown or dynamic identifiers

```go
quoted := keywords.QuoteIdentifier(userInput)  // Always safe
```

### When to Use QuoteIdentifierIfNeeded (Conditional Quote)

Use for:
- Known safe identifiers
- Optimization (reduce quote overhead)
- Readability in logs/output

```go
quoted := keywords.QuoteIdentifierIfNeeded("my_column")  // my_column (unquoted)
quoted := keywords.QuoteIdentifierIfNeeded("select")     // "select" (quoted)
```

### Quoting Necessity Rules

An identifier needs quoting if:

1. It's a reserved SQLite keyword (case-insensitive)
2. It starts with a digit
3. It starts with a character other than letter or underscore
4. It contains special characters (space, hyphen, dot, etc.)
5. It's an empty string

Valid unquoted identifiers:
- Start with letter (a-z, A-Z) or underscore (_)
- Contain only letters, digits, underscores, or dollar signs ($)

## Testing

Run tests:

```bash
cd internal/stmts/keywords
go test -v
```

Or from repository root:

```bash
export SQLSMITH_GO_CONTAINER_TYPE=test
go test -v ./internal/stmts/keywords/
```

Test coverage includes:
- All 147 keywords verified
- Case-insensitive detection
- Turso compatibility flags
- Quote escaping (double-quote handling)
- Edge cases (empty strings, special characters)
- Integration scenarios

## Performance

- **Keyword lookup**: O(1) via hash map
- **String operations**: Optimized for common cases
- **Memory**: Minimal overhead (~2KB for keyword data)

## Implementation Notes

### Case Insensitivity

Keywords are stored in uppercase internally. All lookups convert input to uppercase using a simple ASCII conversion:

```go
func toUpper(s string) string {
    // Efficient ASCII-only uppercase for SQL keywords
}
```

### Quote Escaping

Double quotes in identifiers are escaped per SQLite rules:

```go
QuoteIdentifier("my\"col")  // Returns: "my""col"
```

This ensures SQL syntax validity when identifiers contain quote characters.

## Examples

See `integration_test.go` for comprehensive examples including:

- Keyword-as-identifier scenarios
- Common SQL keyword conflicts  
- Edge case handling
- Turso compatibility checking
- Real-world identifier patterns

## Related Documentation

- [docs/KEYWORDS_SUPPORT.md](../../docs/KEYWORDS_SUPPORT.md) - Full feature documentation
- [SQLite Keywords](https://sqlite.org/lang_keywords.html) - Official SQLite documentation
- [Turso Compatibility](https://github.com/tursodatabase/turso/blob/main/COMPAT.md) - Turso LibSQL constraints

## Contributing

When adding new functionality:

1. Update keyword list if SQLite adds keywords
2. Add tests for new features
3. Update Turso compatibility flags if needed
4. Run full test suite before committing

## License

Part of the sqlsmith-go project.
