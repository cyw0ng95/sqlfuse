package stmts

import (
	"database/sql"
	"sqlsmith-go/internal/common"
)

// StmtGeneratorFactory creates statement generators based on statement type.
// This implements the Factory pattern to centralize generator creation.
type StmtGeneratorFactory struct {
	flavorConfig FlavorConfig
	lcg          *common.LCG
	maxDepth     int
}

// NewStmtGeneratorFactory creates a new factory for statement generators.
func NewStmtGeneratorFactory(lcg *common.LCG, maxDepth int, flavorConfig FlavorConfig) *StmtGeneratorFactory {
	if flavorConfig == nil {
		flavorConfig = GetDefaultFlavor()
	}
	if lcg == nil {
		lcg = common.NewLCG(1)
	}
	return &StmtGeneratorFactory{
		flavorConfig: flavorConfig,
		lcg:          lcg,
		maxDepth:     maxDepth,
	}
}

// CreateGenerator creates a StmtGenerator for the given statement type.
// Returns nil if the statement type is not supported.
func (f *StmtGeneratorFactory) CreateGenerator(stmtType StmtType) StmtGenerator {
	switch stmtType {
	case StmtPragma:
		return &PragmaGenerator{}
	case StmtInsert, StmtInsertMultiple, StmtInsertBulk,
		StmtInsertOrReplace, StmtInsertOrIgnore, StmtInsertOrAbort,
		StmtInsertOrRollback, StmtInsertOrFail:
		return &InsertVariantGenerator{variant: stmtType}
	case StmtUpdate:
		return &UpdateGenerator{}
	case StmtDelete:
		return &DeleteGenerator{}
	case StmtSelectBasic, StmtSelectWhere, StmtSelectWhereComplex, StmtSelectWhereIn,
		StmtSelectSubquery, StmtSelectCase, StmtSelectAggregateComplex, StmtSelectLike,
		StmtSelectLimit, StmtSelectOrder, StmtSelectGroup, StmtSelectHaving,
		StmtSelectJoin, StmtSelectCross, StmtSelectInner, StmtSelectOuter,
		StmtSelectJoinUsing, StmtSelectNatural, StmtSelectRecursive,
		StmtSelectNestedCase, StmtSelectComplexJoin, StmtSelectWindow,
		StmtSelectMultipleWindows, StmtSelectCTE, StmtSelectMultipleCTE,
		StmtSelectRecursiveCTE, StmtSelectJSON, StmtSelectUUID,
		StmtSelectRegexp, StmtSelectVector, StmtSelectTime:
		return &SelectVariantGenerator{variant: stmtType, maxDepth: f.maxDepth}
	case StmtCreateTable:
		return &CreateTableGenerator{}
	case StmtDropTable:
		return &DropTableGenerator{}
	case StmtAlterTable:
		return &AlterTableGenerator{}
	default:
		return nil
	}
}

// CreateContext creates a GenContext for use with generators.
func (f *StmtGeneratorFactory) CreateContext(db *sql.DB) *GenContext {
	return NewGenContextWithFlavor(db, f.lcg, f.maxDepth, f.flavorConfig)
}

// GenerateStmt is a convenience method that creates a generator and generates a statement.
func (f *StmtGeneratorFactory) GenerateStmt(db *sql.DB, stmtType StmtType) (Stmt, error) {
	gen := f.CreateGenerator(stmtType)
	if gen == nil {
		return nil, &UnsupportedStmtTypeError{StmtType: stmtType}
	}
	
	ctx := f.CreateContext(db)
	if !gen.CanGenerate(ctx) {
		return nil, &CannotGenerateError{StmtType: stmtType, Reason: "generator conditions not met"}
	}
	
	return gen.Generate(ctx)
}

// UnsupportedStmtTypeError is returned when a statement type is not supported.
type UnsupportedStmtTypeError struct {
	StmtType StmtType
}

func (e *UnsupportedStmtTypeError) Error() string {
	return "unsupported statement type: " + string(e.StmtType)
}

// CannotGenerateError is returned when a generator cannot generate a statement.
type CannotGenerateError struct {
	StmtType StmtType
	Reason   string
}

func (e *CannotGenerateError) Error() string {
	return "cannot generate " + string(e.StmtType) + ": " + e.Reason
}
