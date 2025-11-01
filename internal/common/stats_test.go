package common

import (
	"testing"
	"time"
)

func TestGenerationStats_Basic(t *testing.T) {
	stats := NewGenerationStats()

	stats.RecordGeneration(10 * time.Millisecond)
	stats.RecordExecution(5 * time.Millisecond)

	snapshot := stats.GetStats()
	if snapshot.QueriesGenerated != 1 {
		t.Errorf("Expected 1 query generated, got %d", snapshot.QueriesGenerated)
	}
	if snapshot.QueriesExecuted != 1 {
		t.Errorf("Expected 1 query executed, got %d", snapshot.QueriesExecuted)
	}
	if snapshot.SuccessfulQueries != 1 {
		t.Errorf("Expected 1 successful query, got %d", snapshot.SuccessfulQueries)
	}
}

func TestGenerationStats_Errors(t *testing.T) {
	stats := NewGenerationStats()

	stats.RecordSyntaxError()
	stats.RecordExecutionError("test error 1")
	stats.RecordExecutionError("test error 1") // duplicate
	stats.RecordExecutionError("test error 2")
	stats.RecordTimeout()
	stats.RecordBrokenConnection()

	snapshot := stats.GetStats()
	if snapshot.SyntaxErrors != 1 {
		t.Errorf("Expected 1 syntax error, got %d", snapshot.SyntaxErrors)
	}
	if snapshot.ExecutionErrors != 3 {
		t.Errorf("Expected 3 execution errors, got %d", snapshot.ExecutionErrors)
	}
	if snapshot.Timeouts != 1 {
		t.Errorf("Expected 1 timeout, got %d", snapshot.Timeouts)
	}
	if snapshot.BrokenConnections != 1 {
		t.Errorf("Expected 1 broken connection, got %d", snapshot.BrokenConnections)
	}

	// Check error messages tracking
	if snapshot.ErrorMessages["test error 1"] != 2 {
		t.Errorf("Expected 2 occurrences of 'test error 1', got %d", snapshot.ErrorMessages["test error 1"])
	}
	if snapshot.ErrorMessages["test error 2"] != 1 {
		t.Errorf("Expected 1 occurrence of 'test error 2', got %d", snapshot.ErrorMessages["test error 2"])
	}
}

func TestGenerationStats_AST(t *testing.T) {
	stats := NewGenerationStats()

	stats.RecordAST(5, 20)
	stats.RecordAST(7, 30)
	stats.RecordAST(3, 10)

	snapshot := stats.GetStats()
	expectedAvgHeight := (5.0 + 7.0 + 3.0) / 3.0
	expectedAvgNodes := (20.0 + 30.0 + 10.0) / 3.0

	if snapshot.AvgASTHeight != expectedAvgHeight {
		t.Errorf("Expected avg AST height %.2f, got %.2f", expectedAvgHeight, snapshot.AvgASTHeight)
	}
	if snapshot.AvgASTNodes != expectedAvgNodes {
		t.Errorf("Expected avg AST nodes %.2f, got %.2f", expectedAvgNodes, snapshot.AvgASTNodes)
	}
}

func TestGenerationStats_Rates(t *testing.T) {
	stats := NewGenerationStats()

	// Wait a bit to ensure non-zero elapsed time
	time.Sleep(10 * time.Millisecond)

	for i := 0; i < 100; i++ {
		stats.RecordGeneration(1 * time.Millisecond)
	}
	for i := 0; i < 50; i++ {
		stats.RecordExecution(1 * time.Millisecond)
	}

	snapshot := stats.GetStats()

	if snapshot.GenPerSec <= 0 {
		t.Error("Expected positive generation rate")
	}
	if snapshot.ExecPerSec <= 0 {
		t.Error("Expected positive execution rate")
	}
	if snapshot.Elapsed <= 0 {
		t.Error("Expected positive elapsed time")
	}
}

func TestGenerationStats_ErrorRate(t *testing.T) {
	stats := NewGenerationStats()

	// 3 successful, 1 error = 25% error rate
	stats.RecordExecution(1 * time.Millisecond)
	stats.RecordExecution(1 * time.Millisecond)
	stats.RecordExecution(1 * time.Millisecond)
	stats.RecordExecutionError("test error")

	snapshot := stats.GetStats()
	expectedErrorRate := 1.0 / 4.0

	if snapshot.ErrorRate < expectedErrorRate-0.01 || snapshot.ErrorRate > expectedErrorRate+0.01 {
		t.Errorf("Expected error rate %.4f, got %.4f", expectedErrorRate, snapshot.ErrorRate)
	}
}

func TestGenerationStats_Reset(t *testing.T) {
	stats := NewGenerationStats()

	stats.RecordGeneration(10 * time.Millisecond)
	stats.RecordExecution(5 * time.Millisecond)
	stats.RecordSyntaxError()

	stats.Reset()

	snapshot := stats.GetStats()
	if snapshot.QueriesGenerated != 0 {
		t.Error("Expected 0 queries generated after reset")
	}
	if snapshot.SyntaxErrors != 0 {
		t.Error("Expected 0 syntax errors after reset")
	}
}

func TestGenerationStats_Report(t *testing.T) {
	stats := NewGenerationStats()

	stats.RecordGeneration(10 * time.Millisecond)
	stats.RecordExecution(5 * time.Millisecond)
	stats.RecordExecutionError("test error")
	stats.RecordAST(5, 20)

	report := stats.Report()
	if report == "" {
		t.Error("Report should not be empty")
	}
	if len(report) < 100 {
		t.Error("Report seems too short")
	}
}

func TestGenerationStats_Concurrent(t *testing.T) {
	stats := NewGenerationStats()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				stats.RecordGeneration(1 * time.Millisecond)
				stats.RecordExecution(1 * time.Millisecond)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	snapshot := stats.GetStats()
	expected := int64(1000) // 10 goroutines * 100 iterations
	if snapshot.QueriesGenerated != expected {
		t.Errorf("Expected %d queries generated, got %d", expected, snapshot.QueriesGenerated)
	}
	if snapshot.QueriesExecuted != expected {
		t.Errorf("Expected %d queries executed, got %d", expected, snapshot.QueriesExecuted)
	}
}
