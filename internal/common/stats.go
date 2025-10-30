package common

import (
	"fmt"
	"sync"
	"time"
)

// GenerationStats tracks statistics about SQL generation and execution.
// Inspired by the original SQLsmith's statistics reporting.
type GenerationStats struct {
	mu sync.RWMutex

	// Generation metrics
	queriesGenerated   int64
	queriesExecuted    int64
	syntaxErrors       int64
	executionErrors    int64
	timeouts           int64
	brokenConnections  int64
	successfulQueries  int64

	// AST metrics (moving averages)
	totalASTHeight int64
	totalASTNodes  int64
	astSamples     int64

	// Timing
	startTime        time.Time
	totalGenTime     time.Duration
	totalExecTime    time.Duration

	// Error tracking by message
	errorMessages    map[string]int64
	maxErrorMessages int
}

// NewGenerationStats creates a new statistics tracker.
func NewGenerationStats() *GenerationStats {
	return &GenerationStats{
		startTime:        time.Now(),
		errorMessages:    make(map[string]int64),
		maxErrorMessages: 100, // Track top 100 error messages
	}
}

// RecordGeneration records that a query was generated.
func (gs *GenerationStats) RecordGeneration(genTime time.Duration) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.queriesGenerated++
	gs.totalGenTime += genTime
}

// RecordExecution records a successful query execution.
func (gs *GenerationStats) RecordExecution(execTime time.Duration) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.queriesExecuted++
	gs.successfulQueries++
	gs.totalExecTime += execTime
}

// RecordSyntaxError records a syntax error.
func (gs *GenerationStats) RecordSyntaxError() {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.syntaxErrors++
}

// RecordExecutionError records an execution error.
func (gs *GenerationStats) RecordExecutionError(errorMsg string) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.executionErrors++
	gs.queriesExecuted++

	// Track error message if we haven't hit the limit
	if len(gs.errorMessages) < gs.maxErrorMessages {
		gs.errorMessages[errorMsg]++
	} else if _, exists := gs.errorMessages[errorMsg]; exists {
		gs.errorMessages[errorMsg]++
	}
}

// RecordTimeout records a query timeout.
func (gs *GenerationStats) RecordTimeout() {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.timeouts++
	gs.queriesExecuted++
}

// RecordBrokenConnection records a broken database connection.
func (gs *GenerationStats) RecordBrokenConnection() {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.brokenConnections++
}

// RecordAST records AST statistics for a generated query.
func (gs *GenerationStats) RecordAST(height, nodes int) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.totalASTHeight += int64(height)
	gs.totalASTNodes += int64(nodes)
	gs.astSamples++
}

// GetStats returns current statistics snapshot.
func (gs *GenerationStats) GetStats() StatsSnapshot {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	elapsed := time.Since(gs.startTime)

	// Calculate rates
	var genPerSec, execPerSec float64
	if elapsed.Seconds() > 0 {
		genPerSec = float64(gs.queriesGenerated) / elapsed.Seconds()
		execPerSec = float64(gs.queriesExecuted) / elapsed.Seconds()
	}

	// Calculate averages
	var avgASTHeight, avgASTNodes float64
	if gs.astSamples > 0 {
		avgASTHeight = float64(gs.totalASTHeight) / float64(gs.astSamples)
		avgASTNodes = float64(gs.totalASTNodes) / float64(gs.astSamples)
	}

	// Calculate error rate
	var errorRate float64
	if gs.queriesExecuted > 0 {
		errorRate = float64(gs.executionErrors) / float64(gs.queriesExecuted)
	}

	// Copy error messages
	errorMsgs := make(map[string]int64, len(gs.errorMessages))
	for msg, count := range gs.errorMessages {
		errorMsgs[msg] = count
	}

	return StatsSnapshot{
		QueriesGenerated:  gs.queriesGenerated,
		QueriesExecuted:   gs.queriesExecuted,
		SyntaxErrors:      gs.syntaxErrors,
		ExecutionErrors:   gs.executionErrors,
		Timeouts:          gs.timeouts,
		BrokenConnections: gs.brokenConnections,
		SuccessfulQueries: gs.successfulQueries,
		AvgASTHeight:      avgASTHeight,
		AvgASTNodes:       avgASTNodes,
		Elapsed:           elapsed,
		GenPerSec:         genPerSec,
		ExecPerSec:        execPerSec,
		ErrorRate:         errorRate,
		ErrorMessages:     errorMsgs,
	}
}

// Report generates a human-readable statistics report.
func (gs *GenerationStats) Report() string {
	stats := gs.GetStats()

	report := "Generation Statistics:\n"
	report += "================================================================================\n"
	report += fmt.Sprintf("Queries generated: %d (%.2f gen/s)\n", stats.QueriesGenerated, stats.GenPerSec)
	report += fmt.Sprintf("Queries executed:  %d (%.2f exec/s)\n", stats.QueriesExecuted, stats.ExecPerSec)
	report += fmt.Sprintf("Successful:        %d\n", stats.SuccessfulQueries)
	report += fmt.Sprintf("Syntax errors:     %d\n", stats.SyntaxErrors)
	report += fmt.Sprintf("Execution errors:  %d\n", stats.ExecutionErrors)
	report += fmt.Sprintf("Timeouts:          %d\n", stats.Timeouts)
	report += fmt.Sprintf("Broken connections:%d\n", stats.BrokenConnections)
	report += fmt.Sprintf("Error rate:        %.4f\n", stats.ErrorRate)
	report += fmt.Sprintf("Elapsed time:      %s\n", stats.Elapsed.Round(time.Millisecond))
	report += "\n"

	if stats.AvgASTNodes > 0 {
		report += fmt.Sprintf("AST stats (avg):   height = %.2f, nodes = %.2f\n", stats.AvgASTHeight, stats.AvgASTNodes)
		report += "\n"
	}

	if len(stats.ErrorMessages) > 0 {
		report += "Top errors:\n"
		report += "--------------------------------------------------------------------------------\n"

		// Sort errors by frequency
		type errorCount struct {
			msg   string
			count int64
		}
		errors := make([]errorCount, 0, len(stats.ErrorMessages))
		for msg, count := range stats.ErrorMessages {
			errors = append(errors, errorCount{msg, count})
		}

		// Simple bubble sort for top errors
		for i := 0; i < len(errors); i++ {
			for j := i + 1; j < len(errors); j++ {
				if errors[j].count > errors[i].count {
					errors[i], errors[j] = errors[j], errors[i]
				}
			}
		}

		// Show top 20
		maxShow := 20
		if len(errors) < maxShow {
			maxShow = len(errors)
		}
		for i := 0; i < maxShow; i++ {
			msg := errors[i].msg
			if len(msg) > 70 {
				msg = msg[:67] + "..."
			}
			report += fmt.Sprintf("%4d  %s\n", errors[i].count, msg)
		}
	}

	report += "================================================================================\n"
	return report
}

// Reset clears all statistics.
func (gs *GenerationStats) Reset() {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.queriesGenerated = 0
	gs.queriesExecuted = 0
	gs.syntaxErrors = 0
	gs.executionErrors = 0
	gs.timeouts = 0
	gs.brokenConnections = 0
	gs.successfulQueries = 0
	gs.totalASTHeight = 0
	gs.totalASTNodes = 0
	gs.astSamples = 0
	gs.startTime = time.Now()
	gs.totalGenTime = 0
	gs.totalExecTime = 0
	gs.errorMessages = make(map[string]int64)
}

// StatsSnapshot is an immutable snapshot of statistics.
type StatsSnapshot struct {
	QueriesGenerated  int64
	QueriesExecuted   int64
	SyntaxErrors      int64
	ExecutionErrors   int64
	Timeouts          int64
	BrokenConnections int64
	SuccessfulQueries int64
	AvgASTHeight      float64
	AvgASTNodes       float64
	Elapsed           time.Duration
	GenPerSec         float64
	ExecPerSec        float64
	ErrorRate         float64
	ErrorMessages     map[string]int64
}
