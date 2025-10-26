package turso

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/generators/turso/stmts"
)

// Generator uses an LCG to drive generation directions and produce SQL snippets.
type Generator struct {
	lcg               *common.LCG
	first             bool // first generation is forced into pragma
	weights           map[StmtType]uint64
	totalWeight       uint64
	maxRecursionDepth int // Maximum depth for recursive generation (default: 2)
}

// StmtType represents a generation direction / statement category.
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
	StmtCreateTable            StmtType = "create_table"
	StmtDropTable              StmtType = "drop_table"
	StmtAlterTable             StmtType = "alter_table"
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
	StmtCreateTable,
	StmtDropTable,
	StmtAlterTable,
}

// DefaultStmtWeights returns a sensible default weight distribution.
// Values are token-like weights; probabilities are weight / sum(weights).
func DefaultStmtWeights() map[StmtType]uint64 {
	w := map[StmtType]uint64{}
	// INSERT variants
	w[StmtInsert] = 250        // reduced to make room for new insert types
	w[StmtInsertMultiple] = 70 // multiple row inserts
	w[StmtInsertBulk] = 15     // bulk inserts for heavy testing
	w[StmtInsertOrReplace] = 25
	w[StmtInsertOrIgnore] = 25
	w[StmtInsertOrAbort] = 15
	w[StmtInsertOrRollback] = 10
	w[StmtInsertOrFail] = 10
	// UPDATE and DELETE
	w[StmtUpdate] = 80 // enhanced UPDATE with schema
	w[StmtDelete] = 60 // enhanced DELETE with schema
	// Basic SELECT variants
	w[StmtSelectBasic] = 120
	w[StmtSelectWhere] = 100
	w[StmtSelectWhereComplex] = 60
	w[StmtSelectWhereIn] = 50
	w[StmtSelectSubquery] = 40
	w[StmtSelectCase] = 40
	w[StmtSelectAggregateComplex] = 30
	w[StmtSelectLike] = 80
	w[StmtSelectLimit] = 60
	w[StmtSelectOrder] = 40
	w[StmtSelectGroup] = 40
	w[StmtSelectHaving] = 40
	// JOIN variants
	w[StmtSelectJoin] = 20
	w[StmtSelectCross] = 20
	w[StmtSelectInner] = 20
	w[StmtSelectOuter] = 20
	w[StmtSelectJoinUsing] = 20
	w[StmtSelectNatural] = 20
	// Advanced SELECT with recursion/nesting
	w[StmtSelectRecursive] = 30
	w[StmtSelectNestedCase] = 25
	w[StmtSelectComplexJoin] = 25
	// Window functions and CTEs (new)
	w[StmtSelectWindow] = 35
	w[StmtSelectMultipleWindows] = 20
	w[StmtSelectCTE] = 30
	w[StmtSelectMultipleCTE] = 20
	w[StmtSelectRecursiveCTE] = 15
	// DDL (keep low by default)
	w[StmtCreateTable] = 40
	w[StmtDropTable] = 40
	w[StmtAlterTable] = 40
	return w
}

// NewGenerator creates a generator seeded with the provided seed and default weights.
func NewGenerator(seed uint64) *Generator {
	g := &Generator{
		lcg:               common.NewLCG(seed),
		first:             true,
		maxRecursionDepth: 2, // Default recursion depth
	}
	g.SetWeights(DefaultStmtWeights())
	return g
}

// SetWeights replaces the current weights and recalculates totals.
func (g *Generator) SetWeights(weights map[StmtType]uint64) {
	if g.weights == nil {
		g.weights = make(map[StmtType]uint64, len(weights))
	}
	for k, v := range weights {
		g.weights[k] = v
	}
	g.recalcTotalWeight()
}

// SetWeight sets a single statement type weight and updates totals.
func (g *Generator) SetWeight(t StmtType, weight uint64) {
	if g.weights == nil {
		g.weights = DefaultStmtWeights()
	}
	g.weights[t] = weight
	g.recalcTotalWeight()
}

// GetWeights returns a copy of the current weights map.
func (g *Generator) GetWeights() map[StmtType]uint64 {
	out := make(map[StmtType]uint64, len(g.weights))
	for k, v := range g.weights {
		out[k] = v
	}
	return out
}

// SetMaxRecursionDepth sets the maximum recursion depth for complex SQL generation.
// A depth of 0 means no recursion (simple queries only).
// A depth of 1 allows one level of nesting (e.g., subquery in WHERE).
// A depth of 2 or more allows deeper nesting.
func (g *Generator) SetMaxRecursionDepth(depth int) {
	if depth < 0 {
		depth = 0
	}
	g.maxRecursionDepth = depth
}

// GetMaxRecursionDepth returns the current maximum recursion depth.
func (g *Generator) GetMaxRecursionDepth() int {
	return g.maxRecursionDepth
}

func (g *Generator) recalcTotalWeight() {
	var sum uint64
	for _, t := range AllStmtTypes {
		sum += g.weights[t]
	}
	g.totalWeight = sum
}

// Direction picks a direction to drive generation.
// On the very first call this is 100% StmtPragma. Afterwards it uses the LCG
// to pick between pragma, ddl, and dml. If weights are set (totalWeight>0)
// selection is proportional to weights.
func (g *Generator) Direction() StmtType {
	if g.first {
		g.first = false
		return StmtPragma
	}
	// If weights are configured, pick proportionally.
	if g.totalWeight > 0 {
		r := g.lcg.Uint64() % g.totalWeight
		var cum uint64
		for _, t := range AllStmtTypes {
			w := g.weights[t]
			cum += w
			if r < cum {
				return t
			}
		}
		// fallback
		return StmtPragma
	} else {
		panic("totalWeight < 0")
	}
}

// GenerateWithDB produces a single SQL statement according to the chosen direction.
// If db is provided, can generate SELECTs using schema.
func (g *Generator) GenerateWithDB(db *sql.DB) string {
	dir := g.Direction()
	switch dir {
	case StmtPragma:
		return stmts.GenPragma(g.lcg).SQL()
	case StmtInsert:
		stmt, err := stmts.GenInsert(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating INSERT:", err)
			return "INSERT INTO sqlite_master DEFAULT VALUES;" // fallback
		}
		return stmt.SQL()
	case StmtInsertMultiple:
		stmt, err := stmts.GenInsertMultiple(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating INSERT MULTIPLE:", err)
			return "INSERT INTO sqlite_master DEFAULT VALUES;" // fallback
		}
		return stmt.SQL()
	case StmtInsertBulk:
		stmt, err := stmts.GenInsertBulk(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating INSERT BULK:", err)
			return "INSERT INTO sqlite_master DEFAULT VALUES;" // fallback
		}
		return stmt.SQL()
	case StmtInsertOrReplace:
		stmt, err := stmts.GenInsertOrReplace(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating INSERT OR REPLACE:", err)
			return "INSERT INTO sqlite_master DEFAULT VALUES;" // fallback
		}
		return stmt.SQL()
	case StmtInsertOrIgnore:
		stmt, err := stmts.GenInsertOrIgnore(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating INSERT OR IGNORE:", err)
			return "INSERT INTO sqlite_master DEFAULT VALUES;" // fallback
		}
		return stmt.SQL()
	case StmtInsertOrAbort:
		stmt, err := stmts.GenInsertOrAbort(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating INSERT OR ABORT:", err)
			return "INSERT INTO sqlite_master DEFAULT VALUES;" // fallback
		}
		return stmt.SQL()
	case StmtInsertOrRollback:
		stmt, err := stmts.GenInsertOrRollback(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating INSERT OR ROLLBACK:", err)
			return "INSERT INTO sqlite_master DEFAULT VALUES;" // fallback
		}
		return stmt.SQL()
	case StmtInsertOrFail:
		stmt, err := stmts.GenInsertOrFail(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating INSERT OR FAIL:", err)
			return "INSERT INTO sqlite_master DEFAULT VALUES;" // fallback
		}
		return stmt.SQL()
	case StmtUpdate:
		stmt, err := stmts.GenUpdate(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating UPDATE:", err)
			return "UPDATE sqlite_master SET name = 'fallback';" // fallback
		}
		return stmt.SQL()
	case StmtDelete:
		stmt, err := stmts.GenDelete(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating DELETE:", err)
			return "DELETE FROM sqlite_master WHERE 0;" // fallback
		}
		return stmt.SQL()
	case StmtSelectBasic:
		stmt, err := stmts.GenSelect(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT:", err)
			return "SELECT 1" // fallback
		}
		return stmt.SQL()
	case StmtSelectWhere:
		stmtW, err := stmts.GenSelectWhere(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT WHERE:", err)
			return "SELECT 1"
		}
		return stmtW.SQL()
	case StmtSelectWhereComplex:
		stmt, err := stmts.GenSelectWhereComplex(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT WHERE COMPLEX:", err)
			return "SELECT 1"
		}
		return stmt.SQL()
	case StmtSelectWhereIn:
		stmt, err := stmts.GenSelectWhereIn(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT WHERE IN:", err)
			return "SELECT 1"
		}
		return stmt.SQL()
	case StmtSelectSubquery:
		stmt, err := stmts.GenSelectSubquery(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT SUBQUERY:", err)
			return "SELECT 1"
		}
		return stmt.SQL()
	case StmtSelectCase:
		stmt, err := stmts.GenSelectCase(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT CASE:", err)
			return "SELECT 1"
		}
		return stmt.SQL()
	case StmtSelectAggregateComplex:
		stmt, err := stmts.GenSelectAggregateComplex(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT AGGREGATE COMPLEX:", err)
			return "SELECT 1"
		}
		return stmt.SQL()
	case StmtSelectLike:
		stmtL, err := stmts.GenSelectWhereLike(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT LIKE:", err)
			return "SELECT 1"
		}
		return stmtL.SQL()
	case StmtSelectLimit:
		stmtLim, err := stmts.GenSelectLimit(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT LIMIT:", err)
			return "SELECT 1"
		}
		return stmtLim.SQL()
	case StmtSelectOrder:
		stmtOrd, err := stmts.GenSelectOrderBy(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT ORDER BY:", err)
			return "SELECT 1"
		}
		return stmtOrd.SQL()
	case StmtSelectGroup:
		stmtG, err := stmts.GenSelectGroupBy(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT GROUP BY:", err)
			return "SELECT 1"
		}
		return stmtG.SQL()
	case StmtSelectHaving:
		stmtH, err := stmts.GenSelectHaving(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT HAVING:", err)
			return "SELECT 1"
		}
		return stmtH.SQL()
	case StmtSelectJoin:
		stmtJ, err := stmts.GenSelectJoin(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT JOIN:", err)
			return "SELECT 1"
		}
		return stmtJ.SQL()
	case StmtSelectCross:
		stmtCJ, err := stmts.GenSelectCrossJoin(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT CROSS JOIN:", err)
			return "SELECT 1"
		}
		return stmtCJ.SQL()
	case StmtSelectInner:
		stmtIJ, err := stmts.GenSelectInnerJoin(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT INNER JOIN:", err)
			return "SELECT 1"
		}
		return stmtIJ.SQL()
	case StmtSelectOuter:
		stmtOJ, err := stmts.GenSelectOuterJoin(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT OUTER JOIN:", err)
			return "SELECT 1"
		}
		return stmtOJ.SQL()
	case StmtSelectJoinUsing:
		stmtJU, err := stmts.GenSelectJoinUsing(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT JOIN USING:", err)
			return "SELECT 1"
		}
		return stmtJU.SQL()
	case StmtSelectNatural:
		stmtNJ, err := stmts.GenSelectNaturalJoin(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT NATURAL JOIN:", err)
			return "SELECT 1"
		}
		return stmtNJ.SQL()
	case StmtSelectRecursive:
		stmt, err := stmts.GenSelectRecursive(db, g.lcg, g.maxRecursionDepth)
		if err != nil {
			fmt.Println("Error generating SELECT RECURSIVE:", err)
			return "SELECT 1"
		}
		return stmt.SQL()
	case StmtSelectNestedCase:
		stmt, err := stmts.GenSelectWithNestedCase(db, g.lcg, g.maxRecursionDepth)
		if err != nil {
			fmt.Println("Error generating SELECT NESTED CASE:", err)
			return "SELECT 1"
		}
		return stmt.SQL()
	case StmtSelectComplexJoin:
		stmt, err := stmts.GenSelectWithComplexJoin(db, g.lcg, g.maxRecursionDepth)
		if err != nil {
			fmt.Println("Error generating SELECT COMPLEX JOIN:", err)
			return "SELECT 1"
		}
		return stmt.SQL()
	case StmtSelectWindow:
		stmt, err := stmts.GenSelectWithWindowFunction(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT WINDOW:", err)
			return "SELECT 1"
		}
		return stmt.SQL()
	case StmtSelectMultipleWindows:
		stmt, err := stmts.GenSelectWithMultipleWindows(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT MULTIPLE WINDOWS:", err)
			return "SELECT 1"
		}
		return stmt.SQL()
	case StmtSelectCTE:
		stmt, err := stmts.GenSelectWithCTE(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT CTE:", err)
			return "SELECT 1"
		}
		return stmt.SQL()
	case StmtSelectMultipleCTE:
		stmt, err := stmts.GenSelectWithMultipleCTE(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT MULTIPLE CTE:", err)
			return "SELECT 1"
		}
		return stmt.SQL()
	case StmtSelectRecursiveCTE:
		stmt, err := stmts.GenSelectWithRecursiveCTE(db, g.lcg)
		if err != nil {
			fmt.Println("Error generating SELECT RECURSIVE CTE:", err)
			return "SELECT 1"
		}
		return stmt.SQL()
	case StmtCreateTable:
		stmt, err := stmts.GenCreateTable(g.lcg)
		if err != nil {
			fmt.Println("Error generating CREATE TABLE:", err)
			return "CREATE TABLE IF NOT EXISTS fallback (id INTEGER);" // fallback
		}
		return stmt.SQL()
	case StmtDropTable:
		stmt, err := stmts.GenDropTable(g.lcg)
		if err != nil {
			fmt.Println("Error generating DROP TABLE:", err)
			return "DROP TABLE IF EXISTS fallback;"
		}
		return stmt.SQL()
	case StmtAlterTable:
		stmt, err := stmts.GenAlterTable(g.lcg)
		if err != nil {
			fmt.Println("Error generating ALTER TABLE:", err)
			return "ALTER TABLE fallback RENAME TO fallback2;"
		}
		return stmt.SQL()
	}

	return "SELECT 1" // placeholder for other directions
}

// TokensUsed returns the number of tokens used by the underlying LCG.
func (g *Generator) TokensUsed() uint64 {
	return g.lcg.TokensUsed()
}
