# Validator: Turso vs SQLite3

This validator compares the behavior of Turso LibSQL and SQLite3 by executing the same SQL statements on both databases and comparing results. It's designed to detect compatibility issues and behavioral differences between the two database implementations.

## Overview

The validator:
1. Opens two in-memory database connections (Turso and SQLite3)
2. Initializes both databases with identical schema from `turso/init.sql`
3. Generates random SQL statements focusing on basic DML operations
4. Executes each statement on both databases
5. Compares execution results, affected rows, and table data
6. Reports any discrepancies as potential bugs

## Usage

```bash
# Basic usage with 100 queries
./validator_turso_sqlite3

# With custom seed for reproducibility
./validator_turso_sqlite3 --seed 42 --queries 100

# Verbose mode to see all SQL statements
./validator_turso_sqlite3 --verbose --queries 50

# Stop on first error instead of continuing
./validator_turso_sqlite3 --stop-on-error --queries 100

# Use custom initialization SQL
./validator_turso_sqlite3 --init-sql /path/to/schema.sql
```

## Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--seed` | Random seed (0 = use timestamp) | `0` |
| `--queries` | Number of queries to execute | `100` |
| `--verbose` | Show all SQL statements | `false` |
| `--stop-on-error` | Stop on first error | `false` |
| `--init-sql` | Path to initialization SQL file | `/opt/assets/turso/init.sql` |

## SQL Generation Strategy

The validator focuses on **basic SQL operations** to minimize expected differences:

**High Weight (Frequently Generated):**
- `INSERT` statements (300)
- `SELECT` basic queries (250)
- `UPDATE` statements (200)
- `SELECT WHERE` queries (200)
- `DELETE` statements (150)

**Lower Weight:**
- `SELECT JOIN` queries (50)
- Complex `WHERE` clauses (40)

**Disabled (Weight = 0):**
- DDL statements (CREATE, DROP, ALTER)
- Transaction control (BEGIN, COMMIT, ROLLBACK)
- Advanced features (PRAGMA, ANALYZE, VACUUM)
- Recursive CTEs and window functions

## Bug Detection

The validator detects several types of issues:

### 1. Error State Differences
When one database succeeds but the other fails:
```
BUG DETECTED: Error state differs
Query: INSERT OR ROLLBACK INTO "misc" ...
Turso error: Parse error: ON CONFLICT clause is not supported
SQLite3 error: <nil>
```

### 2. Result Count Differences
When query results have different row counts:
```
BUG DETECTED: Results differ
Query: SELECT * FROM users WHERE ...
Turso rows: 5, SQLite3 rows: 7
```

### 3. Affected Rows Differences
When DML statements affect different numbers of rows:
```
BUG DETECTED: Rows affected differs
Query: UPDATE users SET ...
Turso: 3 rows, SQLite3: 5 rows
```

### 4. Table Data Differences
When table contents diverge:
```
BUG DETECTED: Table users has different data
Turso hash: abc123..., SQLite3 hash: def456...
```

## Output

The validator provides a summary at the end:

```
=== Validation Summary ===
Total queries: 100
Errors encountered: 0
Bugs detected: 15
VALIDATION FAILED: 15 bugs detected
```

Exit codes:
- `0`: Validation passed (no bugs detected)
- `1`: Validation failed (bugs detected or errors occurred)

## Known Compatibility Issues

The validator commonly detects these known differences between Turso and SQLite3:

1. **ON CONFLICT clauses**: Not supported in Turso (INSERT OR ROLLBACK, etc.)
2. **Recursive CTEs**: Not supported in Turso
3. **ANALYZE statement**: Not supported on tables with primary keys in Turso
4. **CHECK constraints**: Different validation behavior
5. **ATTACH DATABASE**: Different behavior
6. **Date/Time handling**: Some differences in DATETIME default values

## Development

### Building

```bash
cd cmd/executors/validator_turso_sqlite3
go build -o validator_turso_sqlite3
```

### Adding to Build System

The validator is included in the main build script:

```bash
export SQLSMITH_GO_CONTAINER_TYPE=test
bash build.sh
# Output: output/validator_turso_sqlite3
```

## Use Cases

### 1. Compatibility Testing
Verify that SQL workloads are compatible with Turso:
```bash
./validator_turso_sqlite3 --queries 1000 --seed 42
```

### 2. Regression Testing
Ensure changes don't introduce new incompatibilities:
```bash
# Before changes
./validator_turso_sqlite3 --seed 12345 > baseline.log

# After changes
./validator_turso_sqlite3 --seed 12345 > updated.log

# Compare
diff baseline.log updated.log
```

### 3. Bug Discovery
Find edge cases and compatibility issues:
```bash
# High volume testing
for i in {1..10}; do
  ./validator_turso_sqlite3 --queries 500 --seed $i
done
```

## Implementation Details

### Result Comparison
- SELECT queries: Compares full result sets row by row
- DML queries: Compares number of affected rows
- Both: Checks error states for consistency

### Table Data Comparison
- Performed periodically (every 20 queries)
- Final comparison at the end
- Uses SHA-256 hash of sorted, deterministic row representation
- Detects data drift between databases

### Thread Safety
- Uses two separate database connections
- Sequential execution (no concurrency)
- Deterministic with same seed

## Limitations

1. Only compares in-memory databases
2. Does not test file-based database features
3. Does not test concurrent access
4. Some expected differences may be reported as bugs
5. Hash comparison may not detect all semantic differences

## Future Improvements

- Filter known compatibility differences
- Support for custom SQL generation weights
- Parallel execution on both databases
- More detailed difference reporting
- Integration with CI/CD pipelines
- Support for file-based databases
