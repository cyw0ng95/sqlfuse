package dml

import (
	"sqlfuse/internal/stmts/stmts"
	"database/sql"
	"sqlfuse/internal/common"
)

// This file contains temporary stub implementations of flavor-aware Gen* functions.
// These will be replaced with proper implementations that use the flavor parameter.

// GenSelectWithFlavor is a flavor-aware version of GenSelect
func GenSelectWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelect(db, lcg)
}

// GenSelectWhereWithFlavor is a flavor-aware version of GenSelectWhere
func GenSelectWhereWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectWhere(db, lcg)
}

// GenSelectWhereComplexWithFlavor is a flavor-aware version of GenSelectWhereComplex
func GenSelectWhereComplexWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectWhereComplex(db, lcg)
}

// GenSelectWhereInWithFlavor is a flavor-aware version of GenSelectWhereIn
func GenSelectWhereInWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectWhereIn(db, lcg)
}

// GenSelectSubqueryWithFlavor is a flavor-aware version of GenSelectSubquery
func GenSelectSubqueryWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectSubquery(db, lcg)
}

// GenSelectCaseWithFlavor is a flavor-aware version of GenSelectCase
func GenSelectCaseWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectCase(db, lcg)
}

// GenSelectAggregateComplexWithFlavor is a flavor-aware version of GenSelectAggregateComplex
func GenSelectAggregateComplexWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectAggregateComplex(db, lcg)
}

// GenSelectWhereLikeWithFlavor is a flavor-aware version of GenSelectWhereLike
func GenSelectWhereLikeWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectWhereLike(db, lcg)
}

// GenSelectLimitWithFlavor is a flavor-aware version of GenSelectLimit
func GenSelectLimitWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectLimit(db, lcg)
}

// GenSelectOrderByWithFlavor is a flavor-aware version of GenSelectOrderBy
func GenSelectOrderByWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectOrderBy(db, lcg)
}

// GenSelectGroupByWithFlavor is a flavor-aware version of GenSelectGroupBy
func GenSelectGroupByWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectGroupBy(db, lcg)
}

// GenSelectHavingWithFlavor is a flavor-aware version of GenSelectHaving
func GenSelectHavingWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectHaving(db, lcg)
}

// GenSelectJoinWithFlavor is a flavor-aware version of GenSelectJoin
func GenSelectJoinWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectJoin(db, lcg)
}

// GenSelectCrossJoinWithFlavor is a flavor-aware version of GenSelectCrossJoin
func GenSelectCrossJoinWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectCrossJoin(db, lcg)
}

// GenSelectInnerJoinWithFlavor is a flavor-aware version of GenSelectInnerJoin
func GenSelectInnerJoinWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectInnerJoin(db, lcg)
}

// GenSelectOuterJoinWithFlavor is a flavor-aware version of GenSelectOuterJoin
func GenSelectOuterJoinWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectOuterJoin(db, lcg)
}

// GenSelectJoinUsingWithFlavor is a flavor-aware version of GenSelectJoinUsing
func GenSelectJoinUsingWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectJoinUsing(db, lcg)
}

// GenSelectNaturalJoinWithFlavor is a flavor-aware version of GenSelectNaturalJoin
func GenSelectNaturalJoinWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectNaturalJoin(db, lcg)
}

// GenSelectRecursiveWithFlavor is a flavor-aware version of GenSelectRecursive
func GenSelectRecursiveWithFlavor(db *sql.DB, lcg *common.LCG, maxDepth int, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectRecursive(db, lcg, maxDepth)
}

// GenSelectWithNestedCaseWithFlavor is a flavor-aware version of GenSelectWithNestedCase
func GenSelectWithNestedCaseWithFlavor(db *sql.DB, lcg *common.LCG, maxDepth int, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectWithNestedCase(db, lcg, maxDepth)
}

// GenSelectWithComplexJoinWithFlavor is a flavor-aware version of GenSelectWithComplexJoin
func GenSelectWithComplexJoinWithFlavor(db *sql.DB, lcg *common.LCG, maxDepth int, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectWithComplexJoin(db, lcg, maxDepth)
}

// GenSelectDeeplyNestedWithFlavor is a flavor-aware version of GenSelectDeeplyNested
func GenSelectDeeplyNestedWithFlavor(db *sql.DB, lcg *common.LCG, maxDepth int, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectDeeplyNested(db, lcg, maxDepth)
}

// GenSelectWithWindowFunctionWithFlavor is a flavor-aware version of GenSelectWithWindowFunction
func GenSelectWithWindowFunctionWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectWithWindowFunction(db, lcg)
}

// GenSelectWithMultipleWindowsWithFlavor is a flavor-aware version of GenSelectWithMultipleWindows
func GenSelectWithMultipleWindowsWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectWithMultipleWindows(db, lcg)
}

// GenSelectWithCTEWithFlavor is a flavor-aware version of GenSelectWithCTE
func GenSelectWithCTEWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectWithCTE(db, lcg)
}

// GenSelectWithMultipleCTEWithFlavor is a flavor-aware version of GenSelectWithMultipleCTE
func GenSelectWithMultipleCTEWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectWithMultipleCTE(db, lcg)
}

// GenSelectWithRecursiveCTEWithFlavor is a flavor-aware version of GenSelectWithRecursiveCTE
func GenSelectWithRecursiveCTEWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectWithRecursiveCTE(db, lcg)
}

// GenSelectWithJSONFunctionWithFlavor is a flavor-aware version of GenSelectWithJSONFunction
func GenSelectWithJSONFunctionWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectWithJSONFunction(db, lcg)
}

// GenSelectWithUUIDFunctionWithFlavor is a flavor-aware version of GenSelectWithUUIDFunction
func GenSelectWithUUIDFunctionWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectWithUUIDFunction(db, lcg)
}

// GenSelectWithRegexpFunctionWithFlavor is a flavor-aware version of GenSelectWithRegexpFunction
func GenSelectWithRegexpFunctionWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectWithRegexpFunction(db, lcg)
}

// GenSelectWithVectorFunctionWithFlavor is a flavor-aware version of GenSelectWithVectorFunction
func GenSelectWithVectorFunctionWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectWithVectorFunction(db, lcg)
}

// GenSelectWithTimeFunctionWithFlavor is a flavor-aware version of GenSelectWithTimeFunction
func GenSelectWithTimeFunctionWithFlavor(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (SelectStmt, error) {
	return GenSelectWithTimeFunction(db, lcg)
}
