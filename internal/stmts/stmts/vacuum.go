package stmts

import (
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
)

// VacuumGenerator is a StmtGenerator for VACUUM statements.
type VacuumGenerator struct{}

// Generate implements StmtGenerator for VACUUM statements.
func (g *VacuumGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenVacuum(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. VACUUM can always be generated.
func (g *VacuumGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// VacuumStmt represents a VACUUM statement.
// It embeds BaseStmt to avoid boilerplate method implementations.
type VacuumStmt struct {
	*BaseStmt
}

// GenVacuum generates a VACUUM statement.
// VACUUM rebuilds the database file, repacking it into a minimal amount of disk space.
// According to SQLite documentation: https://sqlite.org/lang_vacuum.html
// Turso COMPAT.md: Partial support (VACUUM only, no VACUUM INTO).
func GenVacuum(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = EnsureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	// VACUUM can be used in several ways:
	// 1. VACUUM; (vacuums the main database)
	// 2. VACUUM schema_name; (vacuums a specific schema)
	// 3. VACUUM INTO 'filename'; (vacuums into a new file - not supported by Turso)

	var sql string

	// Check if the flavor supports VACUUM INTO
	if flavor.Name() == "go-sqlite3" && lcg.Intn(5) == 0 {
		// 20% chance for go-sqlite3: VACUUM INTO (full SQLite3 feature)
		filename := fmt.Sprintf("/tmp/vacuum_%d.db", lcg.Uint64()%100000)
		sql = fmt.Sprintf("VACUUM INTO '%s';", filename)
	} else {
		// Standard VACUUM
		choice := lcg.Intn(10)
		if choice < 8 {
			// 80% chance: VACUUM all
			sql = "VACUUM;"
		} else {
			// 20% chance: VACUUM a specific schema (if attached databases exist)
			// For simplicity, we'll just use main schema
			sql = "VACUUM main;"
		}
	}

	return &VacuumStmt{
		BaseStmt: NewBaseStmt(sql, "vacuum", flavor),
	}, nil
}
