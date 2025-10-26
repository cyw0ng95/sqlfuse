package main

import (
	"database/sql"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/executors"
	"sqlsmith-go/internal/generators/turso"
	"sqlsmith-go/internal/generators/turso/helper"

	"github.com/spf13/cobra"
	_ "github.com/tursodatabase/turso-go"
)

func main() {
	var flags executors.CommonFlags

	rootCmd := &cobra.Command{
		Use:   "turso_embedded",
		Short: "Embedded Turso/LibSQL executor for SQL fuzzing",
		RunE: func(cmd *cobra.Command, args []string) error {
			common.InitLogger()
			common.Logger.Info().Msg("Starting turso_embedded executor")

			conn, err := sql.Open("turso", flags.Dsn)
			if err != nil {
				common.Logger.Error().Err(err).Msg("Error opening database")
				os.Exit(1)
			}
			defer conn.Close()

			// Initialize schema if provided
			if strings.TrimSpace(flags.InitSQLPath) != "" {
				initSQL, err := ioutil.ReadFile(flags.InitSQLPath)
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

			workers := flags.Workers
			queries := flags.Queries
			if workers < 1 {
				workers = 1
			}
			if queries < 1 {
				queries = 1
			}

			// determine base seed: use provided seed if non-zero, otherwise time-based
			var baseSeed uint64
			if flags.Seed != 0 {
				baseSeed = uint64(flags.Seed)
			} else {
				baseSeed = uint64(time.Now().UnixNano())
			}

			var wg sync.WaitGroup
			wg.Add(workers)
			tokenCh := make(chan uint64, workers)

			for w := 0; w < workers; w++ {
				go func(workerID int) {
					defer wg.Done()
					// each worker gets a deterministic seed derived from base
					gen := turso.NewGenerator(baseSeed + uint64(workerID))
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

	executors.AddCommonFlags(rootCmd, &flags, "/opt/assets/turso/init.sql")

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
