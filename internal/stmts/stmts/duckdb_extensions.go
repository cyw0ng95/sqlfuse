package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
)

// CreateSecretGenerator is a StmtGenerator for CREATE SECRET statements.
type CreateSecretGenerator struct{}

// Generate implements StmtGenerator for CREATE SECRET statements.
func (g *CreateSecretGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenCreateSecret(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. CREATE SECRET can always be generated.
func (g *CreateSecretGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// CreateSecretStmt represents a CREATE SECRET statement.
type CreateSecretStmt struct {
	*BaseStmt
}

// GenCreateSecret generates a CREATE SECRET statement.
// Secrets store credentials for external services.
// Reference: https://duckdb.org/docs/stable/sql/statements/create_secret
func GenCreateSecret(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	var sql string
	choice := lcg.Intn(4)

	secretName := fmt.Sprintf("secret_%d", lcg.Uint64()%1000)

	if choice == 0 {
		// S3 secret
		sql = fmt.Sprintf(`CREATE SECRET %s (
    TYPE S3,
    KEY_ID 'access_key_id',
    SECRET 'secret_access_key',
    REGION 'us-east-1'
);`, secretName)
	} else if choice == 1 {
		// GCS secret
		sql = fmt.Sprintf(`CREATE SECRET %s (
    TYPE GCS,
    KEY_ID 'key_id',
    SECRET 'secret'
);`, secretName)
	} else if choice == 2 {
		// Bearer token secret
		sql = fmt.Sprintf(`CREATE SECRET %s (
    TYPE BEARER,
    TOKEN 'bearer_token_value'
);`, secretName)
	} else {
		// Generic credential secret
		sql = fmt.Sprintf(`CREATE SECRET %s (
    TYPE CREDENTIAL,
    PROVIDER 'config',
    CREDENTIAL 'my_credential'
);`, secretName)
	}

	return &CreateSecretStmt{
		BaseStmt: NewBaseStmt(sql, "create_secret", flavor),
	}, nil
}

// DropSecretGenerator is a StmtGenerator for DROP SECRET statements.
type DropSecretGenerator struct{}

// Generate implements StmtGenerator for DROP SECRET statements.
func (g *DropSecretGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenDropSecret(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. DROP SECRET can always be generated.
func (g *DropSecretGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// DropSecretStmt represents a DROP SECRET statement.
type DropSecretStmt struct {
	*BaseStmt
}

// GenDropSecret generates a DROP SECRET statement.
func GenDropSecret(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	secretName := fmt.Sprintf("secret_%d", lcg.Uint64()%1000)
	
	var sql string
	if lcg.Intn(2) == 0 {
		sql = fmt.Sprintf("DROP SECRET %s;", secretName)
	} else {
		sql = fmt.Sprintf("DROP SECRET IF EXISTS %s;", secretName)
	}

	return &DropSecretStmt{
		BaseStmt: NewBaseStmt(sql, "drop_secret", flavor),
	}, nil
}

// LoadInstallGenerator is a StmtGenerator for LOAD/INSTALL statements.
type LoadInstallGenerator struct{}

// Generate implements StmtGenerator for LOAD/INSTALL statements.
func (g *LoadInstallGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenLoadInstall(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. LOAD/INSTALL can always be generated.
func (g *LoadInstallGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// LoadInstallStmt represents a LOAD or INSTALL statement.
type LoadInstallStmt struct {
	*BaseStmt
}

// GenLoadInstall generates a LOAD or INSTALL statement.
// Reference: https://duckdb.org/docs/stable/sql/statements/load_and_install
func GenLoadInstall(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	var sql string
	
	// Common DuckDB extensions
	extensions := []string{
		"httpfs",
		"json",
		"parquet",
		"fts",
		"icu",
		"inet",
		"spatial",
		"sqlite",
	}

	ext := extensions[lcg.Intn(len(extensions))]

	if lcg.Intn(2) == 0 {
		// INSTALL extension
		sql = fmt.Sprintf("INSTALL %s;", ext)
	} else {
		// LOAD extension
		sql = fmt.Sprintf("LOAD %s;", ext)
	}

	return &LoadInstallStmt{
		BaseStmt: NewBaseStmt(sql, "load_install", flavor),
	}, nil
}

// CommentOnGenerator is a StmtGenerator for COMMENT ON statements.
type CommentOnGenerator struct{}

// Generate implements StmtGenerator for COMMENT ON statements.
func (g *CommentOnGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenCommentOn(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. COMMENT ON can always be generated.
func (g *CommentOnGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// CommentOnStmt represents a COMMENT ON statement.
type CommentOnStmt struct {
	*BaseStmt
}

// GenCommentOn generates a COMMENT ON statement.
// Reference: https://duckdb.org/docs/stable/sql/statements/comment_on
func GenCommentOn(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	var sql string
	choice := lcg.Intn(6)

	switch choice {
	case 0:
		// COMMENT ON TABLE
		tableName := fmt.Sprintf("table_%d", lcg.Uint64()%1000)
		sql = fmt.Sprintf("COMMENT ON TABLE %s IS 'This is a test table';", tableName)
	case 1:
		// COMMENT ON COLUMN
		tableName := fmt.Sprintf("table_%d", lcg.Uint64()%1000)
		colName := fmt.Sprintf("col_%d", lcg.Uint64()%100)
		sql = fmt.Sprintf("COMMENT ON COLUMN %s.%s IS 'This is a test column';", tableName, colName)
	case 2:
		// COMMENT ON VIEW
		viewName := fmt.Sprintf("view_%d", lcg.Uint64()%1000)
		sql = fmt.Sprintf("COMMENT ON VIEW %s IS 'This is a test view';", viewName)
	case 3:
		// COMMENT ON INDEX
		indexName := fmt.Sprintf("idx_%d", lcg.Uint64()%1000)
		sql = fmt.Sprintf("COMMENT ON INDEX %s IS 'This is a test index';", indexName)
	case 4:
		// COMMENT ON SCHEMA
		schemaName := fmt.Sprintf("schema_%d", lcg.Uint64()%100)
		sql = fmt.Sprintf("COMMENT ON SCHEMA %s IS 'This is a test schema';", schemaName)
	default:
		// COMMENT ON TYPE
		typeName := fmt.Sprintf("type_%d", lcg.Uint64()%100)
		sql = fmt.Sprintf("COMMENT ON TYPE %s IS 'This is a test type';", typeName)
	}

	return &CommentOnStmt{
		BaseStmt: NewBaseStmt(sql, "comment_on", flavor),
	}, nil
}

// ProfilingGenerator is a StmtGenerator for profiling statements.
type ProfilingGenerator struct{}

// Generate implements StmtGenerator for profiling statements.
func (g *ProfilingGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenProfiling(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. Profiling can always be generated.
func (g *ProfilingGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// ProfilingStmt represents a profiling statement.
type ProfilingStmt struct {
	*BaseStmt
}

// GenProfiling generates profiling-related statements.
// Reference: https://duckdb.org/docs/stable/sql/statements/profiling
func GenProfiling(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	var sql string
	choice := lcg.Intn(4)

	switch choice {
	case 0:
		// Enable profiling
		sql = "PRAGMA enable_profiling;"
	case 1:
		// Disable profiling
		sql = "PRAGMA disable_profiling;"
	case 2:
		// Get profiling output
		sql = "PRAGMA profiling_output;"
	default:
		// Set profiling mode
		modes := []string{"'standard'", "'detailed'", "'json'"}
		mode := modes[lcg.Intn(len(modes))]
		sql = fmt.Sprintf("PRAGMA profiling_mode = %s;", mode)
	}

	return &ProfilingStmt{
		BaseStmt: NewBaseStmt(sql, "profiling", flavor),
	}, nil
}

// SetVariableGenerator is a StmtGenerator for SET VARIABLE statements.
type SetVariableGenerator struct{}

// Generate implements StmtGenerator for SET VARIABLE statements.
func (g *SetVariableGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenSetVariable(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. SET VARIABLE can always be generated.
func (g *SetVariableGenerator) CanGenerate(ctx *GenContext) bool {
	return true
}

// SetVariableStmt represents a SET VARIABLE statement.
type SetVariableStmt struct {
	*BaseStmt
}

// GenSetVariable generates a SET VARIABLE statement.
// Reference: https://duckdb.org/docs/stable/sql/statements/set_variable
func GenSetVariable(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	varName := fmt.Sprintf("var_%d", lcg.Uint64()%100)
	
	var sql string
	choice := lcg.Intn(4)

	switch choice {
	case 0:
		// Integer variable
		value := lcg.Uint64() % 1000
		sql = fmt.Sprintf("SET VARIABLE %s = %d;", varName, value)
	case 1:
		// String variable
		sql = fmt.Sprintf("SET VARIABLE %s = 'test_value';", varName)
	case 2:
		// Boolean variable
		value := "true"
		if lcg.Intn(2) == 0 {
			value = "false"
		}
		sql = fmt.Sprintf("SET VARIABLE %s = %s;", varName, value)
	default:
		// Expression variable
		sql = fmt.Sprintf("SET VARIABLE %s = 1 + 2 * 3;", varName)
	}

	return &SetVariableStmt{
		BaseStmt: NewBaseStmt(sql, "set_variable", flavor),
	}, nil
}
