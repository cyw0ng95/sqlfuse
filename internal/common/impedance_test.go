package common

import (
	"testing"
)

func TestImpedanceMatcher_Basic(t *testing.T) {
	im := NewImpedanceMatcher()

	// Record some successes
	im.RecordSuccess("select_basic")
	im.RecordSuccess("select_basic")
	im.RecordSuccess("select_basic")

	// Check not blacklisted (not enough failures)
	if im.IsBlacklisted("select_basic") {
		t.Error("select_basic should not be blacklisted with only successes")
	}

	// Get stats
	failed, ok, _, _ := im.GetStats("select_basic")
	if failed != 0 || ok != 3 {
		t.Errorf("Expected 0 failures, 3 successes, got %d failures, %d successes", failed, ok)
	}
}

func TestImpedanceMatcher_Blacklisting(t *testing.T) {
	im := NewImpedanceMatcher()
	im.SetMinObservations(10)
	im.SetBlacklistThreshold(0.9)

	// Record failures below threshold
	for i := 0; i < 9; i++ {
		im.RecordFailure("bad_production")
	}
	im.RecordSuccess("bad_production")

	// Not enough observations yet
	if im.IsBlacklisted("bad_production") {
		t.Error("Should not be blacklisted with < 10 observations")
	}

	// Add one more failure to reach 10 observations with 100% error rate
	im.RecordFailure("bad_production")

	// Now should be blacklisted (10 failures, 1 success = 90.9% error rate > 90%)
	if !im.IsBlacklisted("bad_production") {
		t.Error("Should be blacklisted with 90.9% error rate")
	}

	// Check error rate
	errorRate := im.GetErrorRate("bad_production")
	expected := 10.0 / 11.0 // 10 failures out of 11 total
	if errorRate < expected-0.01 || errorRate > expected+0.01 {
		t.Errorf("Expected error rate ~%.2f, got %.2f", expected, errorRate)
	}
}

func TestImpedanceMatcher_NotBlacklistedLowErrorRate(t *testing.T) {
	im := NewImpedanceMatcher()
	im.SetMinObservations(10)
	im.SetBlacklistThreshold(0.9)

	// 50% error rate (well below threshold)
	for i := 0; i < 50; i++ {
		im.RecordFailure("medium_production")
	}
	for i := 0; i < 50; i++ {
		im.RecordSuccess("medium_production")
	}

	if im.IsBlacklisted("medium_production") {
		t.Error("Should not be blacklisted with 50% error rate")
	}
}

func TestImpedanceMatcher_Retries(t *testing.T) {
	im := NewImpedanceMatcher()

	im.RecordRetry("retry_production")
	im.RecordRetry("retry_production")
	im.RecordLimit("retry_production")

	_, _, retries, limited := im.GetStats("retry_production")
	if retries != 2 {
		t.Errorf("Expected 2 retries, got %d", retries)
	}
	if limited != 1 {
		t.Errorf("Expected 1 limit hit, got %d", limited)
	}
}

func TestImpedanceMatcher_Disabled(t *testing.T) {
	im := NewImpedanceMatcher()
	im.SetEnabled(false)
	im.SetMinObservations(10)

	// Record 100% failures
	for i := 0; i < 100; i++ {
		im.RecordFailure("should_not_blacklist")
	}

	// Should not be blacklisted when disabled
	if im.IsBlacklisted("should_not_blacklist") {
		t.Error("Should not be blacklisted when impedance matching is disabled")
	}

	// Re-enable
	im.SetEnabled(true)
	if !im.IsBlacklisted("should_not_blacklist") {
		t.Error("Should be blacklisted after re-enabling")
	}
}

func TestImpedanceMatcher_Reset(t *testing.T) {
	im := NewImpedanceMatcher()

	im.RecordSuccess("prod1")
	im.RecordFailure("prod2")
	im.RecordRetry("prod3")

	im.Reset()

	// All stats should be cleared
	failed, ok, retries, limited := im.GetStats("prod1")
	if failed != 0 || ok != 0 || retries != 0 || limited != 0 {
		t.Error("Stats should be cleared after reset")
	}
}

func TestImpedanceMatcher_Report(t *testing.T) {
	im := NewImpedanceMatcher()
	im.SetMinObservations(5)

	// Add some data
	for i := 0; i < 10; i++ {
		im.RecordFailure("bad_prod")
	}
	im.RecordSuccess("bad_prod")

	for i := 0; i < 10; i++ {
		im.RecordSuccess("good_prod")
	}
	im.RecordFailure("good_prod")

	report := im.Report()
	if report == "" {
		t.Error("Report should not be empty")
	}

	// Report should contain production names
	if len(report) < 100 {
		t.Error("Report seems too short")
	}
}

func TestImpedanceMatcher_Concurrent(t *testing.T) {
	im := NewImpedanceMatcher()

	// Test concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				im.RecordSuccess("concurrent_prod")
				im.RecordFailure("concurrent_prod")
				_ = im.IsBlacklisted("concurrent_prod")
				_ = im.GetErrorRate("concurrent_prod")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check that we recorded the expected number of events
	failed, ok, _, _ := im.GetStats("concurrent_prod")
	expected := int64(1000) // 10 goroutines * 100 iterations
	if failed != expected || ok != expected {
		t.Errorf("Expected %d failures and %d successes, got %d failures and %d successes",
			expected, expected, failed, ok)
	}
}
