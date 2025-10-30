package stmts

import (
	"database/sql"
	"sqlsmith-go/internal/common"
)

// BuildGeneratorFuncs returns a map[string]generatorFunc for all built-in statement types.
// Keys match the string values used by turso.StmtType constants.
// This function uses the Factory pattern to create generators and simplify the code.
func BuildGeneratorFuncs(lcg *common.LCG, maxRecursionDepth int, flavorConfig FlavorConfig) map[string]func(db *sql.DB) (string, error) {
	factory := NewStmtGeneratorFactory(lcg, maxRecursionDepth, flavorConfig)
	m := make(map[string]func(db *sql.DB) (string, error), len(AllStmtTypes)+1)

	// Create generator function using factory pattern
	createGenFunc := func(stmtType StmtType, fallback string) func(db *sql.DB) (string, error) {
		return func(db *sql.DB) (string, error) {
			stmt, err := factory.GenerateStmt(db, stmtType)
			if err != nil {
				return fallback, err
			}
			return stmt.SQL(), nil
		}
	}

	// Register all statement types using the factory
	m["pragma"] = createGenFunc(StmtPragma, "PRAGMA integrity_check;")
	m["insert"] = createGenFunc(StmtInsert, "INSERT INTO sqlite_master DEFAULT VALUES;")
	m["insert_multiple"] = createGenFunc(StmtInsertMultiple, "INSERT INTO sqlite_master DEFAULT VALUES;")
	m["insert_bulk"] = createGenFunc(StmtInsertBulk, "INSERT INTO sqlite_master DEFAULT VALUES;")
	m["insert_or_replace"] = createGenFunc(StmtInsertOrReplace, "INSERT INTO sqlite_master DEFAULT VALUES;")
	m["insert_or_ignore"] = createGenFunc(StmtInsertOrIgnore, "INSERT INTO sqlite_master DEFAULT VALUES;")
	m["insert_or_abort"] = createGenFunc(StmtInsertOrAbort, "INSERT INTO sqlite_master DEFAULT VALUES;")
	m["insert_or_rollback"] = createGenFunc(StmtInsertOrRollback, "INSERT INTO sqlite_master DEFAULT VALUES;")
	m["insert_or_fail"] = createGenFunc(StmtInsertOrFail, "INSERT INTO sqlite_master DEFAULT VALUES;")
	m["update"] = createGenFunc(StmtUpdate, "UPDATE sqlite_master SET name = 'fallback';")
	m["delete"] = createGenFunc(StmtDelete, "DELETE FROM sqlite_master WHERE 0;")

	// SELECT statements
	m["select_basic"] = createGenFunc(StmtSelectBasic, "SELECT 1")
	m["select_where"] = createGenFunc(StmtSelectWhere, "SELECT 1")
	m["select_where_complex"] = createGenFunc(StmtSelectWhereComplex, "SELECT 1")
	m["select_where_in"] = createGenFunc(StmtSelectWhereIn, "SELECT 1")
	m["select_subquery"] = createGenFunc(StmtSelectSubquery, "SELECT 1")
	m["select_case"] = createGenFunc(StmtSelectCase, "SELECT 1")
	m["select_aggregate_complex"] = createGenFunc(StmtSelectAggregateComplex, "SELECT 1")
	m["select_like"] = createGenFunc(StmtSelectLike, "SELECT 1")
	m["select_limit"] = createGenFunc(StmtSelectLimit, "SELECT 1")
	m["select_order"] = createGenFunc(StmtSelectOrder, "SELECT 1")
	m["select_group"] = createGenFunc(StmtSelectGroup, "SELECT 1")
	m["select_having"] = createGenFunc(StmtSelectHaving, "SELECT 1")
	m["select_join"] = createGenFunc(StmtSelectJoin, "SELECT 1")
	m["select_crossjoin"] = createGenFunc(StmtSelectCross, "SELECT 1")
	m["select_innerjoin"] = createGenFunc(StmtSelectInner, "SELECT 1")
	m["select_outerjoin"] = createGenFunc(StmtSelectOuter, "SELECT 1")
	m["select_joinusing"] = createGenFunc(StmtSelectJoinUsing, "SELECT 1")
	m["select_naturaljoin"] = createGenFunc(StmtSelectNatural, "SELECT 1")
	m["select_recursive"] = createGenFunc(StmtSelectRecursive, "SELECT 1")
	m["select_nested_case"] = createGenFunc(StmtSelectNestedCase, "SELECT 1")
	m["select_complex_join"] = createGenFunc(StmtSelectComplexJoin, "SELECT 1")
	m["select_deeply_nested"] = createGenFunc(StmtSelectDeeplyNested, "SELECT 1")
	m["select_window"] = createGenFunc(StmtSelectWindow, "SELECT 1")
	m["select_multiple_windows"] = createGenFunc(StmtSelectMultipleWindows, "SELECT 1")
	m["select_cte"] = createGenFunc(StmtSelectCTE, "SELECT 1")
	m["select_multiple_cte"] = createGenFunc(StmtSelectMultipleCTE, "SELECT 1")
	m["select_recursive_cte"] = createGenFunc(StmtSelectRecursiveCTE, "SELECT 1")
	m["select_json"] = createGenFunc(StmtSelectJSON, "SELECT 1")
	m["select_uuid"] = createGenFunc(StmtSelectUUID, "SELECT 1")
	m["select_regexp"] = createGenFunc(StmtSelectRegexp, "SELECT 1")
	m["select_vector"] = createGenFunc(StmtSelectVector, "SELECT 1")
	m["select_time"] = createGenFunc(StmtSelectTime, "SELECT 1")

	// DDL statements
	m["create_table"] = createGenFunc(StmtCreateTable, "CREATE TABLE IF NOT EXISTS fallback (id INTEGER);")
	m["drop_table"] = createGenFunc(StmtDropTable, "DROP TABLE IF EXISTS fallback;")
	m["alter_table"] = createGenFunc(StmtAlterTable, "ALTER TABLE fallback RENAME TO fallback2;")
	m["create_view"] = createGenFunc(StmtCreateView, "CREATE VIEW IF NOT EXISTS fallback AS SELECT 1;")
	m["drop_view"] = createGenFunc(StmtDropView, "DROP VIEW IF EXISTS fallback;")
	m["create_index"] = createGenFunc(StmtCreateIndex, "CREATE INDEX IF NOT EXISTS idx_fallback ON fallback (id);")
	m["drop_index"] = createGenFunc(StmtDropIndex, "DROP INDEX IF EXISTS idx_fallback;")
	m["create_virtual_table"] = createGenFunc(StmtCreateVirtualTable, "CREATE VIRTUAL TABLE IF NOT EXISTS fallback USING fts5(content);")

	// Transaction control
	m["attach"] = createGenFunc(StmtAttach, "ATTACH DATABASE ':memory:' AS fallback;")
	m["detach"] = createGenFunc(StmtDetach, "DETACH DATABASE fallback;")
	m["begin"] = createGenFunc(StmtBegin, "BEGIN;")
	m["commit"] = createGenFunc(StmtCommit, "COMMIT;")
	m["rollback"] = createGenFunc(StmtRollback, "ROLLBACK;")

	// Query analysis
	m["explain"] = createGenFunc(StmtExplain, "EXPLAIN SELECT 1;")
	m["explain_query_plan"] = createGenFunc(StmtExplainQueryPlan, "EXPLAIN QUERY PLAN SELECT 1;")

	// Database maintenance
	m["analyze"] = createGenFunc(StmtAnalyze, "ANALYZE;")
	m["vacuum"] = createGenFunc(StmtVacuum, "VACUUM;")
	m["reindex"] = createGenFunc(StmtReindex, "REINDEX;")

	// Transaction savepoints
	m["savepoint"] = createGenFunc(StmtSavepoint, "SAVEPOINT sp_fallback;")
	m["release"] = createGenFunc(StmtRelease, "RELEASE sp_fallback;")

	// Triggers
	m["create_trigger"] = createGenFunc(StmtCreateTrigger, "CREATE TRIGGER IF NOT EXISTS trg_fallback BEFORE INSERT ON fallback BEGIN SELECT 1; END;")
	m["drop_trigger"] = createGenFunc(StmtDropTrigger, "DROP TRIGGER IF EXISTS trg_fallback;")

	// Compound SELECT statements
	m["select_union"] = createGenFunc(StmtSelectUnion, "SELECT 1 UNION SELECT 2;")
	m["select_intersect"] = createGenFunc(StmtSelectIntersect, "SELECT 1 INTERSECT SELECT 2;")
	m["select_except"] = createGenFunc(StmtSelectExcept, "SELECT 1 EXCEPT SELECT 2;")

	// DuckDB-specific statements
	m["copy"] = createGenFunc(StmtCopy, "COPY (SELECT 1) TO '/tmp/fallback.csv' (FORMAT CSV);")
	m["set"] = createGenFunc(StmtSet, "SET threads = 1;")
	m["reset"] = createGenFunc(StmtReset, "RESET threads;")
	m["create_schema"] = createGenFunc(StmtCreateSchema, "CREATE SCHEMA IF NOT EXISTS fallback;")
	m["drop_schema"] = createGenFunc(StmtDropSchema, "DROP SCHEMA IF EXISTS fallback;")
	m["create_sequence"] = createGenFunc(StmtCreateSequence, "CREATE SEQUENCE IF NOT EXISTS seq_fallback;")
	m["drop_sequence"] = createGenFunc(StmtDropSequence, "DROP SEQUENCE IF EXISTS seq_fallback;")
	m["create_macro"] = createGenFunc(StmtCreateMacro, "CREATE MACRO fallback(x) AS x + 1;")
	m["drop_macro"] = createGenFunc(StmtDropMacro, "DROP MACRO IF EXISTS fallback;")
	m["create_type"] = createGenFunc(StmtCreateType, "CREATE TYPE fallback AS ENUM ('a', 'b');")
	m["drop_type"] = createGenFunc(StmtDropType, "DROP TYPE IF EXISTS fallback;")
	m["describe"] = createGenFunc(StmtDescribe, "DESCRIBE SELECT 1;")
	m["show"] = createGenFunc(StmtShow, "SHOW TABLES;")
	m["summarize"] = createGenFunc(StmtSummarize, "SUMMARIZE SELECT 1;")
	m["use"] = createGenFunc(StmtUse, "USE fallback;")
	m["call"] = createGenFunc(StmtCall, "CALL fallback();")
	m["checkpoint"] = createGenFunc(StmtCheckpoint, "CHECKPOINT;")
	m["export_database"] = createGenFunc(StmtExportDatabase, "EXPORT DATABASE '/tmp/export' (FORMAT PARQUET);")
	m["import_database"] = createGenFunc(StmtImportDatabase, "IMPORT DATABASE '/tmp/import';")
	m["prepare"] = createGenFunc(StmtPrepare, "PREPARE stmt AS SELECT 1;")
	m["execute"] = createGenFunc(StmtExecute, "EXECUTE stmt;")
	m["pivot"] = createGenFunc(StmtPivot, "PIVOT (VALUES (1, 'a', 10)) AS t(id, cat, val) ON cat USING sum(val);")
	m["unpivot"] = createGenFunc(StmtUnpivot, "UNPIVOT (VALUES (1, 10, 20)) AS t(id, a, b) ON a, b INTO NAME col VALUE val;")
	m["merge_into"] = createGenFunc(StmtMergeInto, "MERGE INTO target USING source ON target.id = source.id WHEN MATCHED THEN UPDATE SET value = source.value;")
	m["select_qualify"] = createGenFunc(StmtQualify, "SELECT * FROM (VALUES (1, 100)) AS t(id, val) QUALIFY ROW_NUMBER() OVER (ORDER BY val) = 1;")
	m["alter_database"] = createGenFunc(StmtAlterDatabase, "ALTER DATABASE fallback RENAME TO fallback2;")
	m["alter_view"] = createGenFunc(StmtAlterView, "ALTER VIEW fallback RENAME TO fallback2;")
	m["create_secret"] = createGenFunc(StmtCreateSecret, "CREATE SECRET fallback (TYPE S3, KEY_ID 'key', SECRET 'secret');")
	m["drop_secret"] = createGenFunc(StmtDropSecret, "DROP SECRET IF EXISTS fallback;")
	m["load_install"] = createGenFunc(StmtLoadInstall, "LOAD httpfs;")
	m["comment_on"] = createGenFunc(StmtCommentOn, "COMMENT ON TABLE fallback IS 'comment';")
	m["profiling"] = createGenFunc(StmtProfiling, "PRAGMA enable_profiling;")
	m["set_variable"] = createGenFunc(StmtSetVariable, "SET VARIABLE var = 1;")

	return m
}
