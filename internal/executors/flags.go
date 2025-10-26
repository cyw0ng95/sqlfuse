package executors

import (
	"github.com/spf13/cobra"
)

// CommonFlags holds flags common to executor binaries.
type CommonFlags struct {
	Dsn         string
	InitSQLPath string
	Workers     int
	Queries     int
	Seed        int64
}

// AddCommonFlags registers the standard executor flags onto the provided cobra command,
// storing values into the supplied CommonFlags struct.
func AddCommonFlags(cmd *cobra.Command, f *CommonFlags) {
	cmd.Flags().StringVarP(&f.Dsn, "dsn", "d", ":memory:", "Database DSN for the turso driver (e.g., :memory: or file path)")
	cmd.Flags().StringVarP(&f.InitSQLPath, "init-sql", "i", "/opt/assets/turso/init.sql", "Path to SQL file to initialize schema; set empty to skip")
	cmd.Flags().IntVarP(&f.Workers, "workers", "w", 1, "Number of concurrent workers")
	cmd.Flags().IntVarP(&f.Queries, "queries", "q", 10, "Number of queries per worker")
	cmd.Flags().Int64VarP(&f.Seed, "seed", "s", 0, "Seed for fuzzing (0 means random)")
}
