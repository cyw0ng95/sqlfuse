package main

import (
	"database/sql"
	"os"
	"strings"

	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/executors"
	"sqlsmith-go/internal/generators"
	"sqlsmith-go/internal/stmts/helper"

	_ "github.com/marcboeker/go-duckdb"
	"github.com/spf13/cobra"
)

func main() {
	var flags executors.CommonFlags

	rootCmd := &cobra.Command{
		Use:   "duckdb_embedded",
		Short: "Embedded DuckDB executor for SQL fuzzing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executors.Run(
				"Starting duckdb_embedded executor",
				&flags,
				func(dsn string) (*sql.DB, error) {
					if dsn == ":memory:" {
						dsn = ""
					}
					return sql.Open("duckdb", dsn)
				},
				func(seed uint64) generators.Generator { return generators.NewDuckDBGenerator(seed) },
				print_schema,
			)
		},
	}

	executors.AddCommonFlags(rootCmd, &flags, "/opt/assets/duckdb/init.sql")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func print_schema(db *sql.DB) {
	tables, err := helper.GetAllTablesAndCols(db, "duckdb")
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
