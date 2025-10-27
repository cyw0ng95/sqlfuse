package generators

import "database/sql"

// GeneratorInfo defines a minimal capability interface for generators.
// Implementations (e.g. turso) should provide a value that satisfies
// this interface so other packages (like the server) can query
// generator metadata without importing generator internals.
type GeneratorInfo interface {
	Name() string
	// SupportedStmts returns a map from stmt type name to its default weight.
	SupportedStmts() map[string]uint64
}

// Generator is a lightweight runtime interface that executors use to invoke
// generators without depending on concrete implementations. Executors require
// at minimum the ability to produce SQL given a DB and to report token usage.
// Implementations should keep this surface small to avoid coupling.
type Generator interface {
	// GenerateWithDB produces a single SQL statement. If db is non-nil the generator
	// may query schema information to produce richer SELECTs.
	GenerateWithDB(db *sql.DB) string

	// TokensUsed returns the number of pseudo-random tokens consumed by the
	// generator's PRNG. Executors can aggregate this across workers.
	TokensUsed() uint64
}
