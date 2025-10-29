package stmts

// StmtType represents a generation direction / statement category.
// Defined here so generators across packages can share the same enum.
type StmtType string

const (
	StmtPragma                 StmtType = "pragma"
	StmtInsert                 StmtType = "insert"
	StmtInsertMultiple         StmtType = "insert_multiple"
	StmtInsertBulk             StmtType = "insert_bulk"
	StmtInsertOrReplace        StmtType = "insert_or_replace"
	StmtInsertOrIgnore         StmtType = "insert_or_ignore"
	StmtInsertOrAbort          StmtType = "insert_or_abort"
	StmtInsertOrRollback       StmtType = "insert_or_rollback"
	StmtInsertOrFail           StmtType = "insert_or_fail"
	StmtUpdate                 StmtType = "update"
	StmtDelete                 StmtType = "delete"
	StmtSelectBasic            StmtType = "select_basic"
	StmtSelectWhere            StmtType = "select_where"
	StmtSelectWhereComplex     StmtType = "select_where_complex"
	StmtSelectWhereIn          StmtType = "select_where_in"
	StmtSelectSubquery         StmtType = "select_subquery"
	StmtSelectCase             StmtType = "select_case"
	StmtSelectAggregateComplex StmtType = "select_aggregate_complex"
	StmtSelectLike             StmtType = "select_like"
	StmtSelectLimit            StmtType = "select_limit"
	StmtSelectOrder            StmtType = "select_order"
	StmtSelectGroup            StmtType = "select_group"
	StmtSelectHaving           StmtType = "select_having"
	StmtSelectJoin             StmtType = "select_join"
	StmtSelectCross            StmtType = "select_crossjoin"
	StmtSelectInner            StmtType = "select_innerjoin"
	StmtSelectOuter            StmtType = "select_outerjoin"
	StmtSelectJoinUsing        StmtType = "select_joinusing"
	StmtSelectNatural          StmtType = "select_naturaljoin"
	StmtSelectRecursive        StmtType = "select_recursive"
	StmtSelectNestedCase       StmtType = "select_nested_case"
	StmtSelectComplexJoin      StmtType = "select_complex_join"
	StmtSelectWindow           StmtType = "select_window"
	StmtSelectMultipleWindows  StmtType = "select_multiple_windows"
	StmtSelectCTE              StmtType = "select_cte"
	StmtSelectMultipleCTE      StmtType = "select_multiple_cte"
	StmtSelectRecursiveCTE     StmtType = "select_recursive_cte"
	StmtSelectJSON             StmtType = "select_json"
	StmtSelectUUID             StmtType = "select_uuid"
	StmtSelectRegexp           StmtType = "select_regexp"
	StmtSelectVector           StmtType = "select_vector"
	StmtSelectTime             StmtType = "select_time"
	StmtCreateTable            StmtType = "create_table"
	StmtDropTable              StmtType = "drop_table"
	StmtAlterTable             StmtType = "alter_table"
	StmtCreateView             StmtType = "create_view"
	StmtDropView               StmtType = "drop_view"
	StmtCreateIndex            StmtType = "create_index"
	StmtDropIndex              StmtType = "drop_index"
	StmtCreateVirtualTable     StmtType = "create_virtual_table"
	StmtAttach                 StmtType = "attach"
	StmtDetach                 StmtType = "detach"
	StmtBegin                  StmtType = "begin"
	StmtCommit                 StmtType = "commit"
	StmtRollback               StmtType = "rollback"
	StmtExplain                StmtType = "explain"
	StmtExplainQueryPlan       StmtType = "explain_query_plan"
	StmtAnalyze                StmtType = "analyze"
	StmtVacuum                 StmtType = "vacuum"
	StmtReindex                StmtType = "reindex"
	StmtSavepoint              StmtType = "savepoint"
	StmtRelease                StmtType = "release"
	StmtCreateTrigger          StmtType = "create_trigger"
	StmtDropTrigger            StmtType = "drop_trigger"
	StmtSelectUnion            StmtType = "select_union"
	StmtSelectIntersect        StmtType = "select_intersect"
	StmtSelectExcept           StmtType = "select_except"
)

// AllStmtTypes defines a deterministic ordering used when selecting by weights.
var AllStmtTypes = []StmtType{
	StmtInsert,
	StmtInsertMultiple,
	StmtInsertBulk,
	StmtInsertOrReplace,
	StmtInsertOrIgnore,
	StmtInsertOrAbort,
	StmtInsertOrRollback,
	StmtInsertOrFail,
	StmtUpdate,
	StmtDelete,
	StmtSelectBasic,
	StmtSelectWhere,
	StmtSelectWhereComplex,
	StmtSelectWhereIn,
	StmtSelectSubquery,
	StmtSelectCase,
	StmtSelectAggregateComplex,
	StmtSelectLike,
	StmtSelectLimit,
	StmtSelectOrder,
	StmtSelectGroup,
	StmtSelectHaving,
	StmtSelectJoin,
	StmtSelectCross,
	StmtSelectInner,
	StmtSelectOuter,
	StmtSelectJoinUsing,
	StmtSelectNatural,
	StmtSelectRecursive,
	StmtSelectNestedCase,
	StmtSelectComplexJoin,
	StmtSelectWindow,
	StmtSelectMultipleWindows,
	StmtSelectCTE,
	StmtSelectMultipleCTE,
	StmtSelectRecursiveCTE,
	StmtSelectJSON,
	StmtSelectUUID,
	StmtSelectRegexp,
	StmtSelectVector,
	StmtSelectTime,
	StmtCreateTable,
	StmtDropTable,
	StmtAlterTable,
	StmtCreateView,
	StmtDropView,
	StmtCreateIndex,
	StmtDropIndex,
	StmtCreateVirtualTable,
	StmtAttach,
	StmtDetach,
	StmtBegin,
	StmtCommit,
	StmtRollback,
	StmtExplain,
	StmtExplainQueryPlan,
	StmtAnalyze,
	StmtVacuum,
	StmtReindex,
	StmtSavepoint,
	StmtRelease,
	StmtCreateTrigger,
	StmtDropTrigger,
	StmtSelectUnion,
	StmtSelectIntersect,
	StmtSelectExcept,
}
