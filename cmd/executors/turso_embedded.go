package main

import (
	"database/sql"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/generators/turso"
	"sqlsmith-go/internal/generators/turso/helper"

	"github.com/spf13/cobra"
	_ "github.com/tursodatabase/turso-go"
)

var (
	dsn         string
	initSQLPath string
	workers     int
	queries     int
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "turso_embedded",
		Short: "Embedded Turso/LibSQL executor for SQL fuzzing",
		RunE: func(cmd *cobra.Command, args []string) error {
			common.InitLogger()
			common.Logger.Info().Msg("Starting turso_embedded executor")

			conn, err := sql.Open("turso", dsn)
			if err != nil {
				common.Logger.Error().Err(err).Msg("Error opening database")
				os.Exit(1)
			}
			defer conn.Close()

			// Initialize schema if provided
			if strings.TrimSpace(initSQLPath) != "" {
				initSQL, err := ioutil.ReadFile(initSQLPath)
				if err != nil {
					common.Logger.Error().Err(err).Str("path", initSQLPath).Msg("Failed to read init SQL file")
					os.Exit(1)
				}
				common.Logger.Info().Msgf("Initializing database schema from %s", initSQLPath)
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

			if workers < 1 {
				workers = 1
			}
			if queries < 1 {
				queries = 1
			}

			var wg sync.WaitGroup
			wg.Add(workers)
			tokenCh := make(chan uint64, workers)

			for w := 0; w < workers; w++ {
				go func(workerID int) {
					defer wg.Done()
					gen := turso.NewGenerator(uint64(workerID + 1))
					for i := 0; i < queries; i++ {
						query := gen.GenerateWithDB(conn)
						_, execErr := conn.Exec(query)
						if execErr != nil {
							common.Logger.Info().Msgf("Worker %d executing query %d: %s", workerID, i+1, query)
							continue
						}
						common.Logger.Info().Msgf("Worker %d executing query %d: %s", workerID, i+1, query)
					}
					tokenCh <- gen.TokensUsed()
				}(w)
			}

			wg.Wait()
			close(tokenCh)

			var totalTokens uint64
			for t := range tokenCh {
				totalTokens += t
			}

			common.Logger.Info().Msgf("Total queries executed: %d", workers*queries)
			common.Logger.Info().Msgf("Total tokens used: %d", totalTokens)

			print_schema(conn)
			return nil
		},
	}

	rootCmd.Flags().StringVarP(&dsn, "dsn", "d", ":memory:", "Database DSN for the turso driver (e.g., :memory: or file path)")
	rootCmd.Flags().StringVarP(&initSQLPath, "init-sql", "i", "/opt/assets/turso/init.sql", "Path to SQL file to initialize schema; set empty to skip")
	rootCmd.Flags().IntVarP(&workers, "workers", "w", 1, "Number of concurrent workers")
	rootCmd.Flags().IntVarP(&queries, "queries", "q", 10, "Number of queries per worker")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func print_schema(db *sql.DB) {
	tables, err := helper.GetAllTablesAndCols(db)
	if err != nil {
		common.Logger.Error().Err(err).Msg("Failed to get tables")
		os.Exit(1)
	}
	var b strings.Builder
	b.WriteString("Discovered schema after execution:\n")
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
