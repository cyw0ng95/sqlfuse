package generators

// GeneratorInfo defines a minimal capability interface for generators.
// Implementations (e.g. turso) should provide a value that satisfies
// this interface so other packages (like the server) can query
// generator metadata without importing generator internals.
type GeneratorInfo interface {
	Name() string
	// SupportedStmts returns a map from stmt type name to its default weight.
	SupportedStmts() map[string]uint64
}
