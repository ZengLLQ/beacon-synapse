package test

import (
	"testing"

	"github.com/ZengLLQ/beacon-synapse/pkg/state"
)

func TestValidatorHysteresis(t *testing.T) {
	var pk [48]byte
	var wc [32]byte
	pk[0] = 0x01
	wc[0] = 0x01

	// Initial 32 ETH
	initialBal := 32 * state.GweiPerEth
	v := state.NewValidator(pk, wc, initialBal)

	if v.EffectiveBalance != 32*state.GweiPerEth {
		t.Fatalf("expected initial effective balance 32 ETH, got %d", v.EffectiveBalance)
	}

	// Drop to 31.8 ETH (within hysteresis 0.25 ETH downward threshold)
	v.UpdateEffectiveBalance(31_800_000_000)
	if v.EffectiveBalance != 32*state.GweiPerEth {
		t.Fatalf("expected effective balance to remain 32 ETH under hysteresis, got %d", v.EffectiveBalance)
	}

	// Drop to 31.7 ETH (past 31.75 ETH threshold)
	v.UpdateEffectiveBalance(31_700_000_000)
	if v.EffectiveBalance != 31*state.GweiPerEth {
		t.Fatalf("expected effective balance to drop to 31 ETH, got %d", v.EffectiveBalance)
	}

	// Rise to 32.1 ETH (below 31 + 1.25 = 32.25 ETH upward threshold)
	v.UpdateEffectiveBalance(32_100_000_000)
	if v.EffectiveBalance != 31*state.GweiPerEth {
		t.Fatalf("expected effective balance to remain 31 ETH below upward threshold, got %d", v.EffectiveBalance)
	}

	// Rise to 32.3 ETH (past 32.25 ETH upward threshold)
	v.UpdateEffectiveBalance(32_300_000_000)
	if v.EffectiveBalance != 32*state.GweiPerEth {
		t.Fatalf("expected effective balance to rise to 32 ETH, got %d", v.EffectiveBalance)
	}
}

func TestInactivityLeakPenalties(t *testing.T) {
	var pk [48]byte
	var wc [32]byte

	v0 := state.NewValidator(pk, wc, 32*state.GweiPerEth)
	v1 := state.NewValidator(pk, wc, 32*state.GweiPerEth)

	validators := []*state.Validator{v0, v1}
	balances := []uint64{32 * state.GweiPerEth, 32 * state.GweiPerEth}
	attesting := []bool{true, false} // v0 attests, v1 inactive

	tracker := state.NewInactivityTracker(2)

	// Non-leak epoch: scores change, zero penalty
	penalties := tracker.ProcessEpoch(validators, balances, attesting, false)
	if penalties[0] != 0 || penalties[1] != 0 {
		t.Fatalf("expected zero penalties outside inactivity leak")
	}
	if tracker.Scores[0] != 0 || tracker.Scores[1] != state.InactivityScoreBias {
		t.Fatalf("expected v1 score to increase by bias %d, got %d", state.InactivityScoreBias, tracker.Scores[1])
	}

	// Inactivity leak active
	penaltiesLeak := tracker.ProcessEpoch(validators, balances, attesting, true)
	if penaltiesLeak[0] != 0 {
		t.Fatalf("attesting validator should receive 0 penalty, got %d", penaltiesLeak[0])
	}
	if penaltiesLeak[1] == 0 {
		t.Fatalf("inactive validator must receive non-zero penalty during leak")
	}
	if balances[1] >= 32*state.GweiPerEth {
		t.Fatalf("inactive validator balance should decrease")
	}
}

func TestValidatorHashTreeRoot(t *testing.T) {
	var pk [48]byte
	var wc [32]byte
	pk[0] = 0x42
	wc[0] = 0x99

	v := state.NewValidator(pk, wc, 32*state.GweiPerEth)
	root := v.HashTreeRoot()

	var zero [32]byte
	if root == zero {
		t.Fatalf("hash tree root should not be zero")
	}

	// Same validator should yield identical root
	root2 := v.HashTreeRoot()
	if root != root2 {
		t.Fatalf("hash tree root should be deterministic")
	}

	// Modifying field should change root
	v.Slashed = true
	rootSlashed := v.HashTreeRoot()
	if root == rootSlashed {
		t.Fatalf("hash tree root must change when state changes")
	}
}
