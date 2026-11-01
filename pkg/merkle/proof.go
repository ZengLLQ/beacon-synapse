package merkle

import (
	"crypto/sha256"
	"errors"
	"math/bits"
)

// VerifyProof verifies a Merkle branch proof against a generalized index and root.
func VerifyProof(root [32]byte, leaf [32]byte, proof [][32]byte, gindex uint64) bool {
	if gindex == 0 {
		return false
	}

	current := leaf
	idx := gindex

	for _, sibling := range proof {
		if idx <= 1 {
			return false
		}

		h := sha256.New()
		if idx%2 == 1 {
			// Current is right child, sibling is left child
			h.Write(sibling[:])
			h.Write(current[:])
		} else {
			// Current is left child, sibling is right child
			h.Write(current[:])
			h.Write(sibling[:])
		}
		copy(current[:], h.Sum(nil))
		idx /= 2
	}

	return idx == 1 && current == root
}

// GenerateProof produces a Merkle branch proof for a leaf in a chunk list.
// The proof is ordered from leaf sibling up to the level below the root.
func GenerateProof(leaves [][32]byte, leafIndex uint64) ([][32]byte, uint64, error) {
	count := uint64(len(leaves))
	if leafIndex >= count {
		return nil, 0, errors.New("leaf index out of range")
	}

	if count == 0 {
		return nil, 0, errors.New("empty leaves")
	}

	// Compute depth and total leaves (next power of 2)
	depth := 0
	if count > 1 {
		depth = 64 - bits.LeadingZeros64(count-1)
	}
	targetLeaves := uint64(1) << depth

	tree := make([][32]byte, targetLeaves*2)
	for i := uint64(0); i < count; i++ {
		tree[targetLeaves+i] = leaves[i]
	}

	// Internal nodes
	for i := targetLeaves - 1; i >= 1; i-- {
		h := sha256.New()
		h.Write(tree[2*i][:])
		h.Write(tree[2*i+1][:])
		copy(tree[i][:], h.Sum(nil))
	}

	// Generalized index of target leaf
	gindex := targetLeaves + leafIndex

	proof := make([][32]byte, depth)
	curr := gindex
	for d := 0; d < depth; d++ {
		sibling := curr ^ 1
		proof[d] = tree[sibling]
		curr /= 2
	}

	return proof, gindex, nil
}

// ComputeGIndex returns the generalized index for a leaf at a given depth and index.
func ComputeGIndex(depth int, index uint64) uint64 {
	return (uint64(1) << depth) + index
}
