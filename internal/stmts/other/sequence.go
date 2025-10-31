package other

import (
	"sqlfuse/internal/stmts/stmts"
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
)

// CreateSequenceGenerator is a stmts.StmtGenerator for CREATE SEQUENCE statements.
type CreateSequenceGenerator struct{}

// Generate implements stmts.StmtGenerator for CREATE SEQUENCE statements.
func (g *CreateSequenceGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	return GenCreateSequence(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements stmts.StmtGenerator. CREATE SEQUENCE can always be generated.
func (g *CreateSequenceGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	return true
}

// CreateSequenceStmt represents a CREATE SEQUENCE statement.
type CreateSequenceStmt struct {
	*stmts.BaseStmt
}

// GenCreateSequence generates a CREATE SEQUENCE statement.
// Reference: https://duckdb.org/docs/stable/sql/statements/create_sequence
func GenCreateSequence(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (stmts.Stmt, error) {
	lcg = stmts.EnsureLCG(lcg)
	if flavor == nil {
		flavor = stmts.GetDefaultFlavor()
	}

	seqName := fmt.Sprintf("seq_%d", lcg.Uint64()%10000)

	var sql string
	choice := lcg.Intn(5)

	if choice == 0 {
		// Basic CREATE SEQUENCE
		sql = fmt.Sprintf("CREATE SEQUENCE IF NOT EXISTS \"%s\";", seqName)
	} else if choice == 1 {
		// With START value
		start := lcg.Intn(1000)
		sql = fmt.Sprintf("CREATE SEQUENCE IF NOT EXISTS \"%s\" START %d;", seqName, start)
	} else if choice == 2 {
		// With INCREMENT
		increment := 1 + lcg.Intn(10)
		sql = fmt.Sprintf("CREATE SEQUENCE IF NOT EXISTS \"%s\" INCREMENT %d;", seqName, increment)
	} else if choice == 3 {
		// With MIN and MAX
		minVal := lcg.Intn(100)
		maxVal := minVal + 1000 + lcg.Intn(9000)
		sql = fmt.Sprintf("CREATE SEQUENCE IF NOT EXISTS \"%s\" MINVALUE %d MAXVALUE %d;", seqName, minVal, maxVal)
	} else {
		// Full specification
		start := lcg.Intn(1000)
		increment := 1 + lcg.Intn(10)
		sql = fmt.Sprintf("CREATE SEQUENCE IF NOT EXISTS \"%s\" START %d INCREMENT %d;", seqName, start, increment)
	}

	return &CreateSequenceStmt{
		BaseStmt: stmts.NewBaseStmt(sql, "create_sequence", flavor),
	}, nil
}

// DropSequenceGenerator is a stmts.StmtGenerator for DROP SEQUENCE statements.
type DropSequenceGenerator struct{}

// Generate implements stmts.StmtGenerator for DROP SEQUENCE statements.
func (g *DropSequenceGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	return GenDropSequence(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements stmts.StmtGenerator. DROP SEQUENCE can always be generated.
func (g *DropSequenceGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	return true
}

// DropSequenceStmt represents a DROP SEQUENCE statement.
type DropSequenceStmt struct {
	*stmts.BaseStmt
}

// GenDropSequence generates a DROP SEQUENCE statement.
// Reference: https://duckdb.org/docs/stable/sql/statements/drop
func GenDropSequence(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (stmts.Stmt, error) {
	lcg = stmts.EnsureLCG(lcg)
	if flavor == nil {
		flavor = stmts.GetDefaultFlavor()
	}

	seqName := fmt.Sprintf("seq_%d", lcg.Uint64()%10000)
	sql := fmt.Sprintf("DROP SEQUENCE IF EXISTS \"%s\";", seqName)

	return &DropSequenceStmt{
		BaseStmt: stmts.NewBaseStmt(sql, "drop_sequence", flavor),
	}, nil
}
