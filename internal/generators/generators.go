package generators

import "database/sql"

// Generator is the primary interface that executors use to invoke generators.
// It combines SQL generation capabilities with metadata queries.
// Implementations should embed BaseGenerator for common functionality.
type Generator interface {
	// GenerateWithDB produces a single SQL statement. If db is non-nil the generator
	// may query schema information to produce richer SELECTs.
	GenerateWithDB(db *sql.DB) string

	// TokensUsed returns the number of pseudo-random tokens consumed by the
	// generator's PRNG. Executors can aggregate this across workers.
	TokensUsed() uint64

	// Name returns the generator's identifier (e.g., "turso").
	Name() string

	// SupportedStmts returns a map from stmt type name to its default weight.
	// This allows querying generator capabilities without importing internals.
	SupportedStmts() map[string]uint64
}
