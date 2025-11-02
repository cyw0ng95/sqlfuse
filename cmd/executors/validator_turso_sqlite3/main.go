package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sqlfuse/internal/common"
	"sqlfuse/internal/generators"
	"sqlfuse/internal/stmts/helper"
	"sqlfuse/internal/stmts/stmts"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	_ "github.com/tursodatabase/turso-go"
)

type validatorFlags struct {
	seed      int64
	queries   int
	verbose   bool
	initSQL   string
	stopOnErr bool
}

func main() {
	var flags validatorFlags

	rootCmd := &cobra.Command{
		Use:   "validator_turso_sqlite3",
		Short: "Validator comparing Turso and SQLite3 execution",
		Long: `Validator that executes SQL statements on both Turso and SQLite3 databases,
comparing results to detect differences. This helps ensure compatibility between the two implementations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidator(&flags)
		},
	}

	rootCmd.Flags().Int64Var(&flags.seed, "seed", 0, "Random seed (0 = use timestamp)")
	rootCmd.Flags().IntVar(&flags.queries, "queries", 100, "Number of queries to execute")
	rootCmd.Flags().BoolVar(&flags.verbose, "verbose", false, "Verbose output (show all SQL statements)")
	rootCmd.Flags().StringVar(&flags.initSQL, "init-sql", "/opt/assets/turso/init.sql", "Path to initialization SQL file")
	rootCmd.Flags().BoolVar(&flags.stopOnErr, "stop-on-error", false, "Stop on first error instead of continuing")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runValidator(flags *validatorFlags) error {
	common.InitLogger()
	
	// Validate input flags
	if err := validateFlags(flags); err != nil {
		common.Logger.Error().Msg("Invalid configuration")
		return err
	}
	
	common.Logger.Info().
		Str("action", "validator_start").
		Int("queries", flags.queries).
		Int64("seed", flags.seed).
		Msg("Starting validator_turso_sqlite3")

	// Open Turso database (in-memory)
	tursoDb, err := sql.Open("turso", ":memory:")
	if err != nil {
		common.Logger.Error().Msg("Failed to open database")
		return errors.New("database initialization failed")
	}
	defer tursoDb.Close()

	// Open SQLite3 database (in-memory)
	sqlite3Db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		common.Logger.Error().Msg("Failed to open database")
		return errors.New("database initialization failed")
	}
	defer sqlite3Db.Close()

	// Initialize both databases with the same schema
	if err := initDatabase(tursoDb, flags.initSQL, "Turso"); err != nil {
		return err
	}
	if err := initDatabase(sqlite3Db, flags.initSQL, "SQLite3"); err != nil {
		return err
	}

	common.Logger.Info().
		Str("action", "databases_initialized").
		Msg("Both databases initialized successfully")

	// Verify schema consistency
	if err := verifySchemas(tursoDb, sqlite3Db); err != nil {
		common.Logger.Error().Err(err).Msg("Schema verification failed")
		return err
	}

	// Create generators for both databases
	// Use same seed for reproducibility
	var seed uint64
	if flags.seed != 0 {
		seed = uint64(flags.seed)
	} else {
		seed = uint64(time.Now().UnixNano())
	}

	tursoGen := generators.NewTursoGenerator(seed)
	
	// Focus on easy SQL statements: INSERT, UPDATE, DELETE, SELECT
	// We'll use custom weights to emphasize basic operations
	easyWeights := getEasyStmtWeights()
	tursoGen.SetWeights(easyWeights)

	common.Logger.Info().
		Str("action", "execution_start").
		Int("queries", flags.queries).
		Uint64("seed", seed).
		Msg("Starting query execution")

	bugCount := 0
	errorCount := 0

	for i := 0; i < flags.queries; i++ {
		query := tursoGen.GenerateWithDB(tursoDb)
		queryType := getQueryType(query)

		if flags.verbose {
			common.Logger.Info().
				Int("query_num", i+1).
				Str("query_type", queryType).
				Msg("Executing query")
		}

		// Execute on both databases and compare results
		bug, err := executeAndCompare(tursoDb, sqlite3Db, query, i+1, flags.verbose)
		if err != nil {
			errorCount++
			common.Logger.Warn().
				Str("action", "query_error").
				Int("query_num", i+1).
				Str("query_type", queryType).
				Msg("Query execution error")
			if flags.stopOnErr {
				common.Logger.Error().
					Str("action", "validator_stopped").
					Str("reason", "execution_error").
					Msg("Stopping due to error")
				return errors.New("validation stopped due to execution error")
			}
			continue
		}

		if bug {
			bugCount++
			common.Logger.Error().
				Str("action", "bug_detected").
				Int("query_num", i+1).
				Str("query_type", queryType).
				Msg("Bug detected")
			if flags.stopOnErr {
				common.Logger.Error().
					Str("action", "validator_stopped").
					Str("reason", "bug_detected").
					Msg("Stopping due to bug detection")
				return errors.New("validation stopped due to bug detection")
			}
		}

		// Periodically verify table data consistency
		if (i+1)%20 == 0 {
			if err := compareAllTables(tursoDb, sqlite3Db, i+1, flags.verbose); err != nil {
				bugCount++
				if flags.stopOnErr {
					return err
				}
			}
		}
	}

	// Final table comparison
	common.Logger.Info().
		Str("action", "final_comparison").
		Msg("Performing final table comparison")
	if err := compareAllTables(tursoDb, sqlite3Db, flags.queries, flags.verbose); err != nil {
		bugCount++
	}

	// Summary with audit trail
	common.Logger.Info().
		Str("action", "validation_complete").
		Int("total_queries", flags.queries).
		Int("errors", errorCount).
		Int("bugs", bugCount).
		Time("timestamp", time.Now()).
		Msg("Validation Summary")

	if bugCount > 0 {
		common.Logger.Error().
			Str("action", "validation_failed").
			Int("bug_count", bugCount).
			Msg("VALIDATION FAILED")
		return fmt.Errorf("validation failed with %d bugs", bugCount)
	}

	common.Logger.Info().
		Str("action", "validation_passed").
		Msg("VALIDATION PASSED: No bugs detected")
	return nil
}

func initDatabase(db *sql.DB, initSQLPath, dbName string) error {
	// Validate and sanitize file path
	cleanPath := filepath.Clean(initSQLPath)
	if !filepath.IsAbs(cleanPath) {
		return errors.New("init SQL path must be absolute")
	}

	common.Logger.Info().
		Str("action", "database_init").
		Str("database", dbName).
		Msg("Initializing database")

	initSQL, err := os.ReadFile(cleanPath)
	if err != nil {
		common.Logger.Error().
			Str("database", dbName).
			Msg("Failed to read initialization file")
		return errors.New("failed to read initialization file")
	}

	for _, stmt := range strings.Split(string(initSQL), ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			common.Logger.Error().
				Str("database", dbName).
				Msg("Initialization statement failed")
			return fmt.Errorf("database initialization failed for %s", dbName)
		}
	}

	return nil
}

func verifySchemas(tursoDb, sqlite3Db *sql.DB) error {
	common.Logger.Info().Msg("Verifying schema consistency...")

	tursoTables, err := helper.GetAllTablesAndCols(tursoDb, "turso")
	if err != nil {
		return fmt.Errorf("failed to get Turso schema: %w", err)
	}

	sqlite3Tables, err := helper.GetAllTablesAndCols(sqlite3Db, "go-sqlite3")
	if err != nil {
		return fmt.Errorf("failed to get SQLite3 schema: %w", err)
	}

	if len(tursoTables) != len(sqlite3Tables) {
		return fmt.Errorf("schema mismatch: Turso has %d tables, SQLite3 has %d tables",
			len(tursoTables), len(sqlite3Tables))
	}

	common.Logger.Info().Msgf("Schema verified: %d tables in both databases", len(tursoTables))
	return nil
}

// executeAndCompare executes a query on both databases and compares results
func executeAndCompare(tursoDb, sqlite3Db *sql.DB, query string, queryNum int, verbose bool) (bool, error) {
	// Determine if this is a SELECT query
	trimmedQuery := strings.TrimSpace(strings.ToUpper(query))
	isSelect := strings.HasPrefix(trimmedQuery, "SELECT") || strings.HasPrefix(trimmedQuery, "WITH")

	if isSelect {
		return compareSelectResults(tursoDb, sqlite3Db, query, queryNum, verbose)
	}

	// For non-SELECT queries (INSERT, UPDATE, DELETE), execute and compare affected rows
	return compareModificationResults(tursoDb, sqlite3Db, query, queryNum, verbose)
}

func compareSelectResults(tursoDb, sqlite3Db *sql.DB, query string, queryNum int, verbose bool) (bool, error) {
	// Execute on Turso
	tursoRows, err := tursoDb.Query(query)
	if err != nil {
		// Both should fail or both should succeed
		sqlite3Rows, sqlite3Err := sqlite3Db.Query(query)
		if sqlite3Err != nil {
			// Both failed - this is expected for some queries
			if verbose {
				common.Logger.Debug().
					Int("query_num", queryNum).
					Msg("Both databases failed (expected)")
			}
			return false, nil
		}
		sqlite3Rows.Close()
		common.Logger.Error().
			Str("action", "bug_detected").
			Int("query_num", queryNum).
			Str("issue", "turso_failed_sqlite3_succeeded").
			Msg("BUG: Execution mismatch")
		return true, nil
	}
	defer tursoRows.Close()

	sqlite3Rows, err := sqlite3Db.Query(query)
	if err != nil {
		common.Logger.Error().
			Str("action", "bug_detected").
			Int("query_num", queryNum).
			Str("issue", "turso_succeeded_sqlite3_failed").
			Msg("BUG: Execution mismatch")
		return true, nil
	}
	defer sqlite3Rows.Close()

	// Compare results
	tursoResults := fetchAllRows(tursoRows)
	sqlite3Results := fetchAllRows(sqlite3Rows)

	if !resultsEqual(tursoResults, sqlite3Results) {
		common.Logger.Error().
			Str("action", "bug_detected").
			Int("query_num", queryNum).
			Str("issue", "result_mismatch").
			Int("turso_rows", len(tursoResults)).
			Int("sqlite3_rows", len(sqlite3Results)).
			Msg("BUG: Results differ")
		return true, nil
	}

	return false, nil
}

func compareModificationResults(tursoDb, sqlite3Db *sql.DB, query string, queryNum int, verbose bool) (bool, error) {
	// Execute on Turso
	tursoResult, err := tursoDb.Exec(query)
	tursoAffected := int64(0)
	if err == nil && tursoResult != nil {
		tursoAffected, _ = tursoResult.RowsAffected()
	}

	// Execute on SQLite3
	sqlite3Result, err2 := sqlite3Db.Exec(query)
	sqlite3Affected := int64(0)
	if err2 == nil && sqlite3Result != nil {
		sqlite3Affected, _ = sqlite3Result.RowsAffected()
	}

	// Compare error states
	if (err == nil) != (err2 == nil) {
		common.Logger.Error().
			Str("action", "bug_detected").
			Int("query_num", queryNum).
			Str("issue", "error_state_differs").
			Msg("BUG: Error state mismatch")
		return true, nil
	}

	// If both failed, that's expected for some queries
	if err != nil && err2 != nil {
		if verbose {
			common.Logger.Debug().
				Int("query_num", queryNum).
				Msg("Both databases failed (expected)")
		}
		return false, nil
	}

	// Compare affected rows
	if tursoAffected != sqlite3Affected {
		common.Logger.Error().
			Str("action", "bug_detected").
			Int("query_num", queryNum).
			Str("issue", "rows_affected_differs").
			Int64("turso_rows", tursoAffected).
			Int64("sqlite3_rows", sqlite3Affected).
			Msg("BUG: Rows affected mismatch")
		return true, nil
	}

	return false, nil
}

func fetchAllRows(rows *sql.Rows) []map[string]interface{} {
	results := []map[string]interface{}{}
	
	cols, err := rows.Columns()
	if err != nil {
		return results
	}

	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]interface{})
		for i, col := range cols {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return results
}

func resultsEqual(a, b []map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	// Simple comparison: convert to string and compare
	// This works for basic types but may need refinement for complex cases
	for i := range a {
		if !rowEqual(a[i], b[i]) {
			return false
		}
	}

	return true
}

func rowEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}

		// Handle nil values
		if v == nil && bv == nil {
			continue
		}
		if v == nil || bv == nil {
			return false
		}

		// Convert to strings for comparison (handles most cases)
		aStr := fmt.Sprintf("%v", v)
		bStr := fmt.Sprintf("%v", bv)
		
		if aStr != bStr {
			return false
		}
	}

	return true
}

func compareAllTables(tursoDb, sqlite3Db *sql.DB, queryNum int, verbose bool) error {
	tursoTables, err := helper.GetAllTablesAndCols(tursoDb, "turso")
	if err != nil {
		return fmt.Errorf("failed to get Turso tables: %w", err)
	}

	for _, table := range tursoTables {
		if err := compareTableData(tursoDb, sqlite3Db, table.Name, queryNum, verbose); err != nil {
			return err
		}
	}

	return nil
}

func compareTableData(tursoDb, sqlite3Db *sql.DB, tableName string, queryNum int, verbose bool) error {
	query := fmt.Sprintf("SELECT * FROM %s ORDER BY rowid", tableName)

	tursoRows, err := tursoDb.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query table %s", tableName)
	}
	defer tursoRows.Close()

	sqlite3Rows, err := sqlite3Db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query table %s", tableName)
	}
	defer sqlite3Rows.Close()

	tursoData := fetchAllRows(tursoRows)
	sqlite3Data := fetchAllRows(sqlite3Rows)

	// Compare row counts
	if len(tursoData) != len(sqlite3Data) {
		common.Logger.Error().
			Str("action", "bug_detected").
			Int("query_num", queryNum).
			Str("table", tableName).
			Str("issue", "row_count_mismatch").
			Int("turso_rows", len(tursoData)).
			Int("sqlite3_rows", len(sqlite3Data)).
			Msg("BUG: Table row count differs")
		return fmt.Errorf("table %s row count mismatch", tableName)
	}

	// Compare data using hash for efficiency
	tursoHash := hashTableData(tursoData)
	sqlite3Hash := hashTableData(sqlite3Data)

	if tursoHash != sqlite3Hash {
		common.Logger.Error().
			Str("action", "bug_detected").
			Int("query_num", queryNum).
			Str("table", tableName).
			Str("issue", "data_hash_mismatch").
			Msg("BUG: Table data differs")
		return fmt.Errorf("table %s data mismatch", tableName)
	}

	return nil
}

func hashTableData(data []map[string]interface{}) string {
	h := sha256.New()
	for _, row := range data {
		// Sort keys for deterministic output
		keys := make([]string, 0, len(row))
		for k := range row {
			keys = append(keys, k)
		}
		// Sort keys alphabetically
		for i := 0; i < len(keys); i++ {
			for j := i + 1; j < len(keys); j++ {
				if keys[i] > keys[j] {
					keys[i], keys[j] = keys[j], keys[i]
				}
			}
		}
		// Build deterministic string from sorted keys
		for _, k := range keys {
			h.Write([]byte(k))
			h.Write([]byte(":"))
			h.Write([]byte(fmt.Sprintf("%v", row[k])))
			h.Write([]byte(";"))
		}
		h.Write([]byte("\n"))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// getEasyStmtWeights returns weights focused on basic SQL operations
func getEasyStmtWeights() map[stmts.StmtType]uint64 {
	w := map[stmts.StmtType]uint64{}
	
	// INSERT variants - high weight
	w[stmts.StmtInsert] = 300
	w[stmts.StmtInsertMultiple] = 100
	w[stmts.StmtInsertOrReplace] = 50
	w[stmts.StmtInsertOrIgnore] = 50
	
	// UPDATE - high weight
	w[stmts.StmtUpdate] = 200
	
	// DELETE - high weight
	w[stmts.StmtDelete] = 150
	
	// Basic SELECT variants - high weight
	w[stmts.StmtSelectBasic] = 250
	w[stmts.StmtSelectWhere] = 200
	w[stmts.StmtSelectLimit] = 100
	w[stmts.StmtSelectOrder] = 100
	w[stmts.StmtSelectGroup] = 80
	
	// More complex SELECT - lower weight
	w[stmts.StmtSelectJoin] = 50
	w[stmts.StmtSelectInner] = 50
	w[stmts.StmtSelectWhereComplex] = 40
	
	// Very low weight or disabled for advanced features and DDL
	w[stmts.StmtPragma] = 0
	w[stmts.StmtCreateTable] = 0
	w[stmts.StmtCreateIndex] = 0
	w[stmts.StmtDropTable] = 0
	w[stmts.StmtAlterTable] = 0
	w[stmts.StmtAnalyze] = 0
	w[stmts.StmtCreateView] = 0
	w[stmts.StmtDropView] = 0
	w[stmts.StmtDropIndex] = 0
	w[stmts.StmtBegin] = 0
	w[stmts.StmtCommit] = 0
	w[stmts.StmtRollback] = 0
	w[stmts.StmtVacuum] = 0
	w[stmts.StmtReindex] = 0
	
	return w
}

// validateFlags validates input flags for security and correctness
func validateFlags(flags *validatorFlags) error {
	// Validate queries count
	if flags.queries <= 0 {
		return errors.New("queries must be a positive number")
	}
	if flags.queries > 1000000 {
		return errors.New("queries exceeds maximum allowed (1000000)")
	}

	// Validate and sanitize init SQL path
	if flags.initSQL == "" {
		return errors.New("init SQL path cannot be empty")
	}
	
	cleanPath := filepath.Clean(flags.initSQL)
	if !filepath.IsAbs(cleanPath) {
		return errors.New("init SQL path must be absolute")
	}

	// Check if file exists and is readable
	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("init SQL file does not exist")
		}
		return errors.New("cannot access init SQL file")
	}

	// Verify it's a regular file
	if !info.Mode().IsRegular() {
		return errors.New("init SQL path must be a regular file")
	}

	// Check file size (prevent extremely large files)
	const maxFileSize = 10 * 1024 * 1024 // 10MB
	if info.Size() > maxFileSize {
		return errors.New("init SQL file exceeds maximum size (10MB)")
	}

	return nil
}

// getQueryType returns a simplified query type for logging (without exposing SQL)
func getQueryType(query string) string {
	trimmed := strings.TrimSpace(strings.ToUpper(query))
	if strings.HasPrefix(trimmed, "SELECT") {
		return "SELECT"
	} else if strings.HasPrefix(trimmed, "INSERT") {
		return "INSERT"
	} else if strings.HasPrefix(trimmed, "UPDATE") {
		return "UPDATE"
	} else if strings.HasPrefix(trimmed, "DELETE") {
		return "DELETE"
	} else if strings.HasPrefix(trimmed, "WITH") {
		return "CTE"
	}
	return "OTHER"
}
