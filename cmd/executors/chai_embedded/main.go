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
	"sqlsmith-go/internal/generators"
	"sqlsmith-go/internal/generators/turso"

	// _ "github.com/chaisql/chai"
	"github.com/spf13/cobra"
)

// chai_embedded: uses the turso generator to produce SQL and executes them against
// an embedded chai database driver.
func main() {
	var flags executors.CommonFlags

	rootCmd := &cobra.Command{
		Use:   "chai_embedded",
		Short: "Run generated SQL against chaisql/chai database",
		RunE: func(cmd *cobra.Command, args []string) error {
			common.InitLogger()
			common.Logger.Info().Msg("Starting chai_embedded executor")

			// open chai database using provided DSN
			db, err := sql.Open("chai", flags.Dsn)
			if err != nil {
				common.Logger.Error().Err(err).Msg("Error opening chai database")
				os.Exit(1)
			}
			defer db.Close()

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
					if _, err := db.Exec(stmt); err != nil {
						common.Logger.Error().Err(err).Str("stmt", stmt).Msg("Init SQL error")
						os.Exit(1)
					}
				}
			}

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
					var gen generators.Generator = turso.NewGenerator(baseSeed + uint64(workerID))
					for i := 0; i < queries; i++ {
						query := gen.GenerateWithDB(db)
						if _, execErr := db.Exec(query); execErr != nil {
							if flags.Verbose {
								common.Logger.Info().Msgf("Worker %d executing query %d: %s", workerID, i+1, query)
								common.Logger.Info().Msgf("Execution error: %v", execErr)
							}
							continue
						}
						if flags.Verbose {
							common.Logger.Info().Msgf("Worker %d executed query %d", workerID, i+1)
						}
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

			return nil
		},
	}

	executors.AddCommonFlags(rootCmd, &flags, "/opt/assets/chai/init.sql")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
