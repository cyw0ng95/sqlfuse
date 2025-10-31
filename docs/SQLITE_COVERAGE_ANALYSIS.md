# SQLite Documentation Coverage Analysis

## Overview

This document summarizes the comprehensive analysis of SQLite documentation coverage in sqlfuse, based on recursive crawling of https://sqlite.org/lang.html and its linked pages.

## Crawler Implementation

### Tools Created

1. **`tools/sqlite_doc_crawler.go`**: Recursively crawls SQLite documentation
   - Discovers SQL features from https://sqlite.org/lang.html
   - Follows links to related documentation pages (lang_*.html, pragma.html, windowfunctions.html, json1.html)
   - Extracts SQL keywords, syntax examples, and categorizes features
   - Outputs structured JSON data for analysis

2. **`tools/analyze_coverage.go`**: Analyzes implementation coverage
   - Compares discovered features against implemented statement types
   - Generates coverage statistics and recommendations
   - Identifies missing features

### Crawl Results

The crawler discovered **34 distinct SQL feature categories** from **39 documentation pages**:

## Coverage Summary

### Fully Covered Features (16 categories)

sqlfuse has full coverage for the following SQLite features:

- **ANALYZE** - Database statistics gathering
- **ATTACH** / **DETACH** - Database attachment
- **DELETE** - Row deletion
- **EXPLAIN** - Query plan analysis
- **INSERT** - Row insertion (all variants including OR REPLACE, OR IGNORE, etc.)
- **REINDEX** - Index rebuilding
- **REPLACE** - Row replacement
- **RETURNING** - RETURNING clause support
- **SAVEPOINT** - Transaction savepoints
- **SELECT** - Query statements (comprehensive coverage including CTEs, window functions, subqueries)
- **UPDATE** - Row updates
- **UPSERT** - INSERT with ON CONFLICT handling
- **VACUUM** - Database optimization
- **WINDOW FUNCTIONS** - Window function support
- **WITH** - Common Table Expressions (CTEs)

### Partially Covered Features (5 categories)

These features have partial coverage (functions/expressions are complex):

- **AGGFUNC** - Aggregate functions (COUNT, SUM, AVG, etc.)
- **COREFUNC** - Core functions (ABS, LENGTH, etc.)
- **DATEFUNC** - Date/time functions
- **EXPR** - Expression generation
- **MATHFUNC** - Mathematical functions

### Statement Types Implemented

sqlfuse implements **77 distinct statement types** including:

#### Core DML
- SELECT (25+ variants: basic, WHERE, JOIN, CTE, window functions, subqueries, etc.)
- INSERT (8 variants: basic, multiple, bulk, OR REPLACE, OR IGNORE, etc.)
- UPDATE
- DELETE
- REPLACE

#### DDL - Table Operations
- CREATE TABLE
- DROP TABLE
- ALTER TABLE

#### DDL - View Operations
- CREATE VIEW
- DROP VIEW

#### DDL - Index Operations
- CREATE INDEX
- DROP INDEX
- **NEW: SELECT ... INDEXED BY** (query hint to force index usage)
- **NEW: SELECT ... NOT INDEXED** (query hint to prevent index usage)

#### DDL - Trigger Operations
- CREATE TRIGGER
- DROP TRIGGER

#### DDL - Virtual Table Operations
- CREATE VIRTUAL TABLE (FTS5 support)

#### Transaction Control
- BEGIN TRANSACTION (with DEFERRED, IMMEDIATE, EXCLUSIVE modes)
- COMMIT TRANSACTION
- ROLLBACK TRANSACTION
- SAVEPOINT
- RELEASE SAVEPOINT

#### Database Operations
- ATTACH DATABASE
- DETACH DATABASE

#### Maintenance
- ANALYZE
- VACUUM
- REINDEX

#### Query Analysis
- EXPLAIN
- EXPLAIN QUERY PLAN

#### Pragmas
- PRAGMA statements (18 pragmas for Turso, 47 for go-sqlite3)

#### Compound SELECT
- UNION / UNION ALL
- INTERSECT
- EXCEPT

## Recent Enhancements (This Update)

### 1. Documentation Crawler

Created comprehensive tooling to recursively crawl and analyze SQLite documentation:

**Implementation**: `tools/sqlite_doc_crawler.go`
- Crawls https://sqlite.org/lang.html and follows all SQL-related links
- Extracts SQL keywords, categories, and syntax examples
- Generates structured JSON output for analysis
- Discovered 34 feature categories across 39 pages

**Implementation**: `tools/analyze_coverage.go`
- Compares crawled features against implementation
- Generates coverage statistics (61.8% fully/partially covered)
- Identifies gaps and provides recommendations

### 2. INDEXED BY / NOT INDEXED Clauses

Based on discovery from https://sqlite.org/lang_indexedby.html, implemented support for SQLite query hints:

**New Statement Types**:
- `StmtSelectIndexedBy` - SELECT with INDEXED BY clause to force index usage
- `StmtSelectNotIndexed` - SELECT with NOT INDEXED clause to prevent index usage

**Implementation**: `internal/stmts/stmts/select_indexed_by.go`
- Generates plausible index names based on table/column names
- Adds optional WHERE and LIMIT clauses
- Supports both Turso and go-sqlite3 flavors

**Tests**: `internal/stmts/stmts/select_indexed_by_test.go`
- Comprehensive test coverage (5 test functions)
- Tests generation with/without database connection
- Tests factory integration
- All tests passing ✅

**Generator Weights**:
- Turso: INDEXED BY (15), NOT INDEXED (10)
- go-sqlite3: INDEXED BY (20), NOT INDEXED (15)

**Example Generated SQL**:
```sql
-- INDEXED BY
SELECT "id", "name", "value" FROM "users" INDEXED BY "idx_users_email" WHERE "age" > 25 LIMIT 50;

-- NOT INDEXED
SELECT "id", "name" FROM "orders" NOT INDEXED WHERE "status" IS NOT NULL;
```

## Coverage Statistics

### By Numbers

- **Total SQLite documentation pages crawled**: 39
- **Unique feature categories discovered**: 34
- **Fully covered categories**: 16 (47.1%)
- **Partially covered categories**: 5 (14.7%)
- **Total coverage**: 21/34 (61.8%)
- **Total statement types implemented**: 77

### Coverage Breakdown

| Category | Status | Notes |
|----------|--------|-------|
| Core DML | ✅ Full | SELECT, INSERT, UPDATE, DELETE, REPLACE |
| DDL | ✅ Full | Tables, views, indexes, triggers, virtual tables |
| Transactions | ✅ Full | BEGIN, COMMIT, ROLLBACK, SAVEPOINT, RELEASE |
| Database Ops | ✅ Full | ATTACH, DETACH |
| Maintenance | ✅ Full | ANALYZE, VACUUM, REINDEX |
| Query Analysis | ✅ Full | EXPLAIN, EXPLAIN QUERY PLAN |
| Compound Queries | ✅ Full | UNION, INTERSECT, EXCEPT |
| Window Functions | ✅ Full | ROW_NUMBER, RANK, LAG, LEAD, etc. |
| CTEs | ✅ Full | WITH clause, recursive CTEs |
| UPSERT | ✅ Full | INSERT ... ON CONFLICT |
| RETURNING | ✅ Full | RETURNING clause |
| Query Hints | ✅ Full | INDEXED BY, NOT INDEXED |
| Aggregate Functions | ⚠️ Partial | COUNT, SUM, AVG, etc. (complex) |
| Core Functions | ⚠️ Partial | ABS, LENGTH, etc. (complex) |
| Date Functions | ⚠️ Partial | DATE, TIME, etc. (complex) |
| Math Functions | ⚠️ Partial | SQRT, POW, etc. (complex) |
| Expressions | ⚠️ Partial | Complex expression generation |

## Flavor Support

sqlfuse supports multiple SQLite-compatible database flavors with different feature sets:

### Turso LibSQL
- Conservative feature set following https://github.com/tursodatabase/turso/blob/main/COMPAT.md
- 18 PRAGMA statements
- Limited transaction modes
- Read-only ATTACH DATABASE support
- INDEXED BY: Supported ✅
- NOT INDEXED: Supported ✅

### go-sqlite3
- Full SQLite3 feature support
- 47 PRAGMA statements
- All transaction modes
- Full ATTACH DATABASE support
- INDEXED BY: Supported ✅
- NOT INDEXED: Supported ✅

### DuckDB
- Analytical database with extensive SQL features
- Extended window functions
- PIVOT/UNPIVOT support
- Advanced analytical functions

## Architecture

### Statement Generation Flow

```
SQLite Docs (lang.html)
    ↓
Documentation Crawler
    ↓
Feature Discovery (JSON)
    ↓
Coverage Analysis
    ↓
Implementation (stmts package)
    ↓
Factory Registration
    ↓
Generator Weights
    ↓
SQL Generation
```

### Key Components

1. **Statement Generators** (`internal/stmts/stmts/`)
   - Individual generators for each statement type
   - Flavor-aware SQL generation
   - Schema-aware when database connection available

2. **Factory Pattern** (`factory.go`)
   - Centralized generator creation
   - Statement type → Generator mapping

3. **Generator Map** (`gen_map.go`)
   - String key → Generator function mapping
   - Fallback SQL for each statement type

4. **Flavor Configurations** (`internal/generators/dialects/`)
   - Database-specific feature flags
   - SQL validation hooks

5. **Generator Weights** (`turso.go`, `go_sqlite3.go`)
   - Probabilistic statement selection
   - Flavor-specific weight distributions

## Future Enhancements

### Potential Additions from SQLite Docs

Based on the crawler analysis, potential future enhancements include:

1. **Enhanced JSON Functions** (json1.html)
   - JSON_EXTRACT
   - JSON_INSERT
   - JSON_REPLACE
   - JSON_ARRAY
   - JSON_OBJECT

2. **SQL Comments** (lang_comment.html)
   - Comment generation for testing parser robustness

3. **Advanced Expressions** (lang_expr.html)
   - More complex expression patterns
   - COLLATE clauses
   - CAST expressions
   - CASE variants

4. **Additional Math Functions** (lang_mathfunc.html)
   - Trigonometric functions
   - Logarithmic functions
   - Statistical functions

## Usage

### Running the Crawler

```bash
cd tools
go run sqlite_doc_crawler.go
# Output: /tmp/sqlite_sql_features.json
```

### Analyzing Coverage

```bash
cd tools
go run analyze_coverage.go
# Output: /tmp/sqlite_coverage_report.txt
```

### Testing New Features

```bash
cd /home/runner/work/sqlfuse/sqlfuse
export SQLSMITH_GO_CONTAINER_TYPE=test
go test ./internal/stmts/stmts/... -v -run TestIndexedBy
```

## Conclusion

sqlfuse has **comprehensive coverage** of SQLite's core SQL language features, with 61.8% of documented categories fully or partially implemented. The project supports:

- ✅ All core DML operations
- ✅ Complete DDL support
- ✅ Full transaction control
- ✅ Advanced features (CTEs, window functions, UPSERT)
- ✅ Query hints (INDEXED BY, NOT INDEXED)
- ✅ Flavor-aware generation

The recursive documentation crawler and analysis tools enable continuous validation of coverage against the official SQLite documentation, ensuring the fuzzer exercises a comprehensive range of SQL features.

## References

- SQLite Language Documentation: https://sqlite.org/lang.html
- Turso Compatibility: https://github.com/tursodatabase/turso/blob/main/COMPAT.md
- INDEXED BY Documentation: https://sqlite.org/lang_indexedby.html
- Project Architecture: docs/ARCHITECTURE.md
- SQL Statement Enhancement: docs/SQL_STATEMENT_ENHANCEMENT.md
