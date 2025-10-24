package main

import (
	"database/sql"
	"io/ioutil"
	"os"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/generators/turso"
	"sqlsmith-go/internal/generators/turso/helper"
	"strings"
	"sync"

	_ "github.com/tursodatabase/turso-go"
)

func main() {
	common.InitLogger()
	common.Logger.Info().Msg("Starting turso_embedded executor")

	conn, err := sql.Open("turso", ":memory:")
	if err != nil {
		common.Logger.Error().Err(err).Msg("Error opening database")
		os.Exit(1)
	}
	defer conn.Close()

	// --- INIT DB ---
	initSQL, err := ioutil.ReadFile("/opt/assets/turso/init.sql")
	if err != nil {
		common.Logger.Error().Err(err).Msg("Failed to read init.sql")
		os.Exit(1)
	}
	common.Logger.Info().Msg("Initializing database schema from init.sql")
	for _, stmt := range strings.Split(string(initSQL), ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		_, err := conn.Exec(stmt)
		if err != nil {
			common.Logger.Error().Err(err).Str("stmt", stmt).Msg("Init SQL error")
			os.Exit(1)
		}
	}
	// --- END INIT ---

	print_schema(conn)

	const (
		numWorkers       = 1
		queriesPerWorker = 10
	)

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	errCh := make(chan error, numWorkers*queriesPerWorker)
	tokenCh := make(chan uint64, numWorkers)

	for w := 0; w < numWorkers; w++ {
		go func(workerID int) {
			defer wg.Done()
			gen := turso.NewGenerator(uint64(workerID + 1))
			for i := 0; i < queriesPerWorker; i++ {
				query := gen.GenerateWithDB(conn)
				common.Logger.Info().Int("worker", workerID).Int("query_num", i+1).Str("query", query).Msg("Executing query")
				_, err := conn.Exec(query)
				if err != nil {
					errCh <- err
				}
			}
			tokenCh <- gen.TokensUsed()
		}(w)
	}

	wg.Wait()
	close(errCh)
	close(tokenCh)

	totalTokens := uint64(0)
	for tokens := range tokenCh {
		totalTokens += tokens
	}

	common.Logger.Info().Int("total_queries", numWorkers*queriesPerWorker).Uint64("total_tokens", totalTokens).Msg("Summary of execution")

	for err := range errCh {
		common.Logger.Error().Err(err).Msg("Query execution error")
	}

	print_schema(conn)
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
