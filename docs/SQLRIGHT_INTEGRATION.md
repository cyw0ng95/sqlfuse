# SQLRight Integration

This document describes the integration of ideas from [SQLRight](https://github.com/PSU-Security-Universe/sqlright) into sqlfuse for detecting logical bugs in database systems.

## Overview

SQLRight is a coverage-guided SQL fuzzer that combines coverage-based guidance, validity-oriented mutations, and **test oracles** to detect logical bugs in database management systems. While sqlfuse already implements ideas from the original SQLsmith (impedance matching, statistics tracking), the key innovation from SQLRight is the **oracle-based validation** approach.

## Key Concepts from SQLRight

### 1. Test Oracles

A test oracle generates **functionally equivalent** queries and compares their results to detect logical bugs. Unlike crash bugs or syntax errors, logical bugs produce incorrect results without raising errors.

**Example**: A database bug might cause `SELECT COUNT(*)` to return a different count than the actual number of rows returned by `SELECT *`.

### 2. Oracle Types

SQLRight implements several oracle types:

- **NOREC (No Empty Result Check)**: Verifies that `COUNT(*)` matches the actual row count
- **TLP (Ternary Logic Partitioning)**: Partitions results based on predicate truth values
- **LIKELY/UNLIKELY**: Tests SQLite optimization hints
- **INDEX**: Tests index correctness
- **ROWID**: Tests ROWID consistency

## Implementation in sqlfuse

### Oracle Framework

The oracle framework is implemented in `internal/oracles/` with a clean interface:

```go
// Oracle defines the interface for SQL test oracles
type Oracle interface {
    Name() string
    TransformQuery(baseQuery string) []string
    CompareResults(results []QueryResult) ComparisonResult
    IsApplicable(query string) bool
}
```

### Implemented Oracles

#### NOREC (No Empty Result Check)

**Purpose**: Detect bugs in COUNT(*) aggregation.

**Approach**:
1. Execute base query: `SELECT * FROM t WHERE x > 5`
2. Execute count query: `SELECT COUNT(*) FROM (SELECT * FROM t WHERE x > 5)`
3. Compare: Does COUNT(*) equal the actual number of rows?

**Example Bug Detected**: COUNT returns 10 but only 9 rows are actually returned.

**Usage**:
```go
oracle := oracles.NewNoRecOracle()
result := oracles.ValidateWithOracle(db, oracle, "SELECT * FROM users WHERE age > 25")
if result == oracles.Fail {
    fmt.Println("Logical bug detected!")
}
```

#### TLP (Ternary Logic Partitioning)

**Purpose**: Detect bugs in WHERE clause evaluation using SQL's ternary logic (TRUE, FALSE, NULL).

**Approach**:
1. Execute base query: `SELECT * FROM t WHERE condition`
2. Partition into three queries:
   - `WHERE condition IS TRUE`
   - `WHERE condition IS FALSE`
   - `WHERE condition IS NULL`
3. Verify: `UNION ALL` of partitions equals original result

**Example Bug Detected**: NULL handling in predicates is incorrect.

**Usage**:
```go
oracle := oracles.NewTLPOracle()
result := oracles.ValidateWithOracle(db, oracle, "SELECT * FROM users WHERE age > NULL")
if result == oracles.Fail {
    fmt.Println("Logical bug in NULL handling!")
}
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Oracle Framework                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Oracle Interface ──▶ NOREC, TLP, Future Oracles           │
│                                                              │
│  Query Execution ──▶ Result Comparison ──▶ Bug Detection   │
│                                                              │
│  ValidateWithOracle(db, oracle, query) ──▶ Pass/Fail/Error │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Comparison with SQLRight

| Feature | SQLRight | sqlfuse |
|---------|----------|---------|
| Language | C++ | Go |
| Coverage Guidance | AFL-based | Not yet implemented |
| Mutation Strategy | IR-based | LCG-based generation |
| NOREC Oracle | ✅ | ✅ |
| TLP Oracle | ✅ | ✅ (simplified) |
| LIKELY Oracle | ✅ | ❌ (future) |
| INDEX Oracle | ✅ | ❌ (future) |
| ROWID Oracle | ✅ | ❌ (future) |

## Usage Examples

### Example 1: Basic Oracle Validation

```go
package main

import (
    "database/sql"
    "fmt"
    "sqlfuse/internal/oracles"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    db, _ := sql.Open("sqlite3", ":memory:")
    defer db.Close()
    
    // Setup test data
    db.Exec("CREATE TABLE users (id INT, name TEXT, age INT)")
    db.Exec("INSERT INTO users VALUES (1, 'Alice', 30), (2, 'Bob', 25)")
    
    // Validate with NOREC oracle
    oracle := oracles.NewNoRecOracle()
    query := "SELECT name FROM users WHERE age > 20"
    
    result := oracles.ValidateWithOracle(db, oracle, query)
    
    switch result {
    case oracles.Pass:
        fmt.Println("Query passed validation")
    case oracles.Fail:
        fmt.Println("Logical bug detected!")
    case oracles.Error:
        fmt.Println("Execution error")
    }
}
```

### Example 2: Multiple Oracles

```go
func testWithMultipleOracles(db *sql.DB, query string) {
    oracles := []oracles.Oracle{
        oracles.NewNoRecOracle(),
        oracles.NewTLPOracle(),
    }
    
    for _, oracle := range oracles {
        if oracle.IsApplicable(query) {
            result := oracles.ValidateWithOracle(db, oracle, query)
            fmt.Printf("%s oracle: %v\n", oracle.Name(), result)
        }
    }
}
```

### Example 3: Integration with Generator

```go
func fuzzWithOracles(db *sql.DB, gen *generators.GoSQLite3Generator) {
    norecOracle := oracles.NewNoRecOracle()
    
    for i := 0; i < 1000; i++ {
        query := gen.GenerateWithDB(db)
        
        // Execute query normally
        _, err := db.Exec(query)
        if err != nil {
            continue
        }
        
        // Validate with oracle if applicable
        if norecOracle.IsApplicable(query) {
            result := oracles.ValidateWithOracle(db, norecOracle, query)
            if result == oracles.Fail {
                fmt.Printf("Bug found in query: %s\n", query)
            }
        }
    }
}
```

## Benefits

1. **Logical Bug Detection**: Find bugs that don't cause crashes or errors
2. **Automated Testing**: No manual test case writing required
3. **Database Agnostic**: Works with any SQL database
4. **Compositional**: Multiple oracles can be combined
5. **Extensible**: Easy to add new oracle types

## Limitations

1. **Performance Overhead**: Each oracle execution requires multiple queries
2. **False Positives**: Simplified TLP may report failures on complex queries
3. **Query Parsing**: Simple regex-based parsing may miss edge cases
4. **Coverage Guidance**: Not yet implemented (future work)

## Future Enhancements

### Phase 1 (Completed)
- ✅ Oracle framework interface
- ✅ NOREC oracle implementation
- ✅ TLP oracle implementation
- ✅ Result comparison logic
- ✅ Comprehensive test coverage

### Phase 2 (Planned)
- ⚠️ LIKELY/UNLIKELY oracle for SQLite
- ⚠️ INDEX oracle for index correctness
- ⚠️ ROWID oracle for ROWID consistency
- ⚠️ Integration with executors
- ⚠️ Improved query parsing (ANTLR-based)

### Phase 3 (Future)
- ❌ Coverage-guided fuzzing (AFL/libFuzzer integration)
- ❌ IR-based mutation strategy
- ❌ Query minimization for bug reports
- ❌ Differential testing across databases

## Testing

The oracle implementations include comprehensive tests:

```bash
# Run all oracle tests
go test ./internal/oracles/... -v

# Test specific oracle
go test ./internal/oracles/... -run TestNoRecOracle -v

# Test with coverage
go test ./internal/oracles/... -cover
```

**Test Coverage**:
- `oracle.go`: Framework and utilities
- `norec.go`: NOREC oracle (100% coverage)
- `tlp.go`: TLP oracle (100% coverage)
- `oracle_test.go`: All test cases

## Performance

### Overhead Analysis

- **NOREC Oracle**: 2x queries (original + COUNT)
- **TLP Oracle**: 2x queries (original + UNION of partitions)
- **Total Overhead**: ~100% query execution time

### Optimization Strategies

1. **Selective Application**: Only validate SELECT queries
2. **Sampling**: Validate 1 in N queries
3. **Parallel Execution**: Run oracle queries concurrently
4. **Caching**: Cache transformation results

### Recommended Usage

```go
// Validate every 10th query
if queryCount % 10 == 0 && oracle.IsApplicable(query) {
    ValidateWithOracle(db, oracle, query)
}
```

## References

- **SQLRight Paper**: [Detecting Logical Bugs of DBMS with Coverage-based Guidance](https://huhong789.github.io/papers/liang:sqlright.pdf)
- **SQLRight GitHub**: https://github.com/PSU-Security-Universe/sqlright
- **SQLancer**: https://github.com/sqlancer/sqlancer
- **SQLsmith**: https://github.com/anse1/sqlsmith
- **Original sqlfuse SQLsmith Comparison**: [docs/SQLSMITH_COMPARISON.md](SQLSMITH_COMPARISON.md)

## Acknowledgments

- Yu Liang, Song Liu, and Hong Hu for the SQLRight research and implementation
- The SQLancer team for pioneering oracle-based database testing
- The SQLsmith project for coverage-guided SQL fuzzing

## Contributing

To add a new oracle:

1. Implement the `Oracle` interface in `internal/oracles/`
2. Add comprehensive tests in `internal/oracles/oracle_test.go`
3. Update this documentation
4. Add usage examples

Example:

```go
// NewMyOracle creates a new custom oracle
func NewMyOracle() *MyOracle {
    return &MyOracle{
        BaseOracle: NewBaseOracle("MY_ORACLE"),
    }
}

func (m *MyOracle) IsApplicable(query string) bool {
    // Check if query is suitable
    return true
}

func (m *MyOracle) TransformQuery(baseQuery string) []string {
    // Create equivalent queries
    return []string{transformed}
}

func (m *MyOracle) CompareResults(results []QueryResult) ComparisonResult {
    // Compare results and return Pass/Fail/Error
    return Pass
}
```

## License

This integration follows the same MIT license as sqlfuse.
