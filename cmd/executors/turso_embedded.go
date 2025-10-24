package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/tursodatabase/turso-go"
)

func main() {
	conn, err := sql.Open("turso", ":memory:")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	sql := "CREATE table go_turso (foo INTEGER, bar TEXT)"
	_, _ = conn.Exec(sql)

	sql = "INSERT INTO go_turso (foo, bar) values (?, ?)"
	stmt, _ := conn.Prepare(sql)
	defer stmt.Close()
	_, _ = stmt.Exec(42, "turso")
	rows, _ := conn.Query("SELECT * from go_turso")
	defer rows.Close()
	for rows.Next() {
		var a int
		var b string
		_ = rows.Scan(&a, &b)
		fmt.Printf("%d, %s", a, b)
	}
}
