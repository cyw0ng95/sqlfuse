package main

import (
	"database/sql"
	"os"
	"strings"

	"sqlfuse/internal/common"
	"sqlfuse/internal/executors"
	"sqlfuse/internal/generators"
	"sqlfuse/internal/oracles"
	"sqlfuse/internal/stmts/helper"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

func main() {
	var flags executors.CommonFlags
	var enableOracles bool

	rootCmd := &cobra.Command{
		Use:   "go_sqlite3_oracle_executor",
		Short: "go-sqlite3 executor with SQLancer-inspired oracle testing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithOracles(&flags, enableOracles)
		},
	}

	executors.AddCommonFlags(rootCmd, &flags, "/opt/assets/go_sqlite3/init.sql")
	rootCmd.Flags().BoolVar(&enableOracles, "oracles", false, "Enable SQLancer-inspired oracle testing (TLP, NoREC, PQS, QPG)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runWithOracles(flags *executors.CommonFlags, enableOracles bool) error {
	common.InitLogger()
	common.Logger.Info().Msg("Starting go_sqlite3_oracle_executor with oracle testing")

	conn, err := sql.Open("sqlite3", flags.Dsn)
	if err != nil {
		common.Logger.Error().Err(err).Msg("Error opening database")
		os.Exit(1)
	}
	defer conn.Close()

	// Initialize schema if provided
	if strings.TrimSpace(flags.InitSQLPath) != "" {
		initSQL, err := os.ReadFile(flags.InitSQLPath)
		if err != nil {
			common.Logger.Error().Err(err).Str("path", flags.InitSQLPath).Msg("Failed to read init SQL file")
			os.Exit(1)
		}
		common.Logger.Info().Msgf("Initializing database schema from %s", flags.InitSQLPath)
		for _, stmt := range strings.Split(string(initSQL), ";") {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := conn.Exec(stmt); err != nil {
				common.Logger.Error().Err(err).Str("stmt", stmt).Msg("Init SQL error")
				os.Exit(1)
			}
		}
	}

	print_schema(conn)

	// Setup generator
	gen := generators.NewGoSQLite3Generator(uint64(flags.Seed))

	// Setup oracles if enabled
	if enableOracles {
		common.Logger.Info().Msg("Oracle testing enabled")
		gen.EnableOracles(true)

		// Get table schema for oracles
		tables, err := helper.GetAllTablesAndCols(conn, "go-sqlite3")
		if err != nil {
			common.Logger.Warn().Err(err).Msg("Failed to get schema for oracles")
		} else {
			// Create oracles for each table
			lcg := common.NewLCG(uint64(flags.Seed))
			
			for _, table := range tables {
				if len(table.Cols) == 0 {
					continue
				}
				
				colNames := make([]string, len(table.Cols))
				for i, col := range table.Cols {
					colNames[i] = col.Name
				}

				// Add TLP oracle
				tlp := oracles.NewTLPOracle(conn, lcg, table.Name, colNames)
				gen.AddOracle(tlp)

				// Add NoREC oracle
				norec := oracles.NewNoRECOracle(conn, lcg, table.Name, colNames)
				gen.AddOracle(norec)

				// Add PQS oracle
				pqs := oracles.NewPQSOracle(conn, lcg, table.Name, colNames)
				gen.AddOracle(pqs)

				// Add QPG oracle
				qpg := oracles.NewQPGOracle(conn, lcg, table.Name, colNames)
				gen.AddOracle(qpg)

				common.Logger.Info().Msgf("Added oracles for table: %s", table.Name)
			}
		}
	}

	// Execute queries
	workers := flags.Workers
	if workers < 1 {
		workers = 1
	}
	queries := flags.Queries
	if queries < 1 {
		queries = 1
	}

	totalBugs := 0
	totalChecks := 0

	for i := 0; i < queries; i++ {
		query := gen.GenerateWithDB(conn)
		_, execErr := conn.Exec(query)
		
		if flags.Verbose {
			common.Logger.Info().Msgf("Executing query %d: %s", i+1, query)
		}
		
		if execErr != nil && flags.Verbose {
			common.Logger.Info().Msgf("Query error (expected): %v", execErr)
		}

		// Run oracles if enabled
		if enableOracles {
			errors := gen.RunOracles()
			totalChecks += len(gen.GetOracles())
			
			if len(errors) > 0 {
				totalBugs += len(errors)
				common.Logger.Warn().Msgf("Oracle iteration %d found %d potential bugs:", i+1, len(errors))
				for _, err := range errors {
					common.Logger.Warn().Msgf("  - %v", err)
				}
			}

			// Check QPG coverage periodically
			if (i+1)%20 == 0 {
				for _, oracle := range gen.GetOracles() {
					if qpg, ok := oracle.(*oracles.QPGOracle); ok {
						uniquePlans, queriesSinceNew := qpg.GetPlanCoverage()
						common.Logger.Info().Msgf("QPG Coverage: %d unique plans, %d queries since new plan",
							uniquePlans, queriesSinceNew)
					}
				}
			}
		}
	}

	if enableOracles {
		common.Logger.Info().Msgf("Oracle Testing Summary:")
		common.Logger.Info().Msgf("  Total checks: %d", totalChecks)
		common.Logger.Info().Msgf("  Bugs found: %d", totalBugs)
		if totalChecks > 0 {
			bugRate := float64(totalBugs) / float64(totalChecks) * 100
			common.Logger.Info().Msgf("  Bug rate: %.2f%%", bugRate)
		}
	}

	print_schema(conn)
	common.Logger.Info().Msg("Execution complete")
	return nil
}

func print_schema(db *sql.DB) {
	tables, err := helper.GetAllTablesAndCols(db, "go-sqlite3")
	if err != nil {
		common.Logger.Error().Err(err).Msg("Failed to get tables")
		return
	}
	var b strings.Builder
	b.WriteString("Discovered schema:\n")
	for _, t := range tables {
		colNames := make([]string, len(t.Cols))
		for i, c := range t.Cols {
			colNames[i] = c.Name
		}
		b.WriteString("  ")
		b.WriteString(t.Name)
		b.WriteString(": ")
		b.WriteString(strings.Join(colNames, ", "))
		b.WriteString("\n")
	}
	common.Logger.Info().Msg(b.String())
}
