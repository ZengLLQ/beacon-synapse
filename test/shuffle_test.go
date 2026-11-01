package test

import (
	"crypto/sha256"
	"testing"

	"github.com/ZengLLQ/beacon-synapse/pkg/shuffle"
)

func TestSwapOrNotPermutationBijective(t *testing.T) {
	var seed [32]byte
	h := sha256.Sum256([]byte("eth2_pos_epoch_seed_42"))
	copy(seed[:], h[:])

	count := uint64(128)
	perm, err := shuffle.ComputePermutation(count, seed)
	if err != nil {
		t.Fatalf("permutation failed: %v", err)
	}

	seen := make(map[uint64]bool)
	for i, val := range perm {
		if val >= count {
			t.Fatalf("value %d out of bounds at index %d", val, i)
		}
		if seen[val] {
			t.Fatalf("duplicate value %d detected in permutation", val)
		}
		seen[val] = true
	}

	if len(seen) != int(count) {
		t.Fatalf("expected %d unique items, got %d", count, len(seen))
	}
}

func TestCommitteeAssignment(t *testing.T) {
	var seed [32]byte
	copy(seed[:], []byte("committee_seed_testing_123456789"))

	activeValidators := make([]uint64, 256)
	for i := 0; i < 256; i++ {
		activeValidators[i] = uint64(i * 10)
	}

	totalCommittees := uint64(4)
	totalMembers := 0

	for c := uint64(0); c < totalCommittees; c++ {
		committee, err := shuffle.ComputeCommittee(activeValidators, seed, c, totalCommittees)
		if err != nil {
			t.Fatalf("failed to compute committee %d: %v", c, err)
		}

		if len(committee) != 64 {
			t.Fatalf("expected committee size 64, got %d", len(committee))
		}
		totalMembers += len(committee)
	}

	if totalMembers != 256 {
		t.Fatalf("expected 256 total committee assignments, got %d", totalMembers)
	}
}
