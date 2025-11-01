package executors

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestCommonFlags_DefaultValues(t *testing.T) {
	var flags CommonFlags
	cmd := &cobra.Command{}
	
	AddCommonFlags(cmd, &flags, "default_init.sql")
	
	// Parse with no arguments to get defaults
	cmd.ParseFlags([]string{})
	
	// Verify default values are set correctly
	if flags.Dsn != ":memory:" {
		t.Errorf("Expected default DSN ':memory:', got '%s'", flags.Dsn)
	}
	
	if flags.InitSQLPath != "default_init.sql" {
		t.Errorf("Expected default InitSQLPath 'default_init.sql', got '%s'", flags.InitSQLPath)
	}
	
	if flags.Workers != 1 {
		t.Errorf("Expected default Workers 1, got %d", flags.Workers)
	}
	
	if flags.Queries != 10 {
		t.Errorf("Expected default Queries 10, got %d", flags.Queries)
	}
	
	if flags.Seed != 0 {
		t.Errorf("Expected default Seed 0, got %d", flags.Seed)
	}
	
	if flags.Verbose {
		t.Error("Expected default Verbose false")
	}
	
	if flags.Weights != "" {
		t.Errorf("Expected default Weights empty, got '%s'", flags.Weights)
	}
}

func TestCommonFlags_CustomValues(t *testing.T) {
	var flags CommonFlags
	cmd := &cobra.Command{}
	
	AddCommonFlags(cmd, &flags, "default_init.sql")
	
	// Parse with custom arguments
	err := cmd.ParseFlags([]string{
		"--dsn", "file:test.db",
		"--init-sql", "custom_init.sql",
		"--workers", "4",
		"--queries", "100",
		"--seed", "42",
		"--verbose",
		"--weights", `{"insert":100}`,
	})
	
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}
	
	// Verify custom values
	if flags.Dsn != "file:test.db" {
		t.Errorf("Expected DSN 'file:test.db', got '%s'", flags.Dsn)
	}
	
	if flags.InitSQLPath != "custom_init.sql" {
		t.Errorf("Expected InitSQLPath 'custom_init.sql', got '%s'", flags.InitSQLPath)
	}
	
	if flags.Workers != 4 {
		t.Errorf("Expected Workers 4, got %d", flags.Workers)
	}
	
	if flags.Queries != 100 {
		t.Errorf("Expected Queries 100, got %d", flags.Queries)
	}
	
	if flags.Seed != 42 {
		t.Errorf("Expected Seed 42, got %d", flags.Seed)
	}
	
	if !flags.Verbose {
		t.Error("Expected Verbose true")
	}
	
	if flags.Weights != `{"insert":100}` {
		t.Errorf("Expected Weights '{\"insert\":100}', got '%s'", flags.Weights)
	}
}

func TestCommonFlags_ShortFlags(t *testing.T) {
	var flags CommonFlags
	cmd := &cobra.Command{}
	
	AddCommonFlags(cmd, &flags, "default_init.sql")
	
	// Parse with short flags
	err := cmd.ParseFlags([]string{
		"-d", "file:short.db",
		"-i", "short_init.sql",
		"-w", "2",
		"-q", "50",
		"-s", "123",
		"-v",
	})
	
	if err != nil {
		t.Fatalf("Failed to parse short flags: %v", err)
	}
	
	// Verify values
	if flags.Dsn != "file:short.db" {
		t.Errorf("Expected DSN 'file:short.db', got '%s'", flags.Dsn)
	}
	
	if flags.Workers != 2 {
		t.Errorf("Expected Workers 2, got %d", flags.Workers)
	}
	
	if flags.Queries != 50 {
		t.Errorf("Expected Queries 50, got %d", flags.Queries)
	}
	
	if flags.Seed != 123 {
		t.Errorf("Expected Seed 123, got %d", flags.Seed)
	}
	
	if !flags.Verbose {
		t.Error("Expected Verbose true")
	}
}

func TestCommonFlags_EmptyInitSQL(t *testing.T) {
	var flags CommonFlags
	cmd := &cobra.Command{}
	
	AddCommonFlags(cmd, &flags, "default_init.sql")
	
	// Set init-sql to empty string
	err := cmd.ParseFlags([]string{"--init-sql", ""})
	
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}
	
	if flags.InitSQLPath != "" {
		t.Errorf("Expected empty InitSQLPath, got '%s'", flags.InitSQLPath)
	}
}

func TestCommonFlags_NegativeSeed(t *testing.T) {
	var flags CommonFlags
	cmd := &cobra.Command{}
	
	AddCommonFlags(cmd, &flags, "default_init.sql")
	
	// Use negative seed
	err := cmd.ParseFlags([]string{"--seed", "-999"})
	
	if err != nil {
		t.Fatalf("Failed to parse flags: %v", err)
	}
	
	if flags.Seed != -999 {
		t.Errorf("Expected Seed -999, got %d", flags.Seed)
	}
}
