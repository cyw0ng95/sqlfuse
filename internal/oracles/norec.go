package oracles

import (
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
)

// NoRECOracle implements Non-Optimizing Reference Engine Construction
// This technique compares query results with and without optimizations enabled
// to detect optimizer bugs
type NoRECOracle struct {
	db        *sql.DB
	lcg       *common.LCG
	tableName string
	columns   []string
}

// NewNoRECOracle creates a new NoREC oracle
func NewNoRECOracle(db *sql.DB, lcg *common.LCG, tableName string, columns []string) *NoRECOracle {
	return &NoRECOracle{
		db:        db,
		lcg:       lcg,
		tableName: tableName,
		columns:   columns,
	}
}

// Name returns the oracle name
func (n *NoRECOracle) Name() string {
	return "NoREC (Non-Optimizing Reference Engine Construction)"
}

// Check performs the NoREC check
func (n *NoRECOracle) Check() error {
	query := n.generateQuery()
	
	// Execute with optimizations enabled (default)
	_, err := n.db.Exec("PRAGMA optimize")
	if err != nil {
		return nil // Optimization pragma not supported
	}
	
	optimizedResult, err := ExecuteQuery(n.db, query)
	if err != nil {
		return nil // Query errors are not bugs for this check
	}
	
	// Execute with optimizations disabled
	// For SQLite, we can disable some optimizations
	disableOpts := []string{
		"PRAGMA automatic_index = OFF",
		"PRAGMA query_only = ON",
	}
	
	for _, opt := range disableOpts {
		_, _ = n.db.Exec(opt)
	}
	
	unoptimizedResult, err := ExecuteQuery(n.db, query)
	if err != nil {
		// Re-enable optimizations before returning
		_, _ = n.db.Exec("PRAGMA automatic_index = ON")
		_, _ = n.db.Exec("PRAGMA query_only = OFF")
		return nil
	}
	
	// Re-enable optimizations
	_, _ = n.db.Exec("PRAGMA automatic_index = ON")
	_, _ = n.db.Exec("PRAGMA query_only = OFF")
	
	// Compare results
	if !CompareResults(optimizedResult, unoptimizedResult) {
		return fmt.Errorf("NoREC bug detected: optimized and unoptimized results differ\nQuery: %s\nOptimized rows: %d, Unoptimized rows: %d",
			query, optimizedResult.RowCount, unoptimizedResult.RowCount)
	}
	
	return nil
}

// generateQuery generates a random SELECT query
func (n *NoRECOracle) generateQuery() string {
	if len(n.columns) == 0 {
		return fmt.Sprintf("SELECT * FROM %s", n.tableName)
	}
	
	// Select random columns
	numCols := n.lcg.Intn(len(n.columns)) + 1
	selectedCols := make([]string, 0, numCols)
	used := make(map[int]bool)
	
	for i := 0; i < numCols; i++ {
		idx := n.lcg.Intn(len(n.columns))
		if !used[idx] {
			selectedCols = append(selectedCols, n.columns[idx])
			used[idx] = true
		}
	}
	
	if len(selectedCols) == 0 {
		selectedCols = n.columns[:1]
	}
	
	query := fmt.Sprintf("SELECT %s FROM %s", selectedCols[0], n.tableName)
	for i := 1; i < len(selectedCols); i++ {
		query = fmt.Sprintf("%s, %s", query, selectedCols[i])
	}
	
	// Add optional WHERE clause
	if n.lcg.Intn(2) == 1 && len(n.columns) > 0 {
		col := n.columns[n.lcg.Intn(len(n.columns))]
		value := n.lcg.Intn(100)
		query = fmt.Sprintf("%s WHERE %s > %d", query, col, value)
	}
	
	// Add optional ORDER BY
	if n.lcg.Intn(2) == 1 && len(selectedCols) > 0 {
		orderCol := selectedCols[n.lcg.Intn(len(selectedCols))]
		query = fmt.Sprintf("%s ORDER BY %s", query, orderCol)
	}
	
	return query
}
