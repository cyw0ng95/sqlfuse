package oracles

import (
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
	"strings"
)

// PQSOracle implements Pivoted Query Synthesis
// This technique generates queries that should return specific pivot rows
// and validates that the rows are actually returned
type PQSOracle struct {
	db        *sql.DB
	lcg       *common.LCG
	tableName string
	columns   []string
}

// NewPQSOracle creates a new PQS oracle
func NewPQSOracle(db *sql.DB, lcg *common.LCG, tableName string, columns []string) *PQSOracle {
	return &PQSOracle{
		db:        db,
		lcg:       lcg,
		tableName: tableName,
		columns:   columns,
	}
}

// Name returns the oracle name
func (p *PQSOracle) Name() string {
	return "PQS (Pivoted Query Synthesis)"
}

// Check performs the PQS check
func (p *PQSOracle) Check() error {
	// First, get a random row from the table (pivot row)
	pivotRow, err := p.getPivotRow()
	if err != nil {
		return nil // No data or error getting pivot row
	}
	
	// Generate a query that should definitely include this pivot row
	query := p.generatePivotedQuery(pivotRow)
	
	// Execute the query
	result, err := ExecuteQuery(p.db, query)
	if err != nil {
		return nil // Syntax errors are not bugs
	}
	
	// Check if the pivot row is in the results
	found := false
	for _, row := range result.Rows {
		if p.rowsEqual(row, pivotRow) {
			found = true
			break
		}
	}
	
	if !found {
		return fmt.Errorf("PQS bug detected: pivot row not found in results\nQuery: %s\nPivot row: %v\nResult count: %d",
			query, pivotRow, result.RowCount)
	}
	
	return nil
}

// getPivotRow retrieves a random row from the table
func (p *PQSOracle) getPivotRow() ([]interface{}, error) {
	if len(p.columns) == 0 {
		return nil, fmt.Errorf("no columns")
	}
	
	// Get a random row
	query := fmt.Sprintf("SELECT %s FROM %s ORDER BY RANDOM() LIMIT 1", strings.Join(p.columns, ", "), p.tableName)
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	if !rows.Next() {
		return nil, fmt.Errorf("no rows")
	}
	
	row := make([]interface{}, len(p.columns))
	rowPointers := make([]interface{}, len(p.columns))
	for i := range row {
		rowPointers[i] = &row[i]
	}
	
	if err := rows.Scan(rowPointers...); err != nil {
		return nil, err
	}
	
	return row, nil
}

// generatePivotedQuery generates a query that should include the pivot row
func (p *PQSOracle) generatePivotedQuery(pivotRow []interface{}) string {
	// Build WHERE clause that matches the pivot row exactly
	conditions := make([]string, 0, len(p.columns))
	
	for i, col := range p.columns {
		val := pivotRow[i]
		if val == nil {
			conditions = append(conditions, fmt.Sprintf("%s IS NULL", col))
		} else {
			// Format value based on type
			switch v := val.(type) {
			case int64:
				conditions = append(conditions, fmt.Sprintf("%s = %d", col, v))
			case float64:
				conditions = append(conditions, fmt.Sprintf("%s = %f", col, v))
			case string:
				// Escape single quotes
				escaped := strings.ReplaceAll(v, "'", "''")
				conditions = append(conditions, fmt.Sprintf("%s = '%s'", col, escaped))
			case []byte:
				// For BLOB types, use hex encoding
				conditions = append(conditions, fmt.Sprintf("%s = X'%x'", col, v))
			default:
				conditions = append(conditions, fmt.Sprintf("%s = %v", col, val))
			}
		}
	}
	
	whereClause := strings.Join(conditions, " AND ")
	
	// Add some additional conditions that should not filter out the row
	if p.lcg.Intn(2) == 1 {
		// Add an OR condition that is always false
		whereClause = fmt.Sprintf("(%s) OR (1 = 0)", whereClause)
	}
	
	return fmt.Sprintf("SELECT %s FROM %s WHERE %s", strings.Join(p.columns, ", "), p.tableName, whereClause)
}

// rowsEqual compares two rows for equality
func (p *PQSOracle) rowsEqual(r1, r2 []interface{}) bool {
	if len(r1) != len(r2) {
		return false
	}
	
	for i := range r1 {
		if r1[i] == nil && r2[i] == nil {
			continue
		}
		if r1[i] == nil || r2[i] == nil {
			return false
		}
		
		// Type-specific comparison
		switch v1 := r1[i].(type) {
		case []byte:
			v2, ok := r2[i].([]byte)
			if !ok {
				return false
			}
			if len(v1) != len(v2) {
				return false
			}
			for j := range v1 {
				if v1[j] != v2[j] {
					return false
				}
			}
		default:
			if r1[i] != r2[i] {
				return false
			}
		}
	}
	
	return true
}
