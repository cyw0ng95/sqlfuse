package stmts

// SelectVariantGenerator generates different variants of SELECT statements.
// This implements the Strategy pattern to handle different SELECT types.
type SelectVariantGenerator struct {
	variant  StmtType
	maxDepth int
}

// Generate implements StmtGenerator for SELECT statement variants.
func (g *SelectVariantGenerator) Generate(ctx *GenContext) (Stmt, error) {
	db := ctx.DB
	lcg := ctx.LCG
	flavor := ctx.Flavor

	// Route to appropriate generator based on variant
	// Return as pointer to implement Stmt interface properly
	switch g.variant {
	case StmtSelectBasic:
		s, err := GenSelectWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectWhere:
		s, err := GenSelectWhereWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectWhereComplex:
		s, err := GenSelectWhereComplexWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectWhereIn:
		s, err := GenSelectWhereInWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectSubquery:
		s, err := GenSelectSubqueryWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectCase:
		s, err := GenSelectCaseWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectAggregateComplex:
		s, err := GenSelectAggregateComplexWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectLike:
		s, err := GenSelectWhereLikeWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectLimit:
		s, err := GenSelectLimitWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectOrder:
		s, err := GenSelectOrderByWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectGroup:
		s, err := GenSelectGroupByWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectHaving:
		s, err := GenSelectHavingWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectJoin:
		s, err := GenSelectJoinWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectCross:
		s, err := GenSelectCrossJoinWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectInner:
		s, err := GenSelectInnerJoinWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectOuter:
		s, err := GenSelectOuterJoinWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectJoinUsing:
		s, err := GenSelectJoinUsingWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectNatural:
		s, err := GenSelectNaturalJoinWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectRecursive:
		s, err := GenSelectRecursiveWithFlavor(db, lcg, g.maxDepth, flavor)
		return &s, err
	case StmtSelectNestedCase:
		s, err := GenSelectWithNestedCaseWithFlavor(db, lcg, g.maxDepth, flavor)
		return &s, err
	case StmtSelectComplexJoin:
		s, err := GenSelectWithComplexJoinWithFlavor(db, lcg, g.maxDepth, flavor)
		return &s, err
	case StmtSelectDeeplyNested:
		s, err := GenSelectDeeplyNestedWithFlavor(db, lcg, g.maxDepth, flavor)
		return &s, err
	case StmtSelectWindow:
		s, err := GenSelectWithWindowFunctionWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectMultipleWindows:
		s, err := GenSelectWithMultipleWindowsWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectCTE:
		s, err := GenSelectWithCTEWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectMultipleCTE:
		s, err := GenSelectWithMultipleCTEWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectRecursiveCTE:
		s, err := GenSelectWithRecursiveCTEWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectJSON:
		s, err := GenSelectWithJSONFunctionWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectUUID:
		s, err := GenSelectWithUUIDFunctionWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectRegexp:
		s, err := GenSelectWithRegexpFunctionWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectVector:
		s, err := GenSelectWithVectorFunctionWithFlavor(db, lcg, flavor)
		return &s, err
	case StmtSelectTime:
		s, err := GenSelectWithTimeFunctionWithFlavor(db, lcg, flavor)
		return &s, err
	default:
		s, err := GenSelectWithFlavor(db, lcg, flavor)
		return &s, err
	}
}

// CanGenerate implements StmtGenerator. SELECT can always be generated (creates fallback SELECT 1).
func (g *SelectVariantGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}
