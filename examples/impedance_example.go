// Example demonstrating impedance matching and statistics tracking
// inspired by the original SQLsmith's approach to adaptive fuzzing.
package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"sqlsmith-go/internal/generators"
	"sqlsmith-go/internal/stmts/stmts"
)

func main() {
	// Create an in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create some test tables
	for _, stmt := range []string{
		"CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)",
		"CREATE TABLE accounts (id INTEGER PRIMARY KEY, user_id INTEGER, balance REAL)",
		"INSERT INTO users VALUES (1, 'Alice', 30)",
		"INSERT INTO users VALUES (2, 'Bob', 25)",
		"INSERT INTO accounts VALUES (1, 1, 100.50)",
		"INSERT INTO accounts VALUES (2, 2, 200.75)",
	} {
		if _, err := db.Exec(stmt); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating schema: %v\n", err)
			os.Exit(1)
		}
	}

	// Create generator with impedance matching and statistics enabled
	seed := uint64(time.Now().UnixNano())
	gen := generators.NewGoSQLite3Generator(seed)

	// Enable the new features
	gen.EnableImpedanceMatching(true)
	gen.EnableStatistics(true)

	stats := gen.GetStatistics()
	impedance := gen.GetImpedanceMatcher()

	// Configure impedance matching thresholds
	impedance.SetBlacklistThreshold(0.90) // Blacklist if >90% error rate
	impedance.SetMinObservations(10)       // Need at least 10 observations

	fmt.Println("SQLsmith-Go Example: Impedance Matching & Statistics")
	fmt.Println("====================================================")
	fmt.Println()

	// Generate and execute queries
	totalQueries := 100
	for i := 0; i < totalQueries; i++ {
		stmtType := gen.Direction()
		
		// Skip blacklisted types
		if gen.IsBlacklisted(stmtType) {
			continue
		}

		genStart := time.Now()
		query := gen.GenerateWithDB(db)
		genTime := time.Since(genStart)

		stats.RecordGeneration(genTime)

		// Try to execute
		execStart := time.Now()
		_, execErr := db.Exec(query)
		execTime := time.Since(execStart)

		if execErr != nil {
			// Record failure
			gen.RecordFailure(stmtType)
			stats.RecordExecutionError(execErr.Error())
		} else {
			// Record success
			gen.RecordSuccess(stmtType)
			stats.RecordExecution(execTime)
		}

		// Print progress every 20 queries
		if (i+1)%20 == 0 {
			snapshot := stats.GetStats()
			fmt.Printf("Progress: %d queries | Gen: %.1f/s | Exec: %.1f/s | Errors: %.2f%%\n",
				i+1, snapshot.GenPerSec, snapshot.ExecPerSec, snapshot.ErrorRate*100)
		}
	}

	// Print final statistics
	fmt.Println()
	fmt.Println("Final Statistics:")
	fmt.Println("================")
	fmt.Println(stats.Report())

	// Print impedance report
	fmt.Println()
	fmt.Println("Impedance Report:")
	fmt.Println("================")
	fmt.Println(impedance.Report())

	// Demonstrate blacklisted types
	fmt.Println()
	fmt.Println("Blacklisted Statement Types:")
	fmt.Println("============================")
	foundBlacklisted := false
	for _, t := range stmts.AllStmtTypes {
		if gen.IsBlacklisted(t) {
			errorRate := impedance.GetErrorRate(string(t))
			fmt.Printf("- %s (error rate: %.2f%%)\n", t, errorRate*100)
			foundBlacklisted = true
		}
	}
	if !foundBlacklisted {
		fmt.Println("No statement types blacklisted (good compatibility!)")
	}
}
