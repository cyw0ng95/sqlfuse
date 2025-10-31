package dml

import (
	"sqlfuse/internal/stmts/stmts"
)

// SelectVariantGenerator generates different variants of SELECT statements.
// This implements the Strategy pattern to handle different SELECT types.
type SelectVariantGenerator struct {
	variant  stmts.StmtType
	maxDepth int
}

// Generate implements stmts.StmtGenerator for SELECT statement variants.
func (g *SelectVariantGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	db := ctx.DB
	lcg := ctx.LCG
	flavor := ctx.Flavor

	// Route to appropriate generator based on variant
	// Return as pointer to implement stmts.Stmt interface properly
	switch g.variant {
	case stmts.StmtSelectBasic:
		s, err := GenSelectWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectWhere:
		s, err := GenSelectWhereWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectWhereComplex:
		s, err := GenSelectWhereComplexWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectWhereIn:
		s, err := GenSelectWhereInWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectSubquery:
		s, err := GenSelectSubqueryWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectCase:
		s, err := GenSelectCaseWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectAggregateComplex:
		s, err := GenSelectAggregateComplexWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectLike:
		s, err := GenSelectWhereLikeWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectLimit:
		s, err := GenSelectLimitWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectOrder:
		s, err := GenSelectOrderByWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectGroup:
		s, err := GenSelectGroupByWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectHaving:
		s, err := GenSelectHavingWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectJoin:
		s, err := GenSelectJoinWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectCross:
		s, err := GenSelectCrossJoinWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectInner:
		s, err := GenSelectInnerJoinWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectOuter:
		s, err := GenSelectOuterJoinWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectJoinUsing:
		s, err := GenSelectJoinUsingWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectNatural:
		s, err := GenSelectNaturalJoinWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectRecursive:
		s, err := GenSelectRecursiveWithFlavor(db, lcg, g.maxDepth, flavor)
		return &s, err
	case stmts.StmtSelectNestedCase:
		s, err := GenSelectWithNestedCaseWithFlavor(db, lcg, g.maxDepth, flavor)
		return &s, err
	case stmts.StmtSelectComplexJoin:
		s, err := GenSelectWithComplexJoinWithFlavor(db, lcg, g.maxDepth, flavor)
		return &s, err
	case stmts.StmtSelectDeeplyNested:
		s, err := GenSelectDeeplyNestedWithFlavor(db, lcg, g.maxDepth, flavor)
		return &s, err
	case stmts.StmtSelectWindow:
		s, err := GenSelectWithWindowFunctionWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectMultipleWindows:
		s, err := GenSelectWithMultipleWindowsWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectCTE:
		s, err := GenSelectWithCTEWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectMultipleCTE:
		s, err := GenSelectWithMultipleCTEWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectRecursiveCTE:
		s, err := GenSelectWithRecursiveCTEWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectJSON:
		s, err := GenSelectWithJSONFunctionWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectUUID:
		s, err := GenSelectWithUUIDFunctionWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectRegexp:
		s, err := GenSelectWithRegexpFunctionWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectVector:
		s, err := GenSelectWithVectorFunctionWithFlavor(db, lcg, flavor)
		return &s, err
	case stmts.StmtSelectTime:
		s, err := GenSelectWithTimeFunctionWithFlavor(db, lcg, flavor)
		return &s, err
	default:
		s, err := GenSelectWithFlavor(db, lcg, flavor)
		return &s, err
	}
}

// CanGenerate implements stmts.StmtGenerator. SELECT can always be generated (creates fallback SELECT 1).
func (g *SelectVariantGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	return true
}
