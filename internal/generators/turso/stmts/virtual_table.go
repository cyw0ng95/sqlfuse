package stmts

import (
	"fmt"
	"sqlsmith-go/internal/common"
	"strings"
)

// CreateVirtualTableStmt represents a CREATE VIRTUAL TABLE statement.
type CreateVirtualTableStmt struct {
	sql string
}

func (s *CreateVirtualTableStmt) SQL() string  { return s.sql }
func (s *CreateVirtualTableStmt) Type() string { return "create_virtual_table" }

// GenCreateVirtualTable generates a CREATE VIRTUAL TABLE statement.
// According to Turso COMPAT.md: Yes (full support).
// Supports FTS5 (Full-Text Search) which is commonly available in Turso/LibSQL.
func GenCreateVirtualTable(lcg *common.LCG) (Stmt, error) {
	if lcg == nil {
		lcg = common.NewLCG(1)
	}

	// Generate table name
	tblName := fmt.Sprintf("vt_%d", lcg.Uint64()%100000)

	// Choose virtual table module
	// FTS5 is the most widely supported and useful virtual table in SQLite/LibSQL
	moduleChoice := lcg.Intn(2)
	var sql string

	switch moduleChoice {
	case 0:
		// FTS5 (Full-Text Search version 5)
		sql = genFTS5Table(tblName, lcg)
	default:
		// FTS5 with different configuration
		sql = genFTS5TableComplex(tblName, lcg)
	}

	return &CreateVirtualTableStmt{sql: sql}, nil
}

// genFTS5Table generates a simple FTS5 virtual table
func genFTS5Table(tblName string, lcg *common.LCG) string {
	// Generate 1-4 columns for FTS5
	numCols := 1 + lcg.Intn(4)
	cols := make([]string, numCols)

	for i := 0; i < numCols; i++ {
		cols[i] = fmt.Sprintf("col%d", i+1)
	}

	// Simple FTS5 table
	return fmt.Sprintf("CREATE VIRTUAL TABLE IF NOT EXISTS \"%s\" USING fts5(%s);",
		tblName, strings.Join(cols, ", "))
}

// genFTS5TableComplex generates an FTS5 virtual table with additional options
func genFTS5TableComplex(tblName string, lcg *common.LCG) string {
	// Generate 2-5 columns for FTS5
	numCols := 2 + lcg.Intn(4)
	cols := make([]string, numCols)

	for i := 0; i < numCols; i++ {
		// Sometimes add UNINDEXED keyword to columns (FTS5 feature)
		if i > 0 && lcg.Intn(3) == 0 {
			cols[i] = fmt.Sprintf("col%d UNINDEXED", i+1)
		} else {
			cols[i] = fmt.Sprintf("col%d", i+1)
		}
	}

	// FTS5 options
	options := []string{}

	// Add tokenizer option (50% chance)
	if lcg.Intn(2) == 0 {
		tokenizers := []string{"porter", "unicode61", "ascii"}
		tokenizer := tokenizers[lcg.Intn(len(tokenizers))]
		options = append(options, fmt.Sprintf("tokenize='%s'", tokenizer))
	}

	// Add content option (30% chance)
	if lcg.Intn(10) < 3 {
		// content='' creates a contentless FTS5 table
		// content='table_name' creates an external content FTS5 table
		if lcg.Intn(2) == 0 {
			options = append(options, "content=''")
		}
	}

	// Add columnsize option (20% chance)
	if lcg.Intn(5) == 0 {
		options = append(options, "columnsize=0")
	}

	// Build the full column list with options
	allParts := make([]string, 0, len(cols)+len(options))
	allParts = append(allParts, cols...)
	allParts = append(allParts, options...)

	return fmt.Sprintf("CREATE VIRTUAL TABLE IF NOT EXISTS \"%s\" USING fts5(%s);",
		tblName, strings.Join(allParts, ", "))
}
