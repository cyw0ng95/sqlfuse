package stmts

import (
	"fmt"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	parser "github.com/libsql/sqlite-antlr4-parser/sqliteparser"
	"sqlsmith-go/internal/common"
)

const (
	// testIterations defines the standard number of iterations for statement generation tests
	testIterations = 10
)

// errorListener captures syntax errors from the ANTLR parser
type errorListener struct {
	*antlr.DefaultErrorListener
	errors []string
}

func (el *errorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	el.errors = append(el.errors, fmt.Sprintf("line %d:%d %s", line, column, msg))
}

// ValidateSQL validates SQL syntax using the sqlite-antlr4-parser
func ValidateSQL(sql string) (bool, []string) {
	input := antlr.NewInputStream(sql)
	lexer := parser.NewSQLiteLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewSQLiteParser(stream)

	errListener := &errorListener{errors: []string{}}
	p.RemoveErrorListeners()
	p.AddErrorListener(errListener)

	_ = p.Parse()

	return len(errListener.errors) == 0, errListener.errors
}

// TestGenCreateTable tests CREATE TABLE statement generation
func TestGenCreateTable(t *testing.T) {
	lcg := common.NewLCG(42)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenCreateTable(lcg)
		if err != nil {
			t.Fatalf("GenCreateTable failed: %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenCreateTable returned empty SQL")
		}

		if stmt.Type() != "create_table" {
			t.Errorf("Expected type 'create_table', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid CREATE TABLE SQL: %s\nErrors: %v", sql, errors)
		}
	}
}

// TestGenDropTable tests DROP TABLE statement generation
func TestGenDropTable(t *testing.T) {
	lcg := common.NewLCG(123)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenDropTable(lcg)
		if err != nil {
			t.Fatalf("GenDropTable failed: %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenDropTable returned empty SQL")
		}

		if stmt.Type() != "drop_table" {
			t.Errorf("Expected type 'drop_table', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid DROP TABLE SQL: %s\nErrors: %v", sql, errors)
		}
	}
}

// TestGenUpdate tests UPDATE statement generation
func TestGenUpdate(t *testing.T) {
	lcg := common.NewLCG(456)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenUpdate(lcg)
		if err != nil {
			t.Fatalf("GenUpdate failed: %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenUpdate returned empty SQL")
		}

		if stmt.Type() != "update" {
			t.Errorf("Expected type 'update', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid UPDATE SQL: %s\nErrors: %v", sql, errors)
		}
	}
}

// TestGenDelete tests DELETE statement generation
func TestGenDelete(t *testing.T) {
	lcg := common.NewLCG(789)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenDelete(lcg)
		if err != nil {
			t.Fatalf("GenDelete failed: %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenDelete returned empty SQL")
		}

		if stmt.Type() != "delete" {
			t.Errorf("Expected type 'delete', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid DELETE SQL: %s\nErrors: %v", sql, errors)
		}
	}
}

// TestGenPragma tests PRAGMA statement generation
func TestGenPragma(t *testing.T) {
	lcg := common.NewLCG(999)

	for i := 0; i < testIterations; i++ {
		stmt := GenPragma(lcg)

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenPragma returned empty SQL")
		}

		if stmt.Type() != "pragma" {
			t.Errorf("Expected type 'pragma', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid PRAGMA SQL: %s\nErrors: %v", sql, errors)
		}
	}
}

// TestGenCreateView tests CREATE VIEW statement generation
func TestGenCreateView(t *testing.T) {
	lcg := common.NewLCG(333)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenCreateView(lcg)
		if err != nil {
			t.Fatalf("GenCreateView failed: %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenCreateView returned empty SQL")
		}

		if stmt.Type() != "create_view" {
			t.Errorf("Expected type 'create_view', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid CREATE VIEW SQL: %s\nErrors: %v", sql, errors)
		}
	}
}

// TestGenDropView tests DROP VIEW statement generation
func TestGenDropView(t *testing.T) {
	lcg := common.NewLCG(444)

	for i := 0; i < testIterations; i++ {
		stmt, err := GenDropView(lcg)
		if err != nil {
			t.Fatalf("GenDropView failed: %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenDropView returned empty SQL")
		}

		if stmt.Type() != "drop_view" {
			t.Errorf("Expected type 'drop_view', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid DROP VIEW SQL: %s\nErrors: %v", sql, errors)
		}
	}
}

// TestGenAlterTable tests ALTER TABLE statement generation
func TestGenAlterTable(t *testing.T) {
	lcg := common.NewLCG(555)

	// Test multiple iterations to cover different ALTER TABLE operations
	for i := 0; i < testIterations; i++ {
		stmt, err := GenAlterTable(lcg)
		if err != nil {
			t.Fatalf("GenAlterTable failed: %v", err)
		}

		sql := stmt.SQL()
		if sql == "" {
			t.Error("GenAlterTable returned empty SQL")
		}

		if stmt.Type() != "alter_table" {
			t.Errorf("Expected type 'alter_table', got '%s'", stmt.Type())
		}

		valid, errors := ValidateSQL(sql)
		if !valid {
			t.Errorf("Invalid ALTER TABLE SQL: %s\nErrors: %v", sql, errors)
		}
	}
}

// TestStmtInterface verifies all statement types implement the Stmt interface
func TestStmtInterface(t *testing.T) {
	tests := []struct {
		name     string
		stmt     Stmt
		wantType string
	}{
		{"PragmaStmt", &PragmaStmt{sql: "PRAGMA page_size;"}, "pragma"},
		{"SelectStmt", &SelectStmt{sql: "SELECT 1;"}, "select"},
		{"InsertStmt", &InsertStmt{sql: "INSERT INTO t DEFAULT VALUES;"}, "insert"},
		{"UpdateStmt", &UpdateStmt{sql: "UPDATE t SET c1=1;"}, "update"},
		{"DeleteStmt", &DeleteStmt{sql: "DELETE FROM t;"}, "delete"},
		{"CreateTableStmt", &CreateTableStmt{sql: "CREATE TABLE t (id INTEGER);"}, "create_table"},
		{"DropTableStmt", &DropTableStmt{sql: "DROP TABLE t;"}, "drop_table"},
		{"CreateViewStmt", &CreateViewStmt{sql: "CREATE VIEW v AS SELECT 1;"}, "create_view"},
		{"DropViewStmt", &DropViewStmt{sql: "DROP VIEW v;"}, "drop_view"},
		{"AlterTableStmt", &AlterTableStmt{sql: "ALTER TABLE t ADD COLUMN c TEXT;"}, "alter_table"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.stmt.SQL() == "" {
				t.Error("SQL() returned empty string")
			}
			if tt.stmt.Type() != tt.wantType {
				t.Errorf("Type() = %s, want %s", tt.stmt.Type(), tt.wantType)
			}
		})
	}
}

// TestGeneratedSQLDeterminism tests that same seed produces same SQL
func TestGeneratedSQLDeterminism(t *testing.T) {
	tests := []struct {
		name string
		gen  func(*common.LCG) (Stmt, error)
	}{
		{"CreateTable", func(lcg *common.LCG) (Stmt, error) { return GenCreateTable(lcg) }},
		{"DropTable", func(lcg *common.LCG) (Stmt, error) { return GenDropTable(lcg) }},
		{"Update", func(lcg *common.LCG) (Stmt, error) { return GenUpdate(lcg) }},
		{"Delete", func(lcg *common.LCG) (Stmt, error) { return GenDelete(lcg) }},
		{"CreateView", func(lcg *common.LCG) (Stmt, error) { return GenCreateView(lcg) }},
		{"DropView", func(lcg *common.LCG) (Stmt, error) { return GenDropView(lcg) }},
		{"AlterTable", func(lcg *common.LCG) (Stmt, error) { return GenAlterTable(lcg) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lcg1 := common.NewLCG(12345)
			stmt1, err1 := tt.gen(lcg1)
			if err1 != nil {
				t.Fatalf("First generation failed: %v", err1)
			}

			lcg2 := common.NewLCG(12345)
			stmt2, err2 := tt.gen(lcg2)
			if err2 != nil {
				t.Fatalf("Second generation failed: %v", err2)
			}

			if stmt1.SQL() != stmt2.SQL() {
				t.Errorf("Same seed produced different SQL:\n  First:  %s\n  Second: %s",
					stmt1.SQL(), stmt2.SQL())
			}
		})
	}
}

// TestPragmaDeterminism tests PRAGMA generation determinism
func TestPragmaDeterminism(t *testing.T) {
	lcg1 := common.NewLCG(54321)
	stmt1 := GenPragma(lcg1)

	lcg2 := common.NewLCG(54321)
	stmt2 := GenPragma(lcg2)

	if stmt1.SQL() != stmt2.SQL() {
		t.Errorf("Same seed produced different PRAGMA SQL:\n  First:  %s\n  Second: %s",
			stmt1.SQL(), stmt2.SQL())
	}
}

// TestSQLSyntaxValidity validates that generated SQL is syntactically correct
func TestSQLSyntaxValidity(t *testing.T) {
	lcg := common.NewLCG(777)

	generators := []struct {
		name string
		fn   func() (Stmt, error)
	}{
		{"CreateTable", func() (Stmt, error) { return GenCreateTable(lcg) }},
		{"DropTable", func() (Stmt, error) { return GenDropTable(lcg) }},
		{"Update", func() (Stmt, error) { return GenUpdate(lcg) }},
		{"Delete", func() (Stmt, error) { return GenDelete(lcg) }},
		{"Pragma", func() (Stmt, error) { return GenPragma(lcg), nil }},
		{"CreateView", func() (Stmt, error) { return GenCreateView(lcg) }},
		{"DropView", func() (Stmt, error) { return GenDropView(lcg) }},
		{"AlterTable", func() (Stmt, error) { return GenAlterTable(lcg) }},
	}

	for _, gen := range generators {
		t.Run(gen.name, func(t *testing.T) {
			// Generate multiple statements to test variety
			for i := 0; i < testIterations; i++ {
				stmt, err := gen.fn()
				if err != nil {
					t.Fatalf("Generation failed on iteration %d: %v", i, err)
				}

				sql := stmt.SQL()
				valid, errors := ValidateSQL(sql)
				if !valid {
					t.Errorf("Invalid SQL on iteration %d: %s\nErrors: %v", i, sql, errors)
				}
			}
		})
	}
}

// BenchmarkGenCreateTable benchmarks CREATE TABLE generation
func BenchmarkGenCreateTable(b *testing.B) {
	lcg := common.NewLCG(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stmt, err := GenCreateTable(lcg)
		if err != nil {
			b.Fatal(err)
		}
		_ = stmt.SQL()
	}
}

// BenchmarkGenUpdate benchmarks UPDATE generation
func BenchmarkGenUpdate(b *testing.B) {
	lcg := common.NewLCG(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stmt, err := GenUpdate(lcg)
		if err != nil {
			b.Fatal(err)
		}
		_ = stmt.SQL()
	}
}

// BenchmarkGenDelete benchmarks DELETE generation
func BenchmarkGenDelete(b *testing.B) {
	lcg := common.NewLCG(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stmt, err := GenDelete(lcg)
		if err != nil {
			b.Fatal(err)
		}
		_ = stmt.SQL()
	}
}

// BenchmarkSQLValidation benchmarks SQL syntax validation
func BenchmarkSQLValidation(b *testing.B) {
	sql := "CREATE TABLE test (id INTEGER, name TEXT, value REAL);"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		valid, _ := ValidateSQL(sql)
		if !valid {
			b.Fatal("SQL validation failed")
		}
	}
}
