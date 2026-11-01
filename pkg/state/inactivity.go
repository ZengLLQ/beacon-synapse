package state

const (
	MinEpochsToInactivityPenalty    uint64 = 4
	InactivityScoreBias             uint64 = 4
	InactivityPenaltyQuotientAltair uint64 = 3 * 1024 * 1024 * 1024 // 3 * 2^30
)

// InactivityTracker manages per-validator inactivity scores and leak penalties.
type InactivityTracker struct {
	Scores []uint64
}

// NewInactivityTracker initializes tracker for n validators.
func NewInactivityTracker(validatorCount int) *InactivityTracker {
	return &InactivityTracker{
		Scores: make([]uint64, validatorCount),
	}
}

// ProcessEpoch updates inactivity scores based on attestation participation.
// If isLeak is true (epochs_since_finality > 4), it calculates penalties.
func (it *InactivityTracker) ProcessEpoch(
	validators []*Validator,
	balances []uint64,
	hasAttested []bool,
	isLeak bool,
) []uint64 {
	penalties := make([]uint64, len(validators))

	for i := range validators {
		if i >= len(it.Scores) || i >= len(hasAttested) || i >= len(balances) {
			break
		}

		// Update scores
		if hasAttested[i] {
			if it.Scores[i] > 0 {
				it.Scores[i]--
			}
		} else {
			it.Scores[i] += InactivityScoreBias
		}

		// Apply leak penalty if chain has not finalized
		if isLeak {
			penalty := (validators[i].EffectiveBalance * it.Scores[i]) / InactivityPenaltyQuotientAltair
			if penalty > balances[i] {
				penalty = balances[i]
			}
			penalties[i] = penalty
			balances[i] -= penalty

			// Update effective balance with hysteresis
			validators[i].UpdateEffectiveBalance(balances[i])
		}
	}

	return penalties
}
