package oracles

import (
	"database/sql"
	"sqlfuse/internal/common"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create test table with data
	_, err = db.Exec(`
		CREATE TABLE test_table (
			id INTEGER PRIMARY KEY,
			value INTEGER,
			name TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Insert test data
	for i := 0; i < 100; i++ {
		_, err = db.Exec("INSERT INTO test_table (id, value, name) VALUES (?, ?, ?)",
			i, i%10, "test"+string(rune(i%26+'a')))
		if err != nil {
			t.Fatalf("Failed to insert data: %v", err)
		}
	}

	return db
}

func TestTLPOracle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(42)
	columns := []string{"id", "value", "name"}
	oracle := NewTLPOracle(db, lcg, "test_table", columns)

	// Run multiple checks
	for i := 0; i < 10; i++ {
		err := oracle.Check()
		if err != nil {
			t.Errorf("TLP check %d failed: %v", i, err)
		}
	}
}

func TestNoRECOracle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(42)
	columns := []string{"id", "value", "name"}
	oracle := NewNoRECOracle(db, lcg, "test_table", columns)

	// Run multiple checks
	for i := 0; i < 10; i++ {
		err := oracle.Check()
		if err != nil {
			t.Errorf("NoREC check %d failed: %v", i, err)
		}
	}
}

func TestPQSOracle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(42)
	columns := []string{"id", "value", "name"}
	oracle := NewPQSOracle(db, lcg, "test_table", columns)

	// Run multiple checks
	for i := 0; i < 10; i++ {
		err := oracle.Check()
		if err != nil {
			t.Errorf("PQS check %d failed: %v", i, err)
		}
	}
}

func TestQPGOracle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	lcg := common.NewLCG(42)
	columns := []string{"id", "value", "name"}
	oracle := NewQPGOracle(db, lcg, "test_table", columns)

	// Run multiple checks and track plan coverage
	for i := 0; i < 50; i++ {
		err := oracle.Check()
		// QPG may return an error when suggesting state mutation - this is expected
		if err != nil && oracle.ShouldMutateState() {
			t.Logf("QPG suggests state mutation after %d queries", i)
			oracle.ResetCounter()
		} else if err != nil {
			t.Errorf("QPG check %d failed unexpectedly: %v", i, err)
		}
	}

	uniquePlans, _ := oracle.GetPlanCoverage()
	if uniquePlans == 0 {
		t.Error("QPG found no unique query plans")
	}
	t.Logf("QPG found %d unique query plans", uniquePlans)
}

func TestExecuteQuery(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	result, err := ExecuteQuery(db, "SELECT COUNT(*) FROM test_table")
	if err != nil {
		t.Fatalf("ExecuteQuery failed: %v", err)
	}

	if result.RowCount != 1 {
		t.Errorf("Expected 1 row, got %d", result.RowCount)
	}
}

func TestCompareResults(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	result1, err := ExecuteQuery(db, "SELECT * FROM test_table WHERE id < 10 ORDER BY id")
	if err != nil {
		t.Fatalf("ExecuteQuery 1 failed: %v", err)
	}

	result2, err := ExecuteQuery(db, "SELECT * FROM test_table WHERE id < 10 ORDER BY id")
	if err != nil {
		t.Fatalf("ExecuteQuery 2 failed: %v", err)
	}

	if !CompareResults(result1, result2) {
		t.Error("Identical queries should produce identical results")
	}

	result3, err := ExecuteQuery(db, "SELECT * FROM test_table WHERE id < 5 ORDER BY id")
	if err != nil {
		t.Fatalf("ExecuteQuery 3 failed: %v", err)
	}

	if CompareResults(result1, result3) {
		t.Error("Different queries should produce different results")
	}
}
