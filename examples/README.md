# SQLfuse Examples

This directory contains examples demonstrating various features of sqlfuse.

## Examples

### impedance_example.go

Demonstrates the impedance matching and statistics tracking features inspired by the original SQLsmith.

**Note**: This example uses internal packages for demonstration purposes. In a production setting, these features would be integrated into the executors or exposed through public APIs.

**Features shown:**
- Enabling impedance matching to automatically blacklist problematic statement types
- Configuring error rate thresholds and minimum observations
- Tracking generation and execution statistics
- Recording successes and failures per statement type
- Generating comprehensive reports

**How to run:**
```bash
cd examples
go run impedance_example.go
```

**Expected output:**
```
SQLfuse Example: Impedance Matching & Statistics
====================================================

Progress: 20 queries | Gen: 1234.5/s | Exec: 987.6/s | Errors: 5.23%
Progress: 40 queries | Gen: 1189.3/s | Exec: 945.2/s | Errors: 4.87%
...

Final Statistics:
================
Generation Statistics:
================================================================================
Queries generated: 100 (XXX.XX gen/s)
Queries executed:  100 (XXX.XX exec/s)
Successful:        95
Syntax errors:     0
Execution errors:  5
...

Impedance Report:
================
Production                               Failed       OK  Retries  Limited  ErrorRate Status
--------------------------------------------------------------------------------
select_basic                                  0       45        0        0      0.00% OK
insert                                        2       18        0        0     10.00% OK
...

Blacklisted Statement Types:
============================
No statement types blacklisted (good compatibility!)
```

## Dependencies

All examples require:
- Go 1.24.9 or later
- SQLite3 (for examples using go-sqlite3)

The examples use the `github.com/mattn/go-sqlite3` package which requires CGo.

## Building Examples

```bash
# Build a specific example
go build -o impedance_example examples/impedance_example.go

# Run directly
go run examples/impedance_example.go
```

## Adding New Examples

To add a new example:

1. Create a new `.go` file in this directory
2. Use `package main` and implement a `main()` function
3. Import necessary packages from `sqlfuse/internal/*`
4. Add documentation to this README

## Related Documentation

- [SQLSMITH_COMPARISON.md](../docs/SQLSMITH_COMPARISON.md) - Comparison with original SQLsmith
- [ARCHITECTURE.md](../docs/ARCHITECTURE.md) - Project architecture
- [README.md](../README.md) - Main project README
