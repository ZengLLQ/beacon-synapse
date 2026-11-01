package test

import (
	"testing"

	"github.com/ZengLLQ/beacon-synapse/pkg/merkle"
	"github.com/ZengLLQ/beacon-synapse/pkg/ssz"
)

func TestMerkleProofVerification(t *testing.T) {
	chunks := make([][32]byte, 8)
	for i := 0; i < 8; i++ {
		chunks[i][0] = byte(i + 1)
	}

	root := ssz.MerkleizeChunks(chunks, 8)

	// Test proof generation and verification for each leaf
	for i := uint64(0); i < 8; i++ {
		proof, gindex, err := merkle.GenerateProof(chunks, i)
		if err != nil {
			t.Fatalf("failed to generate proof for leaf %d: %v", i, err)
		}

		if len(proof) != 3 {
			t.Fatalf("expected proof depth 3 for 8 leaves, got %d", len(proof))
		}

		if gindex != 8+i {
			t.Fatalf("expected gindex %d, got %d", 8+i, gindex)
		}

		isValid := merkle.VerifyProof(root, chunks[i], proof, gindex)
		if !isValid {
			t.Fatalf("merkle proof verification failed for leaf %d", i)
		}

		// Tampered leaf should fail
		var tampered [32]byte
		tampered[0] = 0xFF
		if merkle.VerifyProof(root, tampered, proof, gindex) {
			t.Fatalf("tampered leaf should fail verification")
		}

		// Wrong gindex should fail
		if merkle.VerifyProof(root, chunks[i], proof, gindex+1) {
			t.Fatalf("wrong gindex should fail verification")
		}
	}
}
