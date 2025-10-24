package main

import (
	"database/sql"
	"fmt"
	"os"
	"sqlsmith-go/internal/generators/turso"

	_ "github.com/tursodatabase/turso-go"
)

func main() {
	conn, err := sql.Open("turso", ":memory:")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	for i := 0; i < 100; i++ {
		gen := turso.NewGenerator(uint64(i + 1))
		query := gen.Generate()
		fmt.Printf("Executing query: %s\n", query)
		_, err := conn.Exec(query)
		if err != nil {
			fmt.Printf("Error executing query: %v\n", err)
		}
	}
}
