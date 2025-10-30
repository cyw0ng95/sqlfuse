package generators

// DepthHelper provides depth-based decision making for SQL generation.
// Inspired by the original SQLsmith's d6(), d9(), d20(), d42(), d100() functions
// which use dice rolls to make probabilistic decisions.
type DepthHelper struct {
	depth    int
	maxDepth int
}

// NewDepthHelper creates a depth helper with the current and maximum depth.
func NewDepthHelper(depth, maxDepth int) *DepthHelper {
	return &DepthHelper{
		depth:    depth,
		maxDepth: maxDepth,
	}
}

// ShouldRecurse returns true if we should generate a recursive construct.
// The probability decreases as we approach maxDepth.
// Similar to "if (level < d6())" checks in original SQLsmith.
func (dh *DepthHelper) ShouldRecurse(threshold int) bool {
	if dh.depth >= dh.maxDepth {
		return false
	}
	// Higher depth means lower chance of recursion
	// threshold typically 3-6
	return dh.depth < threshold
}

// CanRecurse returns true if we haven't exceeded the maximum depth.
func (dh *DepthHelper) CanRecurse() bool {
	return dh.depth < dh.maxDepth
}

// AtMaxDepth returns true if we're at the maximum recursion depth.
func (dh *DepthHelper) AtMaxDepth() bool {
	return dh.depth >= dh.maxDepth
}

// ShouldGenerateComplex returns true if we should generate a complex construct.
// Used for decisions like "should I generate a subquery vs a simple reference".
// Probability decreases with depth.
func (dh *DepthHelper) ShouldGenerateComplex(baseProbability int, maxProbability int) bool {
	if dh.depth >= dh.maxDepth {
		return false
	}
	// At depth 0, use maxProbability
	// As depth increases, reduce probability
	if dh.maxDepth == 0 {
		return dh.depth < baseProbability
	}
	// Scale probability based on remaining depth
	remaining := dh.maxDepth - dh.depth
	scaledThreshold := baseProbability + (remaining * (maxProbability - baseProbability) / dh.maxDepth)
	return dh.depth < scaledThreshold
}

// Descend returns a new DepthHelper with incremented depth.
func (dh *DepthHelper) Descend() *DepthHelper {
	return &DepthHelper{
		depth:    dh.depth + 1,
		maxDepth: dh.maxDepth,
	}
}

// GetDepth returns the current depth.
func (dh *DepthHelper) GetDepth() int {
	return dh.depth
}

// GetMaxDepth returns the maximum depth.
func (dh *DepthHelper) GetMaxDepth() int {
	return dh.maxDepth
}
