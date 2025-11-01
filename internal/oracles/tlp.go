package oracles

import (
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
	"strings"
)

// TLPOracle implements Ternary Logic Partitioning oracle
// This technique partitions a query into three queries that should return
// the same results when combined: WHERE p, WHERE NOT p, WHERE p IS NULL
type TLPOracle struct {
	db        *sql.DB
	lcg       *common.LCG
	tableName string
	columns   []string
}

// NewTLPOracle creates a new TLP oracle
func NewTLPOracle(db *sql.DB, lcg *common.LCG, tableName string, columns []string) *TLPOracle {
	return &TLPOracle{
		db:        db,
		lcg:       lcg,
		tableName: tableName,
		columns:   columns,
	}
}

// Name returns the oracle name
func (t *TLPOracle) Name() string {
	return "TLP (Ternary Logic Partitioning)"
}

// Check performs the TLP check
func (t *TLPOracle) Check() error {
	// Generate a random predicate
	predicate := t.generatePredicate()
	
	// Build the original query
	originalQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", t.tableName)
	
	// Build the three partition queries
	queryTrue := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", t.tableName, predicate)
	queryFalse := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE NOT (%s)", t.tableName, predicate)
	queryNull := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE (%s) IS NULL", t.tableName, predicate)
	
	// Execute queries
	originalResult, err := t.executeCountQuery(originalQuery)
	if err != nil {
		// Syntax errors are expected and not bugs
		return nil
	}
	
	trueResult, err := t.executeCountQuery(queryTrue)
	if err != nil {
		return nil
	}
	
	falseResult, err := t.executeCountQuery(queryFalse)
	if err != nil {
		return nil
	}
	
	nullResult, err := t.executeCountQuery(queryNull)
	if err != nil {
		return nil
	}
	
	// Verify: original count = true count + false count + null count
	sum := trueResult + falseResult + nullResult
	if originalResult != sum {
		return fmt.Errorf("TLP bug detected: original=%d, true=%d, false=%d, null=%d, sum=%d\nPredicate: %s",
			originalResult, trueResult, falseResult, nullResult, sum, predicate)
	}
	
	return nil
}

// executeCountQuery executes a COUNT query and returns the result
func (t *TLPOracle) executeCountQuery(query string) (int64, error) {
	var count int64
	err := t.db.QueryRow(query).Scan(&count)
	return count, err
}

// generatePredicate generates a random boolean predicate
func (t *TLPOracle) generatePredicate() string {
	if len(t.columns) == 0 {
		return "1 = 1"
	}
	
	predicateType := t.lcg.Intn(5)
	col := t.columns[t.lcg.Intn(len(t.columns))]
	
	switch predicateType {
	case 0:
		// Simple comparison
		value := t.lcg.Intn(100)
		operators := []string{"=", "<", ">", "<=", ">=", "!="}
		op := operators[t.lcg.Intn(len(operators))]
		return fmt.Sprintf("%s %s %d", col, op, value)
	case 1:
		// IS NULL / IS NOT NULL
		if t.lcg.Intn(2) == 0 {
			return fmt.Sprintf("%s IS NULL", col)
		}
		return fmt.Sprintf("%s IS NOT NULL", col)
	case 2:
		// BETWEEN
		low := t.lcg.Intn(50)
		high := low + t.lcg.Intn(50)
		return fmt.Sprintf("%s BETWEEN %d AND %d", col, low, high)
	case 3:
		// IN
		values := make([]string, t.lcg.Intn(5)+1)
		for i := range values {
			values[i] = fmt.Sprintf("%d", t.lcg.Intn(100))
		}
		return fmt.Sprintf("%s IN (%s)", col, strings.Join(values, ", "))
	default:
		// Compound predicate
		if len(t.columns) > 1 {
			col2 := t.columns[t.lcg.Intn(len(t.columns))]
			operators := []string{"=", "<", ">", "!="}
			op := operators[t.lcg.Intn(len(operators))]
			return fmt.Sprintf("%s %s %s", col, op, col2)
		}
		return fmt.Sprintf("%s > 0", col)
	}
}
