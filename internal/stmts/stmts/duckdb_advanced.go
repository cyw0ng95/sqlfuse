package stmts

import (
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
	"sqlfuse/internal/stmts/helper"
)

// MergeIntoGenerator is a StmtGenerator for MERGE INTO statements.
type MergeIntoGenerator struct{}

// Generate implements StmtGenerator for MERGE INTO statements.
func (g *MergeIntoGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenMergeInto(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. MERGE INTO requires tables.
func (g *MergeIntoGenerator) CanGenerate(ctx *GenContext) bool {
	return hasTables(ctx.DB)
}

// MergeIntoStmt represents a MERGE INTO statement.
type MergeIntoStmt struct {
	*BaseStmt
}

// GenMergeInto generates a MERGE INTO statement.
// MERGE INTO performs UPSERT-style operations (INSERT, UPDATE, DELETE based on conditions).
// Reference: https://duckdb.org/docs/stable/sql/statements/merge_into
func GenMergeInto(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	// MERGE INTO basic syntax:
	// MERGE INTO target_table USING source_table ON condition
	// WHEN MATCHED THEN UPDATE SET ...
	// WHEN NOT MATCHED THEN INSERT ...

	var sql string
	var tables []helper.TableInfo
	var err error

	if db != nil {
		tables, err = helper.GetAllTablesAndCols(db, "sqlite")
	}

	if err != nil || len(tables) < 2 {
		// Generate with VALUES as source
		sql = `MERGE INTO target_table AS t
USING (VALUES (1, 'updated'), (2, 'new')) AS s(id, name)
ON t.id = s.id
WHEN MATCHED THEN
  UPDATE SET name = s.name
WHEN NOT MATCHED THEN
  INSERT (id, name) VALUES (s.id, s.name);`
	} else {
		// Use existing tables
		targetTable := tables[lcg.Intn(len(tables))]
		sourceTable := tables[lcg.Intn(len(tables))]

		if len(targetTable.Cols) >= 2 && len(sourceTable.Cols) >= 2 {
			targetKeyCol := targetTable.Cols[0]
			targetValCol := targetTable.Cols[1]
			sourceKeyCol := sourceTable.Cols[0]
			sourceValCol := sourceTable.Cols[min(1, len(sourceTable.Cols)-1)]

			sql = fmt.Sprintf(`MERGE INTO "%s" AS t
USING "%s" AS s
ON t."%s" = s."%s"
WHEN MATCHED THEN
  UPDATE SET "%s" = s."%s"
WHEN NOT MATCHED THEN
  INSERT ("%s", "%s") VALUES (s."%s", s."%s");`,
				targetTable.Name,
				sourceTable.Name,
				targetKeyCol.Name, sourceKeyCol.Name,
				targetValCol.Name, sourceValCol.Name,
				targetKeyCol.Name, targetValCol.Name,
				sourceKeyCol.Name, sourceValCol.Name)
		} else {
			// Fallback if columns are insufficient
			sql = `MERGE INTO target_table USING source_table ON target_table.id = source_table.id
WHEN MATCHED THEN UPDATE SET value = source_table.value
WHEN NOT MATCHED THEN INSERT (id, value) VALUES (source_table.id, source_table.value);`
		}
	}

	return &MergeIntoStmt{
		BaseStmt: NewBaseStmt(sql, "merge_into", flavor),
	}, nil
}

// QualifyGenerator is a StmtGenerator for SELECT with QUALIFY clause.
type QualifyGenerator struct{}

// Generate implements StmtGenerator for QUALIFY clause.
func (g *QualifyGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenQualify(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. QUALIFY can always be generated.
func (g *QualifyGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// QualifyStmt represents a SELECT with QUALIFY clause.
type QualifyStmt struct {
	*BaseStmt
}

// GenQualify generates a SELECT statement with QUALIFY clause.
// QUALIFY filters results based on window function results.
// Reference: https://duckdb.org/docs/stable/sql/query_syntax/qualify
func GenQualify(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	// QUALIFY filters window function results, similar to HAVING for aggregates
	var sql string
	choice := lcg.Intn(3)

	if choice == 0 {
		// ROW_NUMBER with QUALIFY
		sql = `SELECT id, value, ROW_NUMBER() OVER (PARTITION BY category ORDER BY value DESC) AS rn
FROM (VALUES (1, 'A', 100), (2, 'A', 200), (3, 'B', 150), (4, 'B', 250)) AS t(id, category, value)
QUALIFY ROW_NUMBER() OVER (PARTITION BY category ORDER BY value DESC) = 1;`
	} else if choice == 1 {
		// RANK with QUALIFY
		sql = `SELECT name, score, RANK() OVER (ORDER BY score DESC) AS rank
FROM (VALUES ('Alice', 95), ('Bob', 90), ('Charlie', 95), ('David', 85)) AS t(name, score)
QUALIFY RANK() OVER (ORDER BY score DESC) <= 2;`
	} else {
		// NTILE with QUALIFY
		sql = `SELECT id, amount, NTILE(4) OVER (ORDER BY amount) AS quartile
FROM (VALUES (1, 100), (2, 200), (3, 300), (4, 400), (5, 500)) AS t(id, amount)
QUALIFY NTILE(4) OVER (ORDER BY amount) = 1;`
	}

	return &QualifyStmt{
		BaseStmt: NewBaseStmt(sql, "select_qualify", flavor),
	}, nil
}

// AlterDatabaseGenerator is a StmtGenerator for ALTER DATABASE statements.
type AlterDatabaseGenerator struct{}

// Generate implements StmtGenerator for ALTER DATABASE statements.
func (g *AlterDatabaseGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenAlterDatabase(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. ALTER DATABASE can always be generated.
func (g *AlterDatabaseGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// AlterDatabaseStmt represents an ALTER DATABASE statement.
type AlterDatabaseStmt struct {
	*BaseStmt
}

// GenAlterDatabase generates an ALTER DATABASE statement.
// Reference: https://duckdb.org/docs/stable/sql/statements/alter_database
func GenAlterDatabase(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	// ALTER DATABASE supports RENAME operation
	var sql string
	
	// Generate random database names
	dbName := fmt.Sprintf("db_%d", lcg.Uint64()%1000)
	newName := fmt.Sprintf("db_%d", lcg.Uint64()%1000)
	
	sql = fmt.Sprintf("ALTER DATABASE %s RENAME TO %s;", dbName, newName)

	return &AlterDatabaseStmt{
		BaseStmt: NewBaseStmt(sql, "alter_database", flavor),
	}, nil
}

// AlterViewGenerator is a StmtGenerator for ALTER VIEW statements.
type AlterViewGenerator struct{}

// Generate implements StmtGenerator for ALTER VIEW statements.
func (g *AlterViewGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenAlterView(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. ALTER VIEW can always be generated.
func (g *AlterViewGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// AlterViewStmt represents an ALTER VIEW statement.
type AlterViewStmt struct {
	*BaseStmt
}

// GenAlterView generates an ALTER VIEW statement.
// Reference: https://duckdb.org/docs/stable/sql/statements/alter_view
func GenAlterView(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	// ALTER VIEW supports RENAME operation
	var sql string
	
	viewName := fmt.Sprintf("view_%d", lcg.Uint64()%1000)
	newName := fmt.Sprintf("view_%d", lcg.Uint64()%1000)
	
	sql = fmt.Sprintf("ALTER VIEW %s RENAME TO %s;", viewName, newName)

	return &AlterViewStmt{
		BaseStmt: NewBaseStmt(sql, "alter_view", flavor),
	}, nil
}
