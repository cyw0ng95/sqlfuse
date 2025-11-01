package oracles

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"sqlfuse/internal/common"
)

// QPGOracle implements Query Plan Guidance
// This oracle tracks unique query plans to guide test generation toward
// unexplored execution paths
type QPGOracle struct {
	db              *sql.DB
	lcg             *common.LCG
	tableName       string
	columns         []string
	seenPlans       map[string]bool
	queriesSinceNew int
	threshold       int // Number of queries without new plans before mutating state
}

// NewQPGOracle creates a new QPG oracle
func NewQPGOracle(db *sql.DB, lcg *common.LCG, tableName string, columns []string) *QPGOracle {
	return &QPGOracle{
		db:              db,
		lcg:             lcg,
		tableName:       tableName,
		columns:         columns,
		seenPlans:       make(map[string]bool),
		queriesSinceNew: 0,
		threshold:       100, // Mutate state after 100 queries without new plans
	}
}

// Name returns the oracle name
func (q *QPGOracle) Name() string {
	return "QPG (Query Plan Guidance)"
}

// Check performs the QPG check and returns info about plan coverage
func (q *QPGOracle) Check() error {
	query := q.generateQuery()
	
	// Get the query plan using EXPLAIN QUERY PLAN
	planQuery := fmt.Sprintf("EXPLAIN QUERY PLAN %s", query)
	plan, err := q.getQueryPlan(planQuery)
	if err != nil {
		return nil // Query errors are not bugs
	}
	
	// Hash the plan to create a unique identifier
	planHash := q.hashPlan(plan)
	
	// Check if this is a new plan
	if !q.seenPlans[planHash] {
		q.seenPlans[planHash] = true
		q.queriesSinceNew = 0
	} else {
		q.queriesSinceNew++
	}
	
	// If we haven't seen a new plan in a while, suggest state mutation
	if q.queriesSinceNew >= q.threshold {
		return fmt.Errorf("QPG: %d queries without new plans (seen %d unique plans) - consider mutating database state",
			q.queriesSinceNew, len(q.seenPlans))
	}
	
	return nil
}

// GetPlanCoverage returns statistics about plan coverage
func (q *QPGOracle) GetPlanCoverage() (uniquePlans int, queriesSinceNew int) {
	return len(q.seenPlans), q.queriesSinceNew
}

// ShouldMutateState returns true if database state should be mutated
func (q *QPGOracle) ShouldMutateState() bool {
	return q.queriesSinceNew >= q.threshold
}

// ResetCounter resets the queries-since-new counter (after state mutation)
func (q *QPGOracle) ResetCounter() {
	q.queriesSinceNew = 0
}

// getQueryPlan retrieves the query plan for a query
func (q *QPGOracle) getQueryPlan(explainQuery string) (string, error) {
	rows, err := q.db.Query(explainQuery)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	
	var plan string
	for rows.Next() {
		var id int
		var parent int
		var notused int
		var detail string
		
		// EXPLAIN QUERY PLAN returns: id, parent, notused, detail
		if err := rows.Scan(&id, &parent, &notused, &detail); err != nil {
			return "", err
		}
		plan += detail + "\n"
	}
	
	if err := rows.Err(); err != nil {
		return "", err
	}
	
	return plan, nil
}

// hashPlan creates a hash of the query plan
func (q *QPGOracle) hashPlan(plan string) string {
	hash := sha256.Sum256([]byte(plan))
	return fmt.Sprintf("%x", hash)
}

// generateQuery generates a random SELECT query
func (q *QPGOracle) generateQuery() string {
	if len(q.columns) == 0 {
		return fmt.Sprintf("SELECT * FROM %s", q.tableName)
	}
	
	// Select random columns
	numCols := q.lcg.Intn(len(q.columns)) + 1
	selectedCols := make([]string, 0, numCols)
	used := make(map[int]bool)
	
	for i := 0; i < numCols; i++ {
		idx := q.lcg.Intn(len(q.columns))
		if !used[idx] {
			selectedCols = append(selectedCols, q.columns[idx])
			used[idx] = true
		}
	}
	
	if len(selectedCols) == 0 {
		selectedCols = []string{q.columns[0]}
	}
	
	query := fmt.Sprintf("SELECT %s FROM %s", selectedCols[0], q.tableName)
	for i := 1; i < len(selectedCols); i++ {
		query = fmt.Sprintf("%s, %s", query[:len(query)], selectedCols[i])
	}
	
	// Add WHERE clause with varying complexity
	whereType := q.lcg.Intn(4)
	if whereType > 0 && len(q.columns) > 0 {
		col := q.columns[q.lcg.Intn(len(q.columns))]
		value := q.lcg.Intn(100)
		
		switch whereType {
		case 1:
			query = fmt.Sprintf("%s WHERE %s = %d", query, col, value)
		case 2:
			query = fmt.Sprintf("%s WHERE %s > %d", query, col, value)
		case 3:
			query = fmt.Sprintf("%s WHERE %s BETWEEN %d AND %d", query, col, value, value+50)
		}
	}
	
	// Add ORDER BY to influence plan
	if q.lcg.Intn(3) == 1 && len(selectedCols) > 0 {
		orderCol := selectedCols[q.lcg.Intn(len(selectedCols))]
		query = fmt.Sprintf("%s ORDER BY %s", query, orderCol)
	}
	
	// Add LIMIT to influence plan
	if q.lcg.Intn(3) == 1 {
		limit := q.lcg.Intn(50) + 1
		query = fmt.Sprintf("%s LIMIT %d", query, limit)
	}
	
	return query
}
