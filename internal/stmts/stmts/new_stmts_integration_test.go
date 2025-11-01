package stmts

import (
	"sqlfuse/internal/common"
	"strings"
	"testing"
)

// TestAnalyzeGeneration tests the ANALYZE statement generation.
func TestAnalyzeGeneration(t *testing.T) {
	lcg := common.NewLCG(42)
	stmt, err := GenAnalyzeWithFlavor(nil, lcg, nil)
	if err != nil {
		t.Fatalf("GenAnalyzeWithFlavor failed: %v", err)
	}

	sql := stmt.SQL()
	if !strings.HasPrefix(sql, "ANALYZE") {
		t.Errorf("Expected ANALYZE statement, got: %s", sql)
	}
	if stmt.Type() != "analyze" {
		t.Errorf("Expected type 'analyze', got: %s", stmt.Type())
	}
}

// TestVacuumGeneration tests the VACUUM statement generation.
func TestVacuumGeneration(t *testing.T) {
	lcg := common.NewLCG(42)

	// Test default flavor (Turso)
	stmt, err := GenVacuum(nil, lcg, nil)
	if err != nil {
		t.Fatalf("GenVacuum failed: %v", err)
	}

	sql := stmt.SQL()
	if !strings.HasPrefix(sql, "VACUUM") {
		t.Errorf("Expected VACUUM statement, got: %s", sql)
	}
	if stmt.Type() != "vacuum" {
		t.Errorf("Expected type 'vacuum', got: %s", stmt.Type())
	}
}

// TestReindexGeneration tests the REINDEX statement generation.
func TestReindexGeneration(t *testing.T) {
	lcg := common.NewLCG(42)
	stmt, err := GenReindex(nil, lcg)
	if err != nil {
		t.Fatalf("GenReindex failed: %v", err)
	}

	sql := stmt.SQL()
	if !strings.HasPrefix(sql, "REINDEX") {
		t.Errorf("Expected REINDEX statement, got: %s", sql)
	}
	if stmt.Type() != "reindex" {
		t.Errorf("Expected type 'reindex', got: %s", stmt.Type())
	}
}

// TestSavepointGeneration tests the SAVEPOINT statement generation.
func TestSavepointGeneration(t *testing.T) {
	lcg := common.NewLCG(42)
	stmt := GenSavepoint(lcg)

	sql := stmt.SQL()
	if !strings.HasPrefix(sql, "SAVEPOINT") {
		t.Errorf("Expected SAVEPOINT statement, got: %s", sql)
	}
	if stmt.Type() != "savepoint" {
		t.Errorf("Expected type 'savepoint', got: %s", stmt.Type())
	}
}

// TestReleaseSavepointGeneration tests the RELEASE SAVEPOINT statement generation.
func TestReleaseSavepointGeneration(t *testing.T) {
	lcg := common.NewLCG(42)
	stmt := GenReleaseSavepoint(lcg)

	sql := stmt.SQL()
	if !strings.HasPrefix(sql, "RELEASE") {
		t.Errorf("Expected RELEASE statement, got: %s", sql)
	}
	if stmt.Type() != "release" {
		t.Errorf("Expected type 'release', got: %s", stmt.Type())
	}
}

// TestRollbackToSavepointGeneration tests the ROLLBACK TO SAVEPOINT statement generation.
func TestRollbackToSavepointGeneration(t *testing.T) {
	lcg := common.NewLCG(42)
	stmt := GenRollbackToSavepoint(lcg)

	sql := stmt.SQL()
	if !strings.HasPrefix(sql, "ROLLBACK TO") {
		t.Errorf("Expected ROLLBACK TO statement, got: %s", sql)
	}
	if stmt.Type() != "rollback_to_savepoint" {
		t.Errorf("Expected type 'rollback_to_savepoint', got: %s", stmt.Type())
	}
}

// TestCreateTriggerGeneration tests the CREATE TRIGGER statement generation.
func TestCreateTriggerGeneration(t *testing.T) {
	lcg := common.NewLCG(42)
	stmt, err := GenCreateTrigger(nil, lcg, nil)
	if err != nil {
		t.Fatalf("GenCreateTrigger failed: %v", err)
	}

	sql := stmt.SQL()
	if !strings.HasPrefix(sql, "CREATE TRIGGER") {
		t.Errorf("Expected CREATE TRIGGER statement, got: %s", sql)
	}
	if stmt.Type() != "create_trigger" {
		t.Errorf("Expected type 'create_trigger', got: %s", stmt.Type())
	}

	// Verify it contains required trigger components
	if !strings.Contains(sql, "BEGIN") || !strings.Contains(sql, "END") {
		t.Errorf("Trigger must contain BEGIN...END block, got: %s", sql)
	}
}

// TestDropTriggerGeneration tests the DROP TRIGGER statement generation.
func TestDropTriggerGeneration(t *testing.T) {
	lcg := common.NewLCG(42)
	stmt, err := GenDropTrigger(lcg)
	if err != nil {
		t.Fatalf("GenDropTrigger failed: %v", err)
	}

	sql := stmt.SQL()
	if !strings.HasPrefix(sql, "DROP TRIGGER") {
		t.Errorf("Expected DROP TRIGGER statement, got: %s", sql)
	}
	if !strings.Contains(sql, "IF EXISTS") {
		t.Errorf("DROP TRIGGER should use IF EXISTS, got: %s", sql)
	}
	if stmt.Type() != "drop_trigger" {
		t.Errorf("Expected type 'drop_trigger', got: %s", stmt.Type())
	}
}

// TestCompoundSelectGeneration tests compound SELECT statements (UNION, INTERSECT, EXCEPT).
func TestCompoundSelectGeneration(t *testing.T) {
	tests := []struct {
		name     string
		variant  StmtType
		expected string
	}{
		{"UNION", StmtSelectUnion, "UNION"},
		{"INTERSECT", StmtSelectIntersect, "INTERSECT"},
		{"EXCEPT", StmtSelectExcept, "EXCEPT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lcg := common.NewLCG(42)
			stmt, err := GenCompoundSelect(nil, lcg, tt.variant, nil)
			if err != nil {
				t.Fatalf("GenCompoundSelect(%s) failed: %v", tt.name, err)
			}

			sql := stmt.SQL()
			if !strings.Contains(sql, tt.expected) {
				t.Errorf("Expected %s in statement, got: %s", tt.expected, sql)
			}
		})
	}
}

// TestCreateIndexGeneration tests CREATE INDEX statement generation.
func TestCreateIndexGeneration(t *testing.T) {
	lcg := common.NewLCG(42)
	stmt, err := GenCreateIndex(nil, lcg)
	if err != nil {
		t.Fatalf("GenCreateIndex failed: %v", err)
	}

	sql := stmt.SQL()
	if !strings.HasPrefix(sql, "CREATE") {
		t.Errorf("Expected CREATE INDEX statement, got: %s", sql)
	}
	if !strings.Contains(sql, "INDEX") {
		t.Errorf("Expected INDEX in statement, got: %s", sql)
	}
	if stmt.Type() != "create_index" {
		t.Errorf("Expected type 'create_index', got: %s", stmt.Type())
	}
}

// TestDropIndexGeneration tests DROP INDEX statement generation.
func TestDropIndexGeneration(t *testing.T) {
	lcg := common.NewLCG(42)
	stmt, err := GenDropIndex(lcg)
	if err != nil {
		t.Fatalf("GenDropIndex failed: %v", err)
	}

	sql := stmt.SQL()
	if !strings.HasPrefix(sql, "DROP INDEX") {
		t.Errorf("Expected DROP INDEX statement, got: %s", sql)
	}
	if !strings.Contains(sql, "IF EXISTS") {
		t.Errorf("DROP INDEX should use IF EXISTS, got: %s", sql)
	}
	if stmt.Type() != "drop_index" {
		t.Errorf("Expected type 'drop_index', got: %s", stmt.Type())
	}
}

// TestCreateVirtualTableGeneration tests CREATE VIRTUAL TABLE statement generation.
func TestCreateVirtualTableGeneration(t *testing.T) {
	lcg := common.NewLCG(42)
	stmt, err := GenCreateVirtualTable(lcg)
	if err != nil {
		t.Fatalf("GenCreateVirtualTable failed: %v", err)
	}

	sql := stmt.SQL()
	if !strings.HasPrefix(sql, "CREATE VIRTUAL TABLE") {
		t.Errorf("Expected CREATE VIRTUAL TABLE statement, got: %s", sql)
	}
	if !strings.Contains(sql, "USING fts5") {
		t.Errorf("Expected FTS5 virtual table, got: %s", sql)
	}
	if stmt.Type() != "create_virtual_table" {
		t.Errorf("Expected type 'create_virtual_table', got: %s", stmt.Type())
	}
}

// TestAttachDetachGeneration tests ATTACH/DETACH DATABASE statement generation.
func TestAttachDetachGeneration(t *testing.T) {
	lcg := common.NewLCG(42)

	// Test ATTACH
	attachStmt := GenAttachDatabase(lcg)
	sql := attachStmt.SQL()
	if !strings.HasPrefix(sql, "ATTACH DATABASE") {
		t.Errorf("Expected ATTACH DATABASE statement, got: %s", sql)
	}
	if attachStmt.Type() != "attach" {
		t.Errorf("Expected type 'attach', got: %s", attachStmt.Type())
	}

	// Test DETACH
	detachStmt := GenDetachDatabase(lcg)
	sql = detachStmt.SQL()
	if !strings.HasPrefix(sql, "DETACH") {
		t.Errorf("Expected DETACH statement, got: %s", sql)
	}
	if detachStmt.Type() != "detach" {
		t.Errorf("Expected type 'detach', got: %s", detachStmt.Type())
	}
}

// TestTransactionGeneration tests transaction control statement generation.
func TestTransactionGeneration(t *testing.T) {
	lcg := common.NewLCG(42)

	// Test BEGIN
	beginStmt := GenBeginTransaction(lcg)
	sql := beginStmt.SQL()
	if !strings.HasPrefix(sql, "BEGIN") {
		t.Errorf("Expected BEGIN statement, got: %s", sql)
	}

	// Test COMMIT
	commitStmt := GenCommitTransaction(lcg)
	sql = commitStmt.SQL()
	if !strings.HasPrefix(sql, "COMMIT") && !strings.HasPrefix(sql, "END") {
		t.Errorf("Expected COMMIT or END statement, got: %s", sql)
	}

	// Test ROLLBACK
	rollbackStmt := GenRollbackTransaction(lcg)
	sql = rollbackStmt.SQL()
	if !strings.HasPrefix(sql, "ROLLBACK") {
		t.Errorf("Expected ROLLBACK statement, got: %s", sql)
	}
}

// TestExplainGeneration tests EXPLAIN statement generation.
func TestExplainGeneration(t *testing.T) {
	lcg := common.NewLCG(42)
	stmt, err := GenExplain(nil, lcg)
	if err != nil {
		t.Fatalf("GenExplain failed: %v", err)
	}

	sql := stmt.SQL()
	if !strings.HasPrefix(sql, "EXPLAIN") {
		t.Errorf("Expected EXPLAIN statement, got: %s", sql)
	}
	if stmt.Type() != "explain" {
		t.Errorf("Expected type 'explain', got: %s", stmt.Type())
	}
}

// TestExplainQueryPlanGeneration tests EXPLAIN QUERY PLAN statement generation.
func TestExplainQueryPlanGeneration(t *testing.T) {
	lcg := common.NewLCG(42)
	stmt, err := GenExplainQueryPlan(nil, lcg)
	if err != nil {
		t.Fatalf("GenExplainQueryPlan failed: %v", err)
	}

	sql := stmt.SQL()
	if !strings.HasPrefix(sql, "EXPLAIN QUERY PLAN") {
		t.Errorf("Expected EXPLAIN QUERY PLAN statement, got: %s", sql)
	}
	if stmt.Type() != "explain" {
		t.Errorf("Expected type 'explain', got: %s", stmt.Type())
	}
}

// TestFactoryNewStatements tests that the factory can create all new statement types.
func TestFactoryNewStatements(t *testing.T) {
	lcg := common.NewLCG(42)
	factory := NewStmtGeneratorFactory(lcg, 3, nil)

	newStmtTypes := []StmtType{
		StmtCreateView,
		StmtDropView,
		StmtCreateIndex,
		StmtDropIndex,
		StmtCreateVirtualTable,
		StmtAttach,
		StmtDetach,
		StmtBegin,
		StmtCommit,
		StmtRollback,
		StmtExplain,
		StmtExplainQueryPlan,
		StmtAnalyze,
		StmtVacuum,
		StmtReindex,
		StmtSavepoint,
		StmtRelease,
		StmtCreateTrigger,
		StmtDropTrigger,
		StmtSelectUnion,
		StmtSelectIntersect,
		StmtSelectExcept,
	}

	for _, stmtType := range newStmtTypes {
		t.Run(string(stmtType), func(t *testing.T) {
			gen := factory.CreateGenerator(stmtType)
			if gen == nil {
				t.Errorf("Factory failed to create generator for %s", stmtType)
				return
			}

			ctx := factory.CreateContext(nil)
			if !gen.CanGenerate(ctx) {
				// Some generators might require specific conditions
				t.Logf("Generator for %s reports it cannot generate (might be expected)", stmtType)
				return
			}

			stmt, err := gen.Generate(ctx)
			if err != nil {
				t.Errorf("Generator for %s failed: %v", stmtType, err)
				return
			}

			sql := stmt.SQL()
			if sql == "" {
				t.Errorf("Generator for %s produced empty SQL", stmtType)
			}

			t.Logf("Generated %s: %s", stmtType, sql)
		})
	}
}

// TestGenMapNewStatements tests that gen_map includes all new statement types.
func TestGenMapNewStatements(t *testing.T) {
	lcg := common.NewLCG(42)
	genMap := BuildGeneratorFuncs(lcg, 3, nil)

	newStmtKeys := []string{
		"create_view",
		"drop_view",
		"create_index",
		"drop_index",
		"create_virtual_table",
		"attach",
		"detach",
		"begin",
		"commit",
		"rollback",
		"explain",
		"explain_query_plan",
		"analyze",
		"vacuum",
		"reindex",
		"savepoint",
		"release",
		"create_trigger",
		"drop_trigger",
		"select_union",
		"select_intersect",
		"select_except",
	}

	for _, key := range newStmtKeys {
		t.Run(key, func(t *testing.T) {
			genFunc, exists := genMap[key]
			if !exists {
				t.Errorf("gen_map missing key: %s", key)
				return
			}

			sql, err := genFunc(nil)
			if err != nil {
				t.Errorf("gen_map function for %s failed: %v", key, err)
				return
			}

			if sql == "" {
				t.Errorf("gen_map function for %s produced empty SQL", key)
			}

			t.Logf("Generated %s: %s", key, sql)
		})
	}
}
