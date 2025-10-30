package stmts

import (
	"database/sql"
	"fmt"
	"sqlsmith-go/internal/common"
	"sqlsmith-go/internal/stmts/helper"
)

// CopyGenerator is a StmtGenerator for COPY statements.
type CopyGenerator struct{}

// Generate implements StmtGenerator for COPY statements.
func (g *CopyGenerator) Generate(ctx *GenContext) (Stmt, error) {
	return GenCopy(ctx.DB, ctx.LCG, ctx.Flavor)
}

// CanGenerate implements StmtGenerator. COPY requires tables to exist for export.
func (g *CopyGenerator) CanGenerate(ctx *GenContext) bool {
	// COPY can work for both import (COPY TO) and export (COPY FROM)
	// For export, we need tables
	return hasTables(ctx.DB)
}

// CopyStmt represents a COPY statement.
// It embeds BaseStmt to avoid boilerplate method implementations.
type CopyStmt struct {
	*BaseStmt
}

// GenCopy generates a COPY statement.
// COPY is DuckDB's primary mechanism for importing/exporting data.
// Reference: https://duckdb.org/docs/stable/sql/statements/copy
func GenCopy(db *sql.DB, lcg *common.LCG, flavor FlavorConfig) (Stmt, error) {
	lcg = ensureLCG(lcg)
	if flavor == nil {
		flavor = GetDefaultFlavor()
	}

	// COPY has two main forms:
	// 1. COPY table_name TO 'filename' (export)
	// 2. COPY table_name FROM 'filename' (import)
	// 3. COPY (SELECT ...) TO 'filename' (export query result)

	var sql string
	choice := lcg.Intn(10)

	if choice < 7 {
		// 70% chance: COPY table TO file (export)
		var tables []helper.TableInfo
		var err error
		if db != nil {
			tables, err = helper.GetAllTablesAndCols(db)
		}
		if err != nil || len(tables) == 0 {
			// Fallback to a simple CSV export
			sql = "COPY (SELECT 1 AS col) TO '/tmp/export.csv' (FORMAT CSV, HEADER);"
		} else {
			tbl := tables[lcg.Intn(len(tables))]
			format := []string{"CSV", "PARQUET", "JSON"}[lcg.Intn(3)]
			filename := fmt.Sprintf("/tmp/export_%s.%s", tbl.Name, 
				map[string]string{"CSV": "csv", "PARQUET": "parquet", "JSON": "json"}[format])
			
			if format == "CSV" {
				sql = fmt.Sprintf("COPY \"%s\" TO '%s' (FORMAT CSV, HEADER);", tbl.Name, filename)
			} else {
				sql = fmt.Sprintf("COPY \"%s\" TO '%s' (FORMAT %s);", tbl.Name, filename, format)
			}
		}
	} else if choice < 9 {
		// 20% chance: COPY query result TO file
		format := []string{"CSV", "PARQUET", "JSON"}[lcg.Intn(3)]
		filename := fmt.Sprintf("/tmp/query_export.%s",
			map[string]string{"CSV": "csv", "PARQUET": "parquet", "JSON": "json"}[format])
		
		if format == "CSV" {
			sql = fmt.Sprintf("COPY (SELECT 1 AS id, 'test' AS name) TO '%s' (FORMAT CSV, HEADER);", filename)
		} else {
			sql = fmt.Sprintf("COPY (SELECT 1 AS id, 'test' AS name) TO '%s' (FORMAT %s);", filename, format)
		}
	} else {
		// 10% chance: COPY FROM file (import)
		var tables []helper.TableInfo
		var err error
		if db != nil {
			tables, err = helper.GetAllTablesAndCols(db)
		}
		if err != nil || len(tables) == 0 {
			// Fallback to creating a temporary table name
			sql = "COPY temp_import FROM '/tmp/import.csv' (FORMAT CSV, HEADER);"
		} else {
			tbl := tables[lcg.Intn(len(tables))]
			format := []string{"CSV", "PARQUET", "JSON"}[lcg.Intn(3)]
			filename := fmt.Sprintf("/tmp/import.%s",
				map[string]string{"CSV": "csv", "PARQUET": "parquet", "JSON": "json"}[format])
			
			if format == "CSV" {
				sql = fmt.Sprintf("COPY \"%s\" FROM '%s' (FORMAT CSV, HEADER);", tbl.Name, filename)
			} else {
				sql = fmt.Sprintf("COPY \"%s\" FROM '%s' (FORMAT %s);", tbl.Name, filename, format)
			}
		}
	}

	return &CopyStmt{
		BaseStmt: NewBaseStmt(sql, "copy", flavor),
	}, nil
}
