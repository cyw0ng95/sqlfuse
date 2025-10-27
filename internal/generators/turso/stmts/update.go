package stmts

import (
	"fmt"
	"math"
	"sqlsmith-go/internal/common"
)

// UpdateGenerator is a StmtGenerator for UPDATE statements.
type UpdateGenerator struct{}

// Generate implements StmtGenerator for UPDATE statements.
func (g *UpdateGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return genUpdateInternal(ctx.LCG)
}

// CanGenerate implements StmtGenerator. UPDATE can always be generated (creates synthetic tables).
func (g *UpdateGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// UpdateStmt represents an UPDATE statement.
type UpdateStmt struct {
	sql string
}

func (s *UpdateStmt) SQL() string  { return s.sql }
func (s *UpdateStmt) Type() string { return "update" }

// GenUpdate generates a simple UPDATE statement targeting a pseudo-random table
// and random column names with simple literal values.
// This function is kept for backward compatibility with existing code.
func GenUpdate(lcg *common.LCG) (Stmt, error) {
	return genUpdateInternal(lcg)
}

// genUpdateInternal is the internal implementation used by both old and new interfaces.
func genUpdateInternal(lcg *common.LCG) (Stmt, error) {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}

	tbl := fmt.Sprintf("tbl_%d", lcg.Uint64()%1000000)
	// choose 1..3 set assignments
	n := 1 + lcg.Intn(3)
	sets := make([]string, 0, n)
	for i := 0; i < n; i++ {
		col := fmt.Sprintf("col%d", 1+lcg.Intn(6))
		v := genLiteral(lcg)
		sets = append(sets, fmt.Sprintf("\"%s\" = %s", col, v))
	}

	// optionally add a WHERE clause about half the time
	where := ""
	if lcg.Intn(2) == 0 {
		col := fmt.Sprintf("col%d", 1+lcg.Intn(6))
		where = fmt.Sprintf(" WHERE \"%s\" = %s", col, genLiteral(lcg))
	}

	sql := fmt.Sprintf("UPDATE \"%s\" SET %s%s;", tbl, join(sets, ", "), where)
	return &UpdateStmt{sql: sql}, nil
}

func genLiteral(lcg *common.LCG) string {
	switch lcg.Intn(4) {
	case 0:
		return fmt.Sprintf("%d", int64(lcg.Intn(1000)))
	case 1:
		return fmt.Sprintf("'%s'", fmt.Sprintf("v%d", lcg.Intn(1000)))
	case 2:
		f := math.Round(float64(lcg.Intn(10000))) / 100.0
		return fmt.Sprintf("%f", f)
	default:
		return "NULL"
	}
}

func join(a []string, sep string) string {
	if len(a) == 0 {
		return ""
	}
	res := a[0]
	for i := 1; i < len(a); i++ {
		res += sep + a[i]
	}
	return res
}
