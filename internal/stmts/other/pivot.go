package other

import (
	"sqlfuse/internal/stmts/stmts"
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
	"sqlfuse/internal/stmts/helper"
)

// PivotGenerator is a stmts.StmtGenerator for PIVOT statements.
type PivotGenerator struct{}

// Generate implements stmts.StmtGenerator for PIVOT statements.
func (g *PivotGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	return GenPivot(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements stmts.StmtGenerator. PIVOT can always be generated.
func (g *PivotGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	return true
}

// PivotStmt represents a PIVOT statement.
type PivotStmt struct {
	*stmts.BaseStmt
}

// GenPivot generates a PIVOT statement.
// PIVOT transforms rows into columns.
// Reference: https://duckdb.org/docs/stable/sql/statements/pivot
func GenPivot(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (stmts.Stmt, error) {
	lcg = stmts.EnsureLCG(lcg)
	if flavor == nil {
		flavor = stmts.GetDefaultFlavor()
	}

	// PIVOT basic syntax:
	// PIVOT table_or_query ON column_to_pivot IN (value1, value2, ...) USING aggregate_function(column)

	var sql string
	choice := lcg.Intn(3)

	if choice == 0 {
		// Simple PIVOT with VALUES
		sql = `PIVOT (VALUES (1, 'a', 10), (2, 'b', 20), (3, 'a', 30)) AS t(id, category, value)
ON category IN ('a', 'b')
USING sum(value);`
	} else if choice == 1 {
		// PIVOT with a SELECT query
		sql = `PIVOT (SELECT 1 AS id, 'category_a' AS category, 100 AS amount UNION ALL 
SELECT 2, 'category_b', 200 UNION ALL SELECT 3, 'category_a', 150)
ON category
USING sum(amount);`
	} else {
		// PIVOT with table reference
		var tables []helper.TableInfo
		var err error
		if db != nil {
			tables, err = helper.GetAllTablesAndCols(db, "sqlite")
		}

		if err != nil || len(tables) == 0 {
			// Fallback to VALUES
			sql = `PIVOT (VALUES ('Q1', 'Sales', 1000), ('Q1', 'Marketing', 500), ('Q2', 'Sales', 1200))
AS t(quarter, department, budget)
ON department IN ('Sales', 'Marketing')
USING sum(budget);`
		} else {
			tbl := tables[lcg.Intn(len(tables))]
			if len(tbl.Cols) >= 2 {
				col1 := tbl.Cols[lcg.Intn(len(tbl.Cols))]
				col2 := tbl.Cols[lcg.Intn(len(tbl.Cols))]
				sql = fmt.Sprintf(`PIVOT "%s" ON "%s" USING count("%s");`, tbl.Name, col1.Name, col2.Name)
			} else {
				sql = `PIVOT (VALUES ('a', 1), ('b', 2), ('a', 3)) AS t(key, value) ON key USING sum(value);`
			}
		}
	}

	return &PivotStmt{
		BaseStmt: stmts.NewBaseStmt(sql, "pivot", flavor),
	}, nil
}

// UnpivotGenerator is a stmts.StmtGenerator for UNPIVOT statements.
type UnpivotGenerator struct{}

// Generate implements stmts.StmtGenerator for UNPIVOT statements.
func (g *UnpivotGenerator) Generate(ctx *stmts.GenContext) (stmts.Stmt, error) {
	return GenUnpivot(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements stmts.StmtGenerator. UNPIVOT can always be generated.
func (g *UnpivotGenerator) CanGenerate(ctx *stmts.GenContext) bool {
	return true
}

// UnpivotStmt represents an UNPIVOT statement.
type UnpivotStmt struct {
	*stmts.BaseStmt
}

// GenUnpivot generates an UNPIVOT statement.
// UNPIVOT transforms columns into rows.
// Reference: https://duckdb.org/docs/stable/sql/statements/unpivot
func GenUnpivot(db *sql.DB, lcg *common.LCG, flavor stmts.FlavorConfig) (stmts.Stmt, error) {
	lcg = stmts.EnsureLCG(lcg)
	if flavor == nil {
		flavor = stmts.GetDefaultFlavor()
	}

	// UNPIVOT basic syntax:
	// UNPIVOT table_or_query ON column_list INTO NAME column_name VALUE value_column

	var sql string
	choice := lcg.Intn(3)

	if choice == 0 {
		// Simple UNPIVOT with VALUES
		sql = `UNPIVOT (VALUES (1, 100, 200, 300)) AS t(id, q1, q2, q3)
ON q1, q2, q3
INTO NAME quarter VALUE sales;`
	} else if choice == 1 {
		// UNPIVOT with SELECT
		sql = `UNPIVOT (SELECT 'Product A' AS product, 10 AS jan, 20 AS feb, 30 AS mar)
ON jan, feb, mar
INTO NAME month VALUE sales;`
	} else {
		// UNPIVOT with table reference
		var tables []helper.TableInfo
		var err error
		if db != nil {
			tables, err = helper.GetAllTablesAndCols(db, "sqlite")
		}

		if err != nil || len(tables) == 0 {
			// Fallback
			sql = `UNPIVOT (VALUES ('A', 1, 2, 3), ('B', 4, 5, 6)) AS t(id, col1, col2, col3)
ON col1, col2, col3
INTO NAME column_name VALUE column_value;`
		} else {
			tbl := tables[lcg.Intn(len(tables))]
			if len(tbl.Cols) >= 2 {
				col1 := tbl.Cols[0]
				col2 := tbl.Cols[min(1, len(tbl.Cols)-1)]
				sql = fmt.Sprintf(`UNPIVOT "%s" ON "%s", "%s" INTO NAME attr VALUE val;`,
					tbl.Name, col1.Name, col2.Name)
			} else {
				sql = `UNPIVOT (VALUES (1, 10, 20)) AS t(id, a, b) ON a, b INTO NAME col VALUE val;`
			}
		}
	}

	return &UnpivotStmt{
		BaseStmt: stmts.NewBaseStmt(sql, "unpivot", flavor),
	}, nil
}
