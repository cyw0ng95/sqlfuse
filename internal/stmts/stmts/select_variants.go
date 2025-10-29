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
	
	// Route to appropriate generator based on variant
	// Return as pointer to implement Stmt interface properly
	switch g.variant {
	case StmtSelectBasic:
		s, err := GenSelect(db, lcg)
		return &s, err
	case StmtSelectWhere:
		s, err := GenSelectWhere(db, lcg)
		return &s, err
	case StmtSelectWhereComplex:
		s, err := GenSelectWhereComplex(db, lcg)
		return &s, err
	case StmtSelectWhereIn:
		s, err := GenSelectWhereIn(db, lcg)
		return &s, err
	case StmtSelectSubquery:
		s, err := GenSelectSubquery(db, lcg)
		return &s, err
	case StmtSelectCase:
		s, err := GenSelectCase(db, lcg)
		return &s, err
	case StmtSelectAggregateComplex:
		s, err := GenSelectAggregateComplex(db, lcg)
		return &s, err
	case StmtSelectLike:
		s, err := GenSelectWhereLike(db, lcg)
		return &s, err
	case StmtSelectLimit:
		s, err := GenSelectLimit(db, lcg)
		return &s, err
	case StmtSelectOrder:
		s, err := GenSelectOrderBy(db, lcg)
		return &s, err
	case StmtSelectGroup:
		s, err := GenSelectGroupBy(db, lcg)
		return &s, err
	case StmtSelectHaving:
		s, err := GenSelectHaving(db, lcg)
		return &s, err
	case StmtSelectJoin:
		s, err := GenSelectJoin(db, lcg)
		return &s, err
	case StmtSelectCross:
		s, err := GenSelectCrossJoin(db, lcg)
		return &s, err
	case StmtSelectInner:
		s, err := GenSelectInnerJoin(db, lcg)
		return &s, err
	case StmtSelectOuter:
		s, err := GenSelectOuterJoin(db, lcg)
		return &s, err
	case StmtSelectJoinUsing:
		s, err := GenSelectJoinUsing(db, lcg)
		return &s, err
	case StmtSelectNatural:
		s, err := GenSelectNaturalJoin(db, lcg)
		return &s, err
	case StmtSelectRecursive:
		s, err := GenSelectRecursive(db, lcg, g.maxDepth)
		return &s, err
	case StmtSelectNestedCase:
		s, err := GenSelectWithNestedCase(db, lcg, g.maxDepth)
		return &s, err
	case StmtSelectComplexJoin:
		s, err := GenSelectWithComplexJoin(db, lcg, g.maxDepth)
		return &s, err
	case StmtSelectWindow:
		s, err := GenSelectWithWindowFunction(db, lcg)
		return &s, err
	case StmtSelectMultipleWindows:
		s, err := GenSelectWithMultipleWindows(db, lcg)
		return &s, err
	case StmtSelectCTE:
		s, err := GenSelectWithCTE(db, lcg)
		return &s, err
	case StmtSelectMultipleCTE:
		s, err := GenSelectWithMultipleCTE(db, lcg)
		return &s, err
	case StmtSelectRecursiveCTE:
		s, err := GenSelectWithRecursiveCTE(db, lcg)
		return &s, err
	case StmtSelectJSON:
		s, err := GenSelectWithJSONFunction(db, lcg)
		return &s, err
	case StmtSelectUUID:
		s, err := GenSelectWithUUIDFunction(db, lcg)
		return &s, err
	case StmtSelectRegexp:
		s, err := GenSelectWithRegexpFunction(db, lcg)
		return &s, err
	case StmtSelectVector:
		s, err := GenSelectWithVectorFunction(db, lcg)
		return &s, err
	case StmtSelectTime:
		s, err := GenSelectWithTimeFunction(db, lcg)
		return &s, err
	default:
		s, err := GenSelect(db, lcg)
		return &s, err
	}
}

// CanGenerate implements StmtGenerator. SELECT can always be generated (creates fallback SELECT 1).
func (g *SelectVariantGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}
