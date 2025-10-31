package stmts

import (
	"sqlfuse/internal/common"
	"testing"
)

// TestNewDuckDBStatements tests the newly added DuckDB-specific statements.
func TestNewDuckDBStatements(t *testing.T) {
	lcg := common.NewLCG(42)
	flavor := &DefaultFlavorConfig{}

	tests := []struct {
		name      string
		generator StmtGenerator
	}{
		{"Pivot", &PivotGenerator{}},
		{"Unpivot", &UnpivotGenerator{}},
		{"MergeInto", &MergeIntoGenerator{}},
		{"Qualify", &QualifyGenerator{}},
		{"AlterDatabase", &AlterDatabaseGenerator{}},
		{"AlterView", &AlterViewGenerator{}},
		{"CreateSecret", &CreateSecretGenerator{}},
		{"DropSecret", &DropSecretGenerator{}},
		{"LoadInstall", &LoadInstallGenerator{}},
		{"CommentOn", &CommentOnGenerator{}},
		{"Profiling", &ProfilingGenerator{}},
		{"SetVariable", &SetVariableGenerator{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewGenContextWithFlavor(nil, lcg, 3, flavor)

			// Test CanGenerate
			canGen := tt.generator.CanGenerate(ctx)
			t.Logf("%s.CanGenerate() = %v", tt.name, canGen)

			// Test Generate
			stmt, err := tt.generator.Generate(ctx)
			if err != nil {
				t.Fatalf("%s.Generate() error = %v", tt.name, err)
			}

			if stmt == nil {
				t.Fatalf("%s.Generate() returned nil statement", tt.name)
			}

			sql := stmt.SQL()
			if sql == "" {
				t.Fatalf("%s.Generate() returned empty SQL", tt.name)
			}

			t.Logf("%s SQL: %s", tt.name, sql)
		})
	}
}

// TestPivotStatementVariants tests different PIVOT statement variants.
func TestPivotStatementVariants(t *testing.T) {
	lcg := common.NewLCG(123)
	flavor := &DefaultFlavorConfig{}

	// Generate multiple PIVOT statements to test variety
	for i := 0; i < 10; i++ {
		stmt, err := GenPivot(nil, lcg, flavor)
		if err != nil {
			t.Fatalf("GenPivot() error = %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Fatal("GenPivot() returned empty SQL")
		}

		// Verify it contains PIVOT keyword
		if !contains(sql, "PIVOT") {
			t.Errorf("PIVOT statement missing PIVOT keyword: %s", sql)
		}

		t.Logf("PIVOT variant %d: %s", i, sql)
	}
}

// TestUnpivotStatementVariants tests different UNPIVOT statement variants.
func TestUnpivotStatementVariants(t *testing.T) {
	lcg := common.NewLCG(456)
	flavor := &DefaultFlavorConfig{}

	// Generate multiple UNPIVOT statements to test variety
	for i := 0; i < 10; i++ {
		stmt, err := GenUnpivot(nil, lcg, flavor)
		if err != nil {
			t.Fatalf("GenUnpivot() error = %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Fatal("GenUnpivot() returned empty SQL")
		}

		// Verify it contains UNPIVOT keyword
		if !contains(sql, "UNPIVOT") {
			t.Errorf("UNPIVOT statement missing UNPIVOT keyword: %s", sql)
		}

		t.Logf("UNPIVOT variant %d: %s", i, sql)
	}
}

// TestMergeIntoStatement tests MERGE INTO statement generation.
func TestMergeIntoStatement(t *testing.T) {
	lcg := common.NewLCG(789)
	flavor := &DefaultFlavorConfig{}

	stmt, err := GenMergeInto(nil, lcg, flavor)
	if err != nil {
		t.Fatalf("GenMergeInto() error = %v", err)
	}

	sql := stmt.SQL()
	if sql == "" {
		t.Fatal("GenMergeInto() returned empty SQL")
	}

	// Verify it contains required MERGE INTO keywords
	if !contains(sql, "MERGE INTO") {
		t.Errorf("MERGE INTO statement missing MERGE INTO keyword: %s", sql)
	}

	if !contains(sql, "USING") {
		t.Errorf("MERGE INTO statement missing USING keyword: %s", sql)
	}

	t.Logf("MERGE INTO: %s", sql)
}

// TestQualifyStatement tests SELECT with QUALIFY clause.
func TestQualifyStatement(t *testing.T) {
	lcg := common.NewLCG(111)
	flavor := &DefaultFlavorConfig{}

	// Generate multiple QUALIFY statements to test variety
	for i := 0; i < 5; i++ {
		stmt, err := GenQualify(nil, lcg, flavor)
		if err != nil {
			t.Fatalf("GenQualify() error = %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Fatal("GenQualify() returned empty SQL")
		}

		// Verify it contains QUALIFY keyword
		if !contains(sql, "QUALIFY") {
			t.Errorf("QUALIFY statement missing QUALIFY keyword: %s", sql)
		}

		t.Logf("QUALIFY variant %d: %s", i, sql)
	}
}

// TestCreateSecretVariants tests different CREATE SECRET variants.
func TestCreateSecretVariants(t *testing.T) {
	lcg := common.NewLCG(222)
	flavor := &DefaultFlavorConfig{}

	// Generate multiple secrets to test variety
	for i := 0; i < 10; i++ {
		stmt, err := GenCreateSecret(nil, lcg, flavor)
		if err != nil {
			t.Fatalf("GenCreateSecret() error = %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Fatal("GenCreateSecret() returned empty SQL")
		}

		// Verify it contains CREATE SECRET keywords
		if !contains(sql, "CREATE SECRET") {
			t.Errorf("CREATE SECRET statement missing CREATE SECRET keyword: %s", sql)
		}

		if !contains(sql, "TYPE") {
			t.Errorf("CREATE SECRET statement missing TYPE keyword: %s", sql)
		}

		t.Logf("CREATE SECRET variant %d: %s", i, sql)
	}
}

// TestLoadInstallStatement tests LOAD/INSTALL statements.
func TestLoadInstallStatement(t *testing.T) {
	lcg := common.NewLCG(333)
	flavor := &DefaultFlavorConfig{}

	// Generate multiple LOAD/INSTALL statements
	for i := 0; i < 10; i++ {
		stmt, err := GenLoadInstall(nil, lcg, flavor)
		if err != nil {
			t.Fatalf("GenLoadInstall() error = %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Fatal("GenLoadInstall() returned empty SQL")
		}

		// Verify it contains either LOAD or INSTALL
		if !contains(sql, "LOAD") && !contains(sql, "INSTALL") {
			t.Errorf("LOAD/INSTALL statement missing LOAD or INSTALL keyword: %s", sql)
		}

		t.Logf("LOAD/INSTALL variant %d: %s", i, sql)
	}
}

// TestCommentOnVariants tests different COMMENT ON variants.
func TestCommentOnVariants(t *testing.T) {
	lcg := common.NewLCG(444)
	flavor := &DefaultFlavorConfig{}

	// Generate multiple COMMENT ON statements to cover different object types
	for i := 0; i < 15; i++ {
		stmt, err := GenCommentOn(nil, lcg, flavor)
		if err != nil {
			t.Fatalf("GenCommentOn() error = %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Fatal("GenCommentOn() returned empty SQL")
		}

		// Verify it contains COMMENT ON keywords
		if !contains(sql, "COMMENT ON") {
			t.Errorf("COMMENT ON statement missing COMMENT ON keyword: %s", sql)
		}

		if !contains(sql, "IS") {
			t.Errorf("COMMENT ON statement missing IS keyword: %s", sql)
		}

		t.Logf("COMMENT ON variant %d: %s", i, sql)
	}
}

// TestSetVariableStatement tests SET VARIABLE statements.
func TestSetVariableStatement(t *testing.T) {
	lcg := common.NewLCG(555)
	flavor := &DefaultFlavorConfig{}

	// Generate multiple SET VARIABLE statements
	for i := 0; i < 10; i++ {
		stmt, err := GenSetVariable(nil, lcg, flavor)
		if err != nil {
			t.Fatalf("GenSetVariable() error = %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Fatal("GenSetVariable() returned empty SQL")
		}

		// Verify it contains SET VARIABLE keywords
		if !contains(sql, "SET VARIABLE") {
			t.Errorf("SET VARIABLE statement missing SET VARIABLE keyword: %s", sql)
		}

		t.Logf("SET VARIABLE variant %d: %s", i, sql)
	}
}
