package stmts

import (
	"database/sql"
	"sqlsmith-go/internal/common"
)

// BuildGeneratorFuncs returns a map[string]generatorFunc for all built-in statement types.
// Keys match the string values used by turso.StmtType constants.
func BuildGeneratorFuncs(lcg *common.LCG, maxRecursionDepth int, flavorConfig FlavorConfig) map[string]func(db *sql.DB) (string, error) {
	m := make(map[string]func(db *sql.DB) (string, error), 64)

	m["pragma"] = func(db *sql.DB) (string, error) {
		return GenPragma(lcg).SQL(), nil
	}

	m["insert"] = func(db *sql.DB) (string, error) {
		s, err := GenInsert(db, lcg)
		if err != nil {
			return "INSERT INTO sqlite_master DEFAULT VALUES;", err
		}
		return s.SQL(), nil
	}
	m["insert_multiple"] = func(db *sql.DB) (string, error) {
		s, err := GenInsertMultiple(db, lcg)
		if err != nil {
			return "INSERT INTO sqlite_master DEFAULT VALUES;", err
		}
		return s.SQL(), nil
	}
	m["insert_bulk"] = func(db *sql.DB) (string, error) {
		s, err := GenInsertBulk(db, lcg)
		if err != nil {
			return "INSERT INTO sqlite_master DEFAULT VALUES;", err
		}
		return s.SQL(), nil
	}
	m["insert_or_replace"] = func(db *sql.DB) (string, error) {
		s, err := GenInsertOrReplace(db, lcg)
		if err != nil {
			return "INSERT INTO sqlite_master DEFAULT VALUES;", err
		}
		return s.SQL(), nil
	}
	m["insert_or_ignore"] = func(db *sql.DB) (string, error) {
		s, err := GenInsertOrIgnore(db, lcg)
		if err != nil {
			return "INSERT INTO sqlite_master DEFAULT VALUES;", err
		}
		return s.SQL(), nil
	}
	m["insert_or_abort"] = func(db *sql.DB) (string, error) {
		s, err := GenInsertOrAbort(db, lcg)
		if err != nil {
			return "INSERT INTO sqlite_master DEFAULT VALUES;", err
		}
		return s.SQL(), nil
	}
	m["insert_or_rollback"] = func(db *sql.DB) (string, error) {
		s, err := GenInsertOrRollback(db, lcg)
		if err != nil {
			return "INSERT INTO sqlite_master DEFAULT VALUES;", err
		}
		return s.SQL(), nil
	}
	m["insert_or_fail"] = func(db *sql.DB) (string, error) {
		s, err := GenInsertOrFail(db, lcg)
		if err != nil {
			return "INSERT INTO sqlite_master DEFAULT VALUES;", err
		}
		return s.SQL(), nil
	}

	m["update"] = func(db *sql.DB) (string, error) {
		s, err := GenUpdate(db, lcg)
		if err != nil {
			return "UPDATE sqlite_master SET name = 'fallback';", err
		}
		return s.SQL(), nil
	}
	m["delete"] = func(db *sql.DB) (string, error) {
		s, err := GenDelete(db, lcg)
		if err != nil {
			return "DELETE FROM sqlite_master WHERE 0;", err
		}
		return s.SQL(), nil
	}

	m["select_basic"] = func(db *sql.DB) (string, error) {
		s, err := GenSelect(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_where"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectWhere(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_where_complex"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectWhereComplex(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_where_in"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectWhereIn(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_subquery"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectSubquery(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_case"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectCase(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_aggregate_complex"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectAggregateComplex(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_like"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectWhereLike(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_limit"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectLimit(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_order"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectOrderBy(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_group"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectGroupBy(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_having"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectHaving(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}

	m["select_join"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectJoin(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_crossjoin"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectCrossJoin(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_innerjoin"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectInnerJoin(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_outerjoin"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectOuterJoin(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_joinusing"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectJoinUsing(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_naturaljoin"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectNaturalJoin(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}

	m["select_recursive"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectRecursive(db, lcg, maxRecursionDepth)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_nested_case"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectWithNestedCase(db, lcg, maxRecursionDepth)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_complex_join"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectWithComplexJoin(db, lcg, maxRecursionDepth)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}

	m["select_window"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectWithWindowFunction(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_multiple_windows"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectWithMultipleWindows(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_cte"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectWithCTE(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_multiple_cte"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectWithMultipleCTE(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_recursive_cte"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectWithRecursiveCTE(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}

	m["select_uuid"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectWithUUIDFunction(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_regexp"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectWithRegexpFunction(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_vector"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectWithVectorFunction(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}
	m["select_time"] = func(db *sql.DB) (string, error) {
		s, err := GenSelectWithTimeFunction(db, lcg)
		if err != nil {
			return "SELECT 1", err
		}
		return s.SQL(), nil
	}

	m["create_table"] = func(db *sql.DB) (string, error) {
		s, err := GenCreateTable(lcg)
		if err != nil {
			return "CREATE TABLE IF NOT EXISTS fallback (id INTEGER);", err
		}
		return s.SQL(), nil
	}
	m["drop_table"] = func(db *sql.DB) (string, error) {
		s, err := GenDropTable(lcg)
		if err != nil {
			return "DROP TABLE IF EXISTS fallback;", err
		}
		return s.SQL(), nil
	}
	m["alter_table"] = func(db *sql.DB) (string, error) {
		s, err := GenAlterTable(lcg)
		if err != nil {
			return "ALTER TABLE fallback RENAME TO fallback2;", err
		}
		return s.SQL(), nil
	}

	return m
}
