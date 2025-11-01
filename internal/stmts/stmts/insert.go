package stmts

import (
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
	"sqlfuse/internal/stmts/helper"
	"sqlfuse/internal/stmts/types"
	"strings"
)

// InsertGenerator is a StmtGenerator for basic INSERT statements.
type InsertGenerator struct{}

// Generate implements StmtGenerator for INSERT statements.
func (g *InsertGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return genInsertSingleWithFlavor(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. INSERT requires tables to exist.
func (g *InsertGenerator) CanGenerate(ctx *GenContext) bool {
	return hasTables(ctx.DB)
}

// InsertStmt represents an INSERT statement.
type InsertStmt struct {
	sql    string
	flavor FlavorConfig
}

func (s *InsertStmt) SQL() string          { return s.sql }
func (s *InsertStmt) Type() string         { return "insert" }
func (s *InsertStmt) Flavor() FlavorConfig { return s.flavor }

// GenInsert generates a type-aware INSERT for a random user table.
// This function is kept for backward compatibility with existing code.
// lcg should be *common.LCG.
func GenInsert(db *sql.DB, lcg *common.LCG) (Stmt, error) {
	return genInsertSingle(db, lcg)
}

// genInsertSingle is the internal implementation for single-row inserts.
func genInsertSingle(db *sql.DB, lcg *common.LCG) (Stmt, error) {
	return genInsertSingleWithFlavor(db, lcg, GetDefaultFlavor())
}

// genInsertSingleWithFlavor is the internal implementation for single-row inserts with flavor support.
func genInsertSingleWithFlavor(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}
	tables, err := helper.GetAllTablesAndCols(db, flavor.Name())
	if err != nil || len(tables) == 0 {
		return nil, fmt.Errorf("no tables available for INSERT: %v", err)
	}

	// rnd and u64 helpers
	var rnd func(int) int
	if lcg != nil {
		rnd = lcg.Intn
	} else {
		rnd = func(n int) int { return 0 }
	}

	// pick a table
	tbl := tables[rnd(len(tables))]
	if len(tbl.Cols) == 0 {
		// no columns -> use DEFAULT VALUES
		return &InsertStmt{sql: fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", QuoteIdent(tbl.Name)), flavor: flavor}, nil
	}

	// Filter columns (skip 'id') and build values together
	filteredCols := []helper.ColumnInfo{}
	for _, c := range tbl.Cols {
		if strings.EqualFold(c.Name, "id") { // skip common autoincrement id
			continue
		}
		filteredCols = append(filteredCols, c)
	}

	if len(filteredCols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", QuoteIdent(tbl.Name)), flavor: flavor}, nil
	}

	cols := colsNames(filteredCols)
	vals, err := buildValuesRow(filteredCols, lcg)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", QuoteIdent(tbl.Name), strings.Join(cols, ", "), vals)
	return &InsertStmt{sql: sql, flavor: flavor}, nil
}

// QuoteIdent quotes an SQL identifier (table/column name) for safe use in SQL.
func QuoteIdent(s string) string {
	return fmt.Sprintf("\"%s\"", strings.ReplaceAll(s, "\"", "\"\""))
}

// GenInsertMultiple generates an INSERT with multiple VALUES rows (2..N rows).
func GenInsertMultiple(db *sql.DB, lcgOrRand interface{}) (Stmt, error) {
	// pick number of rows 2..20 for heavier testing
	n := 2
	switch r := lcgOrRand.(type) {
	case interface{ Intn(int) int }:
		n = 2 + r.Intn(19) // r.Intn(19) generates 0..18, so 2 + r.Intn(19) gives 2..20 inclusive
	}
	return genInsertInternal(db, lcgOrRand, n)
}

// GenInsertBulk generates an INSERT with many VALUES rows (20..100 rows) for heavy stress testing.
func GenInsertBulk(db *sql.DB, lcgOrRand interface{}) (Stmt, error) {
	// pick number of rows 20..100 for very heavy testing
	n := 20
	switch r := lcgOrRand.(type) {
	case interface{ Intn(int) int }:
		n = 20 + r.Intn(81) // 20..100
	}
	return genInsertInternal(db, lcgOrRand, n)
}

// GenUpsert generates an INSERT ... ON CONFLICT(...) DO UPDATE statement when possible.
func GenUpsert(db *sql.DB, lcgOrRand interface{}) (Stmt, error) {
	tables, err := helper.GetAllTablesAndCols(db, "sqlite")
	if err != nil || len(tables) == 0 {
		return nil, fmt.Errorf("no tables for upsert: %v", err)
	}

	var rnd func(int) int
	switch r := lcgOrRand.(type) {
	case interface{ Intn(int) int }:
		rnd = r.Intn
	default:
		rnd = func(n int) int { return 0 }
	}

	tbl := tables[rnd(len(tables))]
	if len(tbl.Cols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", QuoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
	}

	// choose a subset of non-PK columns for insert (skip id if present)
	cols := []helper.ColumnInfo{}
	for _, c := range tbl.Cols {
		if strings.EqualFold(c.Name, "id") {
			continue
		}
		cols = append(cols, c)
	}
	if len(cols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", QuoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
	}

	// pick conflict target: try to use first UNIQUE-like column name (email, key) else use id
	conflictCol := ""
	for _, c := range tbl.Cols {
		n := strings.ToLower(c.Name)
		if n == "email" || n == "key" || n == "name" {
			conflictCol = c.Name
			break
		}
	}
	if conflictCol == "" {
		// if id exists use id, else use first column
		for _, c := range tbl.Cols {
			if strings.EqualFold(c.Name, "id") {
				conflictCol = c.Name
				break
			}
		}
		if conflictCol == "" {
			conflictCol = tbl.Cols[0].Name
		}
	}

	// single row values
	vals, err := buildValuesRow(cols, lcgOrRand)
	if err != nil {
		return nil, err
	}

	// build set clause updating non-conflict columns
	setParts := []string{}
	for _, c := range cols {
		if c.Name == conflictCol {
			continue
		}
		setParts = append(setParts, fmt.Sprintf("%s=excluded.%s", QuoteIdent(c.Name), QuoteIdent(c.Name)))
	}
	setClause := ""
	if len(setParts) > 0 {
		setClause = " DO UPDATE SET " + strings.Join(setParts, ", ")
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT(%s)%s;", QuoteIdent(tbl.Name), strings.Join(colsNames(cols), ", "), vals, QuoteIdent(conflictCol), setClause)
	return &InsertStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}

// GenInsertFromSelect generates INSERT INTO t(cols) SELECT ... FROM other_table LIMIT n
func GenInsertFromSelect(db *sql.DB, lcgOrRand interface{}) (Stmt, error) {
	tables, err := helper.GetAllTablesAndCols(db, "sqlite")
	if err != nil || len(tables) < 1 {
		return nil, fmt.Errorf("no tables for insert-select: %v", err)
	}

	var rnd func(int) int
	switch r := lcgOrRand.(type) {
	case interface{ Intn(int) int }:
		rnd = r.Intn
	default:
		rnd = func(n int) int { return 0 }
	}

	target := tables[rnd(len(tables))]
	source := tables[rnd(len(tables))]
	// ensure at least one column
	if len(target.Cols) == 0 || len(source.Cols) == 0 {
		return nil, fmt.Errorf("tables lack columns for insert-select")
	}

	// Find columns present in both tables by name (excluding 'id')
	commonCols := []helper.ColumnInfo{}
	for _, tc := range target.Cols {
		if strings.EqualFold(tc.Name, "id") {
			continue
		}
		for _, sc := range source.Cols {
			if strings.EqualFold(tc.Name, sc.Name) {
				commonCols = append(commonCols, tc)
				break
			}
		}
	}
	if len(commonCols) == 0 {
		return nil, fmt.Errorf("no common columns for insert-select")
	}

	// pick up to 3 columns
	n := 1
	if len(commonCols) > 1 {
		n = 1 + rnd(min(3, len(commonCols)))
	}
	cols := commonCols[:min(n, len(commonCols))]

	// build SELECT projection from source: use the same column names
	selectCols := []string{}
	for _, c := range cols {
		selectCols = append(selectCols, QuoteIdent(c.Name))
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) SELECT %s FROM %s LIMIT %d;", QuoteIdent(target.Name), strings.Join(colsNames(cols), ", "), strings.Join(selectCols, ", "), QuoteIdent(source.Name), 1+rnd(10))
	return &InsertStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}

// genInsertInternal generates n rows inserted into a chosen table.
func genInsertInternal(db *sql.DB, lcgOrRand interface{}, rowsCount int) (Stmt, error) {
	tables, err := helper.GetAllTablesAndCols(db, "sqlite")
	if err != nil || len(tables) == 0 {
		return nil, fmt.Errorf("no tables for insert: %v", err)
	}

	var rnd func(int) int
	switch r := lcgOrRand.(type) {
	case interface{ Intn(int) int }:
		rnd = r.Intn
	default:
		rnd = func(n int) int { return 0 }
	}

	tbl := tables[rnd(len(tables))]
	if len(tbl.Cols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", QuoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
	}

	// Filter columns (skip 'id')
	filteredCols := []helper.ColumnInfo{}
	for _, c := range tbl.Cols {
		if strings.EqualFold(c.Name, "id") {
			continue
		}
		filteredCols = append(filteredCols, c)
	}
	if len(filteredCols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT INTO %s DEFAULT VALUES;", QuoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
	}

	rowsVals := []string{}
	for rIdx := 0; rIdx < rowsCount; rIdx++ {
		vals, err := buildValuesRow(filteredCols, lcgOrRand)
		if err != nil {
			return nil, err
		}
		rowsVals = append(rowsVals, fmt.Sprintf("(%s)", vals))
	}

	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s;", QuoteIdent(tbl.Name), strings.Join(colsNames(filteredCols), ", "), strings.Join(rowsVals, ", "))
	return &InsertStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}

// buildValuesRow returns comma-joined values for the provided columns using lcgOrRand
func buildValuesRow(cols []helper.ColumnInfo, lcgOrRand interface{}) (string, error) {
	// obtain a *common.LCG to pass into types.ValueForType
	var lcg *common.LCG
	switch v := lcgOrRand.(type) {
	case *common.LCG:
		lcg = v
	case interface{ Uint64() uint64 }:
		// seed a new LCG from the provided Uint64 source
		lcg = common.NewLCG(v.Uint64())
	default:
		lcg = common.NewLCG(1)
	}

	vals := []string{}
	for _, c := range cols {
		val := types.ValueForType(c.Type, lcg, c.Name)
		if val == "" {
			val = "NULL"
		}
		vals = append(vals, val)
	}
	return strings.Join(vals, ", "), nil
}

func colsNames(cols []helper.ColumnInfo) []string {
	n := make([]string, len(cols))
	for i, c := range cols {
		n[i] = QuoteIdent(c.Name)
	}
	return n
}

func escapeSingle(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GenInsertOrReplace generates an INSERT OR REPLACE statement
func GenInsertOrReplace(db *sql.DB, lcg *common.LCG) (Stmt, error) {
	tables, err := helper.GetAllTablesAndCols(db, "sqlite")
	if err != nil || len(tables) == 0 {
		return nil, fmt.Errorf("no tables for insert or replace: %v", err)
	}

	var rnd func(int) int
	if lcg != nil {
		rnd = lcg.Intn
	} else {
		rnd = func(n int) int { return 0 }
	}

	tbl := tables[rnd(len(tables))]
	if len(tbl.Cols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT OR REPLACE INTO %s DEFAULT VALUES;", QuoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
	}

	// Filter columns (skip 'id')
	filteredCols := []helper.ColumnInfo{}
	for _, c := range tbl.Cols {
		if strings.EqualFold(c.Name, "id") {
			continue
		}
		filteredCols = append(filteredCols, c)
	}

	if len(filteredCols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT OR REPLACE INTO %s DEFAULT VALUES;", QuoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
	}

	cols := colsNames(filteredCols)
	vals, err := buildValuesRow(filteredCols, lcg)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf("INSERT OR REPLACE INTO %s (%s) VALUES (%s);", QuoteIdent(tbl.Name), strings.Join(cols, ", "), vals)
	return &InsertStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}

// GenInsertOrIgnore generates an INSERT OR IGNORE statement
func GenInsertOrIgnore(db *sql.DB, lcg *common.LCG) (Stmt, error) {
	tables, err := helper.GetAllTablesAndCols(db, "sqlite")
	if err != nil || len(tables) == 0 {
		return nil, fmt.Errorf("no tables for insert or ignore: %v", err)
	}

	var rnd func(int) int
	if lcg != nil {
		rnd = lcg.Intn
	} else {
		rnd = func(n int) int { return 0 }
	}

	tbl := tables[rnd(len(tables))]
	if len(tbl.Cols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT OR IGNORE INTO %s DEFAULT VALUES;", QuoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
	}

	// Filter columns (skip 'id')
	filteredCols := []helper.ColumnInfo{}
	for _, c := range tbl.Cols {
		if strings.EqualFold(c.Name, "id") {
			continue
		}
		filteredCols = append(filteredCols, c)
	}

	if len(filteredCols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT OR IGNORE INTO %s DEFAULT VALUES;", QuoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
	}

	cols := colsNames(filteredCols)
	vals, err := buildValuesRow(filteredCols, lcg)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf("INSERT OR IGNORE INTO %s (%s) VALUES (%s);", QuoteIdent(tbl.Name), strings.Join(cols, ", "), vals)
	return &InsertStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}

// GenInsertOrAbort generates an INSERT OR ABORT statement
func GenInsertOrAbort(db *sql.DB, lcg *common.LCG) (Stmt, error) {
	tables, err := helper.GetAllTablesAndCols(db, "sqlite")
	if err != nil || len(tables) == 0 {
		return nil, fmt.Errorf("no tables for insert or abort: %v", err)
	}

	var rnd func(int) int
	if lcg != nil {
		rnd = lcg.Intn
	} else {
		rnd = func(n int) int { return 0 }
	}

	tbl := tables[rnd(len(tables))]
	if len(tbl.Cols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT OR ABORT INTO %s DEFAULT VALUES;", QuoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
	}

	filteredCols := []helper.ColumnInfo{}
	for _, c := range tbl.Cols {
		if strings.EqualFold(c.Name, "id") {
			continue
		}
		filteredCols = append(filteredCols, c)
	}

	if len(filteredCols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT OR ABORT INTO %s DEFAULT VALUES;", QuoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
	}

	cols := colsNames(filteredCols)
	vals, err := buildValuesRow(filteredCols, lcg)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf("INSERT OR ABORT INTO %s (%s) VALUES (%s);", QuoteIdent(tbl.Name), strings.Join(cols, ", "), vals)
	return &InsertStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}

// GenInsertOrRollback generates an INSERT OR ROLLBACK statement
func GenInsertOrRollback(db *sql.DB, lcg *common.LCG) (Stmt, error) {
	tables, err := helper.GetAllTablesAndCols(db, "sqlite")
	if err != nil || len(tables) == 0 {
		return nil, fmt.Errorf("no tables for insert or rollback: %v", err)
	}

	var rnd func(int) int
	if lcg != nil {
		rnd = lcg.Intn
	} else {
		rnd = func(n int) int { return 0 }
	}

	tbl := tables[rnd(len(tables))]
	if len(tbl.Cols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT OR ROLLBACK INTO %s DEFAULT VALUES;", QuoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
	}

	filteredCols := []helper.ColumnInfo{}
	for _, c := range tbl.Cols {
		if strings.EqualFold(c.Name, "id") {
			continue
		}
		filteredCols = append(filteredCols, c)
	}

	if len(filteredCols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT OR ROLLBACK INTO %s DEFAULT VALUES;", QuoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
	}

	cols := colsNames(filteredCols)
	vals, err := buildValuesRow(filteredCols, lcg)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf("INSERT OR ROLLBACK INTO %s (%s) VALUES (%s);", QuoteIdent(tbl.Name), strings.Join(cols, ", "), vals)
	return &InsertStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}

// GenInsertOrFail generates an INSERT OR FAIL statement
func GenInsertOrFail(db *sql.DB, lcg *common.LCG) (Stmt, error) {
	tables, err := helper.GetAllTablesAndCols(db, "sqlite")
	if err != nil || len(tables) == 0 {
		return nil, fmt.Errorf("no tables for insert or fail: %v", err)
	}

	var rnd func(int) int
	if lcg != nil {
		rnd = lcg.Intn
	} else {
		rnd = func(n int) int { return 0 }
	}

	tbl := tables[rnd(len(tables))]
	if len(tbl.Cols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT OR FAIL INTO %s DEFAULT VALUES;", QuoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
	}

	filteredCols := []helper.ColumnInfo{}
	for _, c := range tbl.Cols {
		if strings.EqualFold(c.Name, "id") {
			continue
		}
		filteredCols = append(filteredCols, c)
	}

	if len(filteredCols) == 0 {
		return &InsertStmt{sql: fmt.Sprintf("INSERT OR FAIL INTO %s DEFAULT VALUES;", QuoteIdent(tbl.Name)), flavor: GetDefaultFlavor()}, nil
	}

	cols := colsNames(filteredCols)
	vals, err := buildValuesRow(filteredCols, lcg)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf("INSERT OR FAIL INTO %s (%s) VALUES (%s);", QuoteIdent(tbl.Name), strings.Join(cols, ", "), vals)
	return &InsertStmt{sql: sql, flavor: GetDefaultFlavor()}, nil
}
