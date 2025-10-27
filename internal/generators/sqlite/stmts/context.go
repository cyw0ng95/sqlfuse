package stmts

import (
	"database/sql"
	"sqlsmith-go/internal/common"
)

// FlavorConfig defines SQL dialect-specific behavior and constraints.
// Different SQL flavors (Turso, go-sqlite3, PostgreSQL, MySQL, etc.) can
// implement this interface to customize statement generation.
type FlavorConfig interface {
	// Name returns the SQL flavor name (e.g., "turso", "sqlite3", "postgres")
	Name() string

	// SupportsFeature checks if a specific SQL feature is supported by this flavor.
	// Feature names are standardized strings like "window_functions", "cte_recursive",
	// "exists_subquery", "regexp", etc.
	SupportsFeature(feature string) bool

	// ValidateSQL optionally validates SQL syntax for this flavor.
	// Returns nil if SQL is valid or validation is not implemented.
	ValidateSQL(sql string) error
}

// DefaultFlavorConfig provides a permissive default that supports all SQLite features.
type DefaultFlavorConfig struct{}

func (d *DefaultFlavorConfig) Name() string {
	return "sqlite"
}

func (d *DefaultFlavorConfig) SupportsFeature(feature string) bool {
	// Default SQLite supports all standard features
	return true
}

func (d *DefaultFlavorConfig) ValidateSQL(sql string) error {
	// No validation by default
	return nil
}

// GenContext provides context for recursive SQL generation.
// It tracks recursion depth to prevent infinite recursion and provides
// access to the database connection and random number generator.
type GenContext struct {
	DB       *sql.DB
	LCG      *common.LCG
	Depth    int          // Current recursion depth
	MaxDepth int          // Maximum allowed recursion depth
	Flavor   FlavorConfig // SQL dialect configuration
}

// NewGenContext creates a new generation context with the given database and LCG.
// maxDepth controls how deep recursive generation can go (0 = no recursion).
func NewGenContext(db *sql.DB, lcg *common.LCG, maxDepth int) *GenContext {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}
	return &GenContext{
		DB:       db,
		LCG:      lcg,
		Depth:    0,
		MaxDepth: maxDepth,
		Flavor:   &DefaultFlavorConfig{},
	}
}

// NewGenContextWithFlavor creates a new generation context with a specific SQL flavor.
func NewGenContextWithFlavor(db *sql.DB, lcg *common.LCG, maxDepth int, flavor FlavorConfig) *GenContext {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}
	if flavor == nil {
		flavor = &DefaultFlavorConfig{}
	}
	return &GenContext{
		DB:       db,
		LCG:      lcg,
		Depth:    0,
		MaxDepth: maxDepth,
		Flavor:   flavor,
	}
}

// CanRecurse returns true if we haven't reached the maximum recursion depth.
func (ctx *GenContext) CanRecurse() bool {
	return ctx.Depth < ctx.MaxDepth
}

// Descend creates a new context with incremented depth for recursive calls.
func (ctx *GenContext) Descend() *GenContext {
	return &GenContext{
		DB:       ctx.DB,
		LCG:      ctx.LCG,
		Depth:    ctx.Depth + 1,
		MaxDepth: ctx.MaxDepth,
		Flavor:   ctx.Flavor,
	}
}

// Intn is a convenience method to get a random integer from the LCG.
func (ctx *GenContext) Intn(n int) int {
	if ctx.LCG != nil {
		return ctx.LCG.Intn(n)
	}
	return 0
}

// SupportsFeature is a convenience method to check if the current flavor supports a feature.
func (ctx *GenContext) SupportsFeature(feature string) bool {
	if ctx.Flavor != nil {
		return ctx.Flavor.SupportsFeature(feature)
	}
	return true // Default to supporting everything if no flavor is set
}
