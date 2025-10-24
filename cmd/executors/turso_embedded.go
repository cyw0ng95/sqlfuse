package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"sqlsmith-go/internal/generators/turso"
	"sqlsmith-go/internal/generators/turso/helper"
	"strings"
	"sync"

	_ "github.com/tursodatabase/turso-go"
)

func main() {
	conn, err := sql.Open("turso", ":memory:")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// --- INIT DB ---
	initSQL, err := ioutil.ReadFile("/opt/assets/turso/init.sql")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read init.sql: %v\n", err)
		os.Exit(1)
	}
	for _, stmt := range strings.Split(string(initSQL), ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		_, err := conn.Exec(stmt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Init SQL error: %v (stmt: %s)\n", err, stmt)
			os.Exit(1)
		}
	}
	// --- END INIT ---

	// Print all tables and columns discovered after initialization
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
				fmt.Printf("[worker %d] Executing query: %s\n", workerID, query)
				_, err := conn.Exec(query)
				if err != nil {
					errCh <- fmt.Errorf("[worker %d] Error executing query: %v", workerID, err)
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

	fmt.Println("\n---[Summary]---")
	fmt.Println("Total queries executed:", numWorkers*queriesPerWorker)
	fmt.Printf("Total tokens used: %d\n", totalTokens)

	for err := range errCh {
		fmt.Println(err)
	}

	print_schema(conn)
}

func print_schema(db *sql.DB) {
	tables, err := helper.GetAllTablesAndCols(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get tables: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Discovered schema:")
	for _, t := range tables {
		fmt.Printf("- %s: %v\n", t.Name, t.Cols)
	}
}
