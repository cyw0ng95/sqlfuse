# PRAGMA Support Documentation

This document describes the PRAGMA statement support for different database flavors in SQLsmith-go.

## Flavor-Aware PRAGMA Generation

The PRAGMA generator (`internal/stmts/stmts/pragma.go`) generates flavor-specific PRAGMA statements based on the database being targeted.

## Supported Flavors

### Turso LibSQL / Default Flavor

**Total pragmas**: 18

Turso LibSQL has limited PRAGMA support compared to full SQLite3. The generator only includes pragmas with "Yes" or "Partial" support according to the [Turso compatibility documentation](https://github.com/tursodatabase/turso/blob/main/COMPAT.md#pragma).

#### Read-Only Pragmas (Query Information)
- `database_list` - List attached databases
- `freelist_count` - Number of unused pages
- `integrity_check` - Database integrity verification
- `page_count` - Total number of pages
- `pragma_list` - List all pragmas
- `table_info` - Table structure information (without table name parameter)

#### Read-Write Pragmas (Configuration)
- `application_id` - Application ID (integer value)
- `cache_size` - Page cache size (integer value)
- `encoding` - Text encoding (UTF-8, UTF-16, UTF-16le, UTF-16be)
- `journal_mode` - Journal mode (WAL only for Turso)
- `legacy_file_format` - Legacy file format flag (ON/OFF)
- `max_page_count` - Maximum page count (integer value)
- `page_size` - Database page size (integer value)
- `query_only` - Read-only mode (ON/OFF)
- `schema_version` - Schema version number (write is noop in defensive mode)
- `synchronous` - Synchronous mode (OFF or FULL only)
- `user_version` - User-defined version (integer value)
- `wal_checkpoint` - WAL checkpoint (no mode parameter)

### go-sqlite3 Flavor

**Total pragmas**: 47

The go-sqlite3 flavor supports full SQLite3 PRAGMA functionality as documented in the [SQLite documentation](https://sqlite.org/pragma.html).

#### Read-Only Pragmas (Query Information)
All Turso pragmas plus:
- `collation_list` - List available collations
- `compile_options` - SQLite compile-time options
- `parser_trace` - Parser tracing (debugging)
- `quick_check` - Quick integrity check

#### Read-Write Pragmas (Configuration)
All Turso pragmas plus:

**Auto-vacuum and Storage**
- `auto_vacuum` - Auto-vacuum mode (NONE, FULL, INCREMENTAL)
- `incremental_vacuum` - Incremental vacuum with optional page count

**Boolean/Toggle Settings**
- `automatic_index` - Automatic index creation (ON/OFF)
- `case_sensitive_like` - Case-sensitive LIKE operator (ON/OFF)
- `cell_size_check` - Cell size validation (ON/OFF)
- `checkpoint_fullfsync` - Full fsync on checkpoint (ON/OFF)
- `foreign_keys` - Foreign key constraints (ON/OFF)
- `fullfsync` - Full fsync mode (ON/OFF)
- `ignore_check_constraints` - Ignore CHECK constraints (ON/OFF)
- `legacy_alter_table` - Legacy ALTER TABLE behavior (ON/OFF)
- `read_uncommitted` - Read uncommitted isolation (ON/OFF)
- `recursive_triggers` - Recursive trigger support (ON/OFF)
- `reverse_unordered_selects` - Reverse query result order (ON/OFF)
- `secure_delete` - Secure deletion (ON/OFF)
- `trusted_schema` - Schema trust mode (ON/OFF)
- `writable_schema` - Allow schema modification (ON/OFF)

**Numeric Settings**
- `busy_timeout` - Busy timeout in milliseconds (0-60000)
- `cache_spill` - Cache spill threshold (integer or -1)
- `journal_size_limit` - Journal size limit (-1 or positive)
- `mmap_size` - Memory-mapped I/O size (0 to 1GB)
- `soft_heap_limit` - Soft heap limit (0 to 100MB)
- `threads` - Worker thread count (0-7)
- `wal_autocheckpoint` - WAL auto-checkpoint threshold (pages)

**Mode Settings**
- `journal_mode` - Journal mode (DELETE, TRUNCATE, PERSIST, MEMORY, WAL, OFF)
- `locking_mode` - Locking mode (NORMAL, EXCLUSIVE)
- `synchronous` - Synchronous mode (OFF, NORMAL, FULL, EXTRA)
- `temp_store` - Temporary storage (DEFAULT, FILE, MEMORY)

**Parameterized Pragmas**
- `table_info(table_name)` - Table structure with table name
- `wal_checkpoint(mode)` - WAL checkpoint with mode (PASSIVE, FULL, RESTART, TRUNCATE)

## Implementation Details

### Flavor Detection

The pragma generator checks `flavor.Name()` to determine which pragmas to generate:
- If `flavor.Name() == "go-sqlite3"` → Generate full SQLite3 pragma set
- Otherwise → Generate Turso-compatible subset (conservative default)

### SQL Generation Examples

**Turso/Default:**
```sql
PRAGMA journal_mode = WAL;
PRAGMA synchronous = OFF;
PRAGMA table_info;
PRAGMA wal_checkpoint;
```

**go-sqlite3:**
```sql
PRAGMA journal_mode = DELETE;
PRAGMA synchronous = NORMAL;
PRAGMA table_info(users);
PRAGMA wal_checkpoint(PASSIVE);
PRAGMA foreign_keys = ON;
PRAGMA mmap_size = 268435456;
PRAGMA auto_vacuum = INCREMENTAL;
```

## Testing

Tests are located in `internal/stmts/stmts/stmt_test.go`:
- `TestGenPragmaTursoCompatibility` - Verifies Turso pragma constraints
- `TestGenPragmaGoSQLite3Support` - Verifies default flavor doesn't leak go-sqlite3 pragmas
- `TestGenPragmaWithGoSQLite3Flavor` - Verifies go-sqlite3 flavor generates extended pragmas

## Future Enhancements

Potential improvements for PRAGMA support:
1. Add more validation for pragma value ranges
2. Support database-specific pragma extensions (Turso vector pragmas, etc.)
3. Add pragma dependency checking (some pragmas require others to be set first)
4. Implement pragma conflict detection (mutually exclusive settings)
5. Add pragma effectiveness scoring (how likely a pragma is to trigger bugs)
