package main

import (
	"database/sql"
	"os"
	"strings"

	"sqlfuse/internal/common"
	"sqlfuse/internal/executors"
	"sqlfuse/internal/generators"
	"sqlfuse/internal/stmts/helper"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

func main() {
	var flags executors.CommonFlags

	rootCmd := &cobra.Command{
		Use:   "go_sqlite3_embedded",
		Short: "Embedded go-sqlite3 executor for SQL fuzzing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executors.Run(
				"Starting go_sqlite3_embedded executor",
				&flags,
				func(dsn string) (*sql.DB, error) { return sql.Open("sqlite3", dsn) },
				func(seed uint64) generators.Generator { return generators.NewGoSQLite3Generator(seed) },
				print_schema,
			)
		},
	}

	executors.AddCommonFlags(rootCmd, &flags, "/opt/assets/go_sqlite3/init.sql")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func print_schema(db *sql.DB) {
	tables, err := helper.GetAllTablesAndCols(db, "go-sqlite3")
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
