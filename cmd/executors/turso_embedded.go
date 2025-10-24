package main

import (
	"database/sql"
	"fmt"
	"os"
	"sqlsmith-go/internal/generators/turso"
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

	const (
		numWorkers       = 8
		queriesPerWorker = 100
	)

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	errCh := make(chan error, numWorkers*queriesPerWorker)

	for w := 0; w < numWorkers; w++ {
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < queriesPerWorker; i++ {
				gen := turso.NewGenerator(uint64(workerID*queriesPerWorker + i + 1))
				query := gen.Generate()
				fmt.Printf("[worker %d] Executing query: %s\n", workerID, query)
				_, err := conn.Exec(query)
				if err != nil {
					errCh <- fmt.Errorf("[worker %d] Error executing query: %v", workerID, err)
				}
			}
		}(w)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		fmt.Println(err)
	}
}
