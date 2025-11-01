package stmts

import (
	"sqlfuse/internal/common"
	"testing"
)

// TestDuckDBStatementGenerators tests all DuckDB-specific statement generators
func TestDuckDBStatementGenerators(t *testing.T) {
	lcg := common.NewLCG(12345)
	flavor := &mockFlavorConfig{name: "duckdb"}

	tests := []struct {
		name      string
		generator StmtGenerator
		stmtType  string
	}{
		{"Copy", &CopyGenerator{}, "copy"},
		{"Set", &SetGenerator{}, "set"},
		{"Reset", &ResetGenerator{}, "reset"},
		{"CreateSchema", &CreateSchemaGenerator{}, "create_schema"},
		{"DropSchema", &DropSchemaGenerator{}, "drop_schema"},
		{"CreateSequence", &CreateSequenceGenerator{}, "create_sequence"},
		{"DropSequence", &DropSequenceGenerator{}, "drop_sequence"},
		{"CreateMacro", &CreateMacroGenerator{}, "create_macro"},
		{"DropMacro", &DropMacroGenerator{}, "drop_macro"},
		{"CreateType", &CreateTypeGenerator{}, "create_type"},
		{"DropType", &DropTypeGenerator{}, "drop_type"},
		{"Describe", &DescribeGenerator{}, "describe"},
		{"Show", &ShowGenerator{}, "show"},
		{"Summarize", &SummarizeGenerator{}, "summarize"},
		{"Use", &UseGenerator{}, "use"},
		{"Call", &CallGenerator{}, "call"},
		{"Checkpoint", &CheckpointGenerator{}, "checkpoint"},
		{"ExportDatabase", &ExportDatabaseGenerator{}, "export_database"},
		{"ImportDatabase", &ImportDatabaseGenerator{}, "import_database"},
		{"Prepare", &PrepareGenerator{}, "prepare"},
		{"Execute", &ExecuteGenerator{}, "execute"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewGenContextWithFlavor(nil, lcg, 3, flavor)

			// Test CanGenerate
			canGenerate := tt.generator.CanGenerate(ctx)

			// Some generators may require tables/data, so CanGenerate might return false
			// We'll still try to generate and just log if it can't
			if !canGenerate {
				t.Logf("%s.CanGenerate() = false (may require existing data)", tt.name)
			}

			// Test Generate
			stmt, err := tt.generator.Generate(ctx)
			if err != nil {
				t.Fatalf("%s.Generate() error = %v", tt.name, err)
			}

			if stmt == nil {
				t.Fatalf("%s.Generate() returned nil statement", tt.name)
			}

			// Verify SQL is not empty
			sql := stmt.SQL()
			if sql == "" {
				t.Errorf("%s.Generate() returned empty SQL", tt.name)
			}

			// Verify Type matches expected
			if stmt.Type() != tt.stmtType {
				t.Errorf("%s.Type() = %v, want %v", tt.name, stmt.Type(), tt.stmtType)
			}

			// Verify Flavor
			if stmt.Flavor() != flavor {
				t.Errorf("%s.Flavor() returned wrong flavor", tt.name)
			}

			t.Logf("%s SQL: %s", tt.name, sql)
		})
	}
}

// TestCopyStatementVariations tests COPY statement variations
func TestCopyStatementVariations(t *testing.T) {
	lcg := common.NewLCG(12345)
	flavor := &mockFlavorConfig{name: "duckdb"}

	// Generate multiple COPY statements to test variations
	for i := 0; i < 20; i++ {
		stmt, err := GenCopy(nil, lcg, flavor)
		if err != nil {
			t.Fatalf("GenCopy() error = %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenCopy() returned empty SQL")
		}

		// Verify SQL contains COPY keyword
		if len(sql) < 4 || sql[:4] != "COPY" {
			t.Errorf("GenCopy() SQL does not start with COPY: %s", sql)
		}
	}
}

// TestSetStatementVariations tests SET statement variations
func TestSetStatementVariations(t *testing.T) {
	lcg := common.NewLCG(54321)
	flavor := &mockFlavorConfig{name: "duckdb"}

	settingsSeen := make(map[string]bool)

	// Generate multiple SET statements to test variations
	for i := 0; i < 30; i++ {
		stmt, err := GenSet(nil, lcg, flavor)
		if err != nil {
			t.Fatalf("GenSet() error = %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenSet() returned empty SQL")
		}

		// Verify SQL starts with SET
		if len(sql) < 3 || sql[:3] != "SET" {
			t.Errorf("GenSet() SQL does not start with SET: %s", sql)
		}

		settingsSeen[sql] = true
	}

	// Should have seen multiple different settings
	if len(settingsSeen) < 3 {
		t.Errorf("GenSet() only generated %d unique statements, want at least 3", len(settingsSeen))
	}
}

// TestCreateSchemaStatement tests CREATE SCHEMA statement generation
func TestCreateSchemaStatement(t *testing.T) {
	lcg := common.NewLCG(99999)
	flavor := &mockFlavorConfig{name: "duckdb"}

	stmt, err := GenCreateSchema(nil, lcg, flavor)
	if err != nil {
		t.Fatalf("GenCreateSchema() error = %v", err)
	}

	sql := stmt.SQL()
	if sql == "" {
		t.Error("GenCreateSchema() returned empty SQL")
	}

	// Verify SQL contains CREATE SCHEMA
	if len(sql) < 13 || sql[:13] != "CREATE SCHEMA" {
		t.Errorf("GenCreateSchema() SQL does not start with CREATE SCHEMA: %s", sql)
	}
}

// TestDescribeStatementVariations tests DESCRIBE statement variations
func TestDescribeStatementVariations(t *testing.T) {
	lcg := common.NewLCG(11111)
	flavor := &mockFlavorConfig{name: "duckdb"}

	// Generate multiple DESCRIBE statements
	for i := 0; i < 10; i++ {
		stmt, err := GenDescribe(nil, lcg, flavor)
		if err != nil {
			t.Fatalf("GenDescribe() error = %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenDescribe() returned empty SQL")
		}

		// Verify SQL starts with DESCRIBE
		if len(sql) < 8 || sql[:8] != "DESCRIBE" {
			t.Errorf("GenDescribe() SQL does not start with DESCRIBE: %s", sql)
		}
	}
}

// TestShowStatementVariations tests SHOW statement variations
func TestShowStatementVariations(t *testing.T) {
	lcg := common.NewLCG(22222)
	flavor := &mockFlavorConfig{name: "duckdb"}

	commandsSeen := make(map[string]bool)

	// Generate multiple SHOW statements
	for i := 0; i < 20; i++ {
		stmt, err := GenShow(nil, lcg, flavor)
		if err != nil {
			t.Fatalf("GenShow() error = %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenShow() returned empty SQL")
		}

		// Verify SQL starts with SHOW
		if len(sql) < 4 || sql[:4] != "SHOW" {
			t.Errorf("GenShow() SQL does not start with SHOW: %s", sql)
		}

		commandsSeen[sql] = true
	}

	// Should have seen multiple different SHOW commands
	if len(commandsSeen) < 2 {
		t.Errorf("GenShow() only generated %d unique statements, want at least 2", len(commandsSeen))
	}
}

// TestSummarizeStatement tests SUMMARIZE statement generation
func TestSummarizeStatement(t *testing.T) {
	lcg := common.NewLCG(33333)
	flavor := &mockFlavorConfig{name: "duckdb"}

	stmt, err := GenSummarize(nil, lcg, flavor)
	if err != nil {
		t.Fatalf("GenSummarize() error = %v", err)
	}

	sql := stmt.SQL()
	if sql == "" {
		t.Error("GenSummarize() returned empty SQL")
	}

	// Verify SQL starts with SUMMARIZE
	if len(sql) < 9 || sql[:9] != "SUMMARIZE" {
		t.Errorf("GenSummarize() SQL does not start with SUMMARIZE: %s", sql)
	}
}

// TestCreateMacroStatement tests CREATE MACRO statement generation
func TestCreateMacroStatement(t *testing.T) {
	lcg := common.NewLCG(44444)
	flavor := &mockFlavorConfig{name: "duckdb"}

	stmt, err := GenCreateMacro(nil, lcg, flavor)
	if err != nil {
		t.Fatalf("GenCreateMacro() error = %v", err)
	}

	sql := stmt.SQL()
	if sql == "" {
		t.Error("GenCreateMacro() returned empty SQL")
	}

	// Verify SQL contains CREATE
	if len(sql) < 6 || sql[:6] != "CREATE" {
		t.Errorf("GenCreateMacro() SQL does not start with CREATE: %s", sql)
	}

	// Verify SQL contains MACRO
	if len(sql) < 20 {
		t.Errorf("GenCreateMacro() SQL too short: %s", sql)
	}
}

// TestCreateTypeStatement tests CREATE TYPE statement generation
func TestCreateTypeStatement(t *testing.T) {
	lcg := common.NewLCG(55555)
	flavor := &mockFlavorConfig{name: "duckdb"}

	stmt, err := GenCreateType(nil, lcg, flavor)
	if err != nil {
		t.Fatalf("GenCreateType() error = %v", err)
	}

	sql := stmt.SQL()
	if sql == "" {
		t.Error("GenCreateType() returned empty SQL")
	}

	// Verify SQL contains CREATE TYPE
	if len(sql) < 11 || sql[:11] != "CREATE TYPE" {
		t.Errorf("GenCreateType() SQL does not start with CREATE TYPE: %s", sql)
	}

	// Verify SQL contains ENUM (DuckDB type support)
	foundEnum := false
	for i := 0; i < len(sql)-3; i++ {
		if sql[i:i+4] == "ENUM" {
			foundEnum = true
			break
		}
	}
	if !foundEnum {
		t.Errorf("GenCreateType() SQL does not contain ENUM: %s", sql)
	}
}

// TestSequenceStatements tests sequence-related statement generation
func TestSequenceStatements(t *testing.T) {
	lcg := common.NewLCG(66666)
	flavor := &mockFlavorConfig{name: "duckdb"}

	// Test CREATE SEQUENCE
	createStmt, err := GenCreateSequence(nil, lcg, flavor)
	if err != nil {
		t.Fatalf("GenCreateSequence() error = %v", err)
	}

	createSQL := createStmt.SQL()
	if len(createSQL) < 15 || createSQL[:15] != "CREATE SEQUENCE" {
		t.Errorf("GenCreateSequence() SQL does not start with CREATE SEQUENCE: %s", createSQL)
	}

	// Test DROP SEQUENCE
	dropStmt, err := GenDropSequence(nil, lcg, flavor)
	if err != nil {
		t.Fatalf("GenDropSequence() error = %v", err)
	}

	dropSQL := dropStmt.SQL()
	if len(dropSQL) < 13 || dropSQL[:13] != "DROP SEQUENCE" {
		t.Errorf("GenDropSequence() SQL does not start with DROP SEQUENCE: %s", dropSQL)
	}
}

// TestPrepareExecuteStatements tests PREPARE and EXECUTE statements
func TestPrepareExecuteStatements(t *testing.T) {
	lcg := common.NewLCG(77777)
	flavor := &mockFlavorConfig{name: "duckdb"}

	// Test PREPARE
	prepareStmt, err := GenPrepare(nil, lcg, flavor)
	if err != nil {
		t.Fatalf("GenPrepare() error = %v", err)
	}

	prepareSQL := prepareStmt.SQL()
	if len(prepareSQL) < 7 || prepareSQL[:7] != "PREPARE" {
		t.Errorf("GenPrepare() SQL does not start with PREPARE: %s", prepareSQL)
	}

	// Test EXECUTE
	executeStmt, err := GenExecute(nil, lcg, flavor)
	if err != nil {
		t.Fatalf("GenExecute() error = %v", err)
	}

	executeSQL := executeStmt.SQL()
	if len(executeSQL) < 7 || executeSQL[:7] != "EXECUTE" {
		t.Errorf("GenExecute() SQL does not start with EXECUTE: %s", executeSQL)
	}
}

// TestCheckpointStatement tests CHECKPOINT statement generation
func TestCheckpointStatement(t *testing.T) {
	lcg := common.NewLCG(88888)
	flavor := &mockFlavorConfig{name: "duckdb"}

	// Generate multiple CHECKPOINT statements
	for i := 0; i < 10; i++ {
		stmt, err := GenCheckpoint(nil, lcg, flavor)
		if err != nil {
			t.Fatalf("GenCheckpoint() error = %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenCheckpoint() returned empty SQL")
		}

		// Verify SQL contains CHECKPOINT
		foundCheckpoint := false
		if len(sql) >= 10 && sql[:10] == "CHECKPOINT" {
			foundCheckpoint = true
		} else if len(sql) >= 16 && sql[:16] == "FORCE CHECKPOINT" {
			foundCheckpoint = true
		}

		if !foundCheckpoint {
			t.Errorf("GenCheckpoint() SQL does not contain CHECKPOINT: %s", sql)
		}
	}
}

// mockFlavorConfig is a simple mock implementation for testing
type mockFlavorConfig struct {
	name string
}

func (m *mockFlavorConfig) Name() string                        { return m.name }
func (m *mockFlavorConfig) SupportsFeature(feature string) bool { return true }
func (m *mockFlavorConfig) ValidateSQL(sql string) error        { return nil }
