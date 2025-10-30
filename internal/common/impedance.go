package common

import (
	"fmt"
	"sort"
	"sync"
)

// ImpedanceMatcher tracks production success/failure rates to blacklist
// problematic productions. This is inspired by the original SQLsmith's
// impedance mismatch detection.
type ImpedanceMatcher struct {
	mu                 sync.RWMutex
	failedQueries      map[string]int64 // production name -> failure count
	okQueries          map[string]int64 // production name -> success count
	retries            map[string]int64 // production name -> retry count
	limited            map[string]int64 // production name -> limit hit count
	blacklistThreshold float64          // error rate threshold for blacklisting (default 0.99)
	minObservations    int64            // minimum observations before blacklisting (default 100)
	enabled            bool
}

// NewImpedanceMatcher creates a new impedance matcher with default settings.
func NewImpedanceMatcher() *ImpedanceMatcher {
	return &ImpedanceMatcher{
		failedQueries:      make(map[string]int64),
		okQueries:          make(map[string]int64),
		retries:            make(map[string]int64),
		limited:            make(map[string]int64),
		blacklistThreshold: 0.99,
		minObservations:    100,
		enabled:            true,
	}
}

// SetEnabled enables or disables impedance matching.
func (im *ImpedanceMatcher) SetEnabled(enabled bool) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.enabled = enabled
}

// SetBlacklistThreshold sets the error rate threshold for blacklisting (0.0 to 1.0).
// Productions with error_rate > threshold are blacklisted.
func (im *ImpedanceMatcher) SetBlacklistThreshold(threshold float64) {
	im.mu.Lock()
	defer im.mu.Unlock()
	if threshold >= 0.0 && threshold <= 1.0 {
		im.blacklistThreshold = threshold
	}
}

// SetMinObservations sets the minimum number of observations before blacklisting.
func (im *ImpedanceMatcher) SetMinObservations(min int64) {
	im.mu.Lock()
	defer im.mu.Unlock()
	if min > 0 {
		im.minObservations = min
	}
}

// RecordSuccess records a successful execution of a production.
func (im *ImpedanceMatcher) RecordSuccess(productionName string) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.okQueries[productionName]++
}

// RecordFailure records a failed execution of a production.
func (im *ImpedanceMatcher) RecordFailure(productionName string) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.failedQueries[productionName]++
}

// RecordRetry records a retry attempt for a production.
func (im *ImpedanceMatcher) RecordRetry(productionName string) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.retries[productionName]++
}

// RecordLimit records when a production hits its retry limit.
func (im *ImpedanceMatcher) RecordLimit(productionName string) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.limited[productionName]++
}

// IsBlacklisted checks if a production is blacklisted due to high error rate.
func (im *ImpedanceMatcher) IsBlacklisted(productionName string) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()

	if !im.enabled {
		return false
	}

	failed := im.failedQueries[productionName]
	ok := im.okQueries[productionName]

	// Need minimum observations before blacklisting
	if failed < im.minObservations {
		return false
	}

	// Calculate error rate
	total := failed + ok
	if total == 0 {
		return false
	}

	errorRate := float64(failed) / float64(total)
	return errorRate > im.blacklistThreshold
}

// GetErrorRate returns the error rate for a production (0.0 to 1.0).
func (im *ImpedanceMatcher) GetErrorRate(productionName string) float64 {
	im.mu.RLock()
	defer im.mu.RUnlock()

	failed := im.failedQueries[productionName]
	ok := im.okQueries[productionName]
	total := failed + ok

	if total == 0 {
		return 0.0
	}

	return float64(failed) / float64(total)
}

// GetStats returns statistics for a production.
func (im *ImpedanceMatcher) GetStats(productionName string) (failed, ok, retries, limited int64) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return im.failedQueries[productionName],
		im.okQueries[productionName],
		im.retries[productionName],
		im.limited[productionName]
}

// Report generates a human-readable report of impedance statistics.
func (im *ImpedanceMatcher) Report() string {
	im.mu.RLock()
	defer im.mu.RUnlock()

	if len(im.failedQueries) == 0 {
		return "No impedance data collected"
	}

	// Collect all productions
	productions := make([]string, 0, len(im.failedQueries))
	for prod := range im.failedQueries {
		productions = append(productions, prod)
	}

	// Sort by error rate (descending)
	sort.Slice(productions, func(i, j int) bool {
		errRateI := im.getErrorRateUnsafe(productions[i])
		errRateJ := im.getErrorRateUnsafe(productions[j])
		return errRateI > errRateJ
	})

	report := "Impedance Report:\n"
	report += "================================================================================\n"
	report += fmt.Sprintf("%-40s %8s %8s %8s %8s %10s %s\n",
		"Production", "Failed", "OK", "Retries", "Limited", "ErrorRate", "Status")
	report += "--------------------------------------------------------------------------------\n"

	for _, prod := range productions {
		failed := im.failedQueries[prod]
		ok := im.okQueries[prod]
		retries := im.retries[prod]
		limited := im.limited[prod]
		errorRate := im.getErrorRateUnsafe(prod)

		status := "OK"
		if im.isBlacklistedUnsafe(prod) {
			status = "BLACKLISTED"
		}

		report += fmt.Sprintf("%-40s %8d %8d %8d %8d %9.2f%% %s\n",
			truncate(prod, 40), failed, ok, retries, limited, errorRate*100, status)
	}

	report += "================================================================================\n"
	return report
}

// getErrorRateUnsafe calculates error rate without locking (caller must hold lock).
func (im *ImpedanceMatcher) getErrorRateUnsafe(productionName string) float64 {
	failed := im.failedQueries[productionName]
	ok := im.okQueries[productionName]
	total := failed + ok

	if total == 0 {
		return 0.0
	}

	return float64(failed) / float64(total)
}

// isBlacklistedUnsafe checks blacklist status without locking (caller must hold lock).
func (im *ImpedanceMatcher) isBlacklistedUnsafe(productionName string) bool {
	if !im.enabled {
		return false
	}

	failed := im.failedQueries[productionName]
	if failed < im.minObservations {
		return false
	}

	errorRate := im.getErrorRateUnsafe(productionName)
	return errorRate > im.blacklistThreshold
}

// Reset clears all impedance statistics.
func (im *ImpedanceMatcher) Reset() {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.failedQueries = make(map[string]int64)
	im.okQueries = make(map[string]int64)
	im.retries = make(map[string]int64)
	im.limited = make(map[string]int64)
}

// truncate truncates a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
