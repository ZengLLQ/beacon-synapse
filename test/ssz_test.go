package test

import (
	"crypto/sha256"
	"testing"

	"github.com/ZengLLQ/beacon-synapse/pkg/ssz"
)

func TestSSZMerkleizeChunks(t *testing.T) {
	// 2 chunks: leaf0, leaf1
	var c0, c1 [32]byte
	c0[0] = 0xAA
	c1[0] = 0xBB

	chunks := [][32]byte{c0, c1}
	root := ssz.MerkleizeChunks(chunks, 2)

	h := sha256.New()
	h.Write(c0[:])
	h.Write(c1[:])
	expected := h.Sum(nil)

	for i := 0; i < 32; i++ {
		if root[i] != expected[i] {
			t.Fatalf("root mismatch at byte %d: got %x, expected %x", i, root[i], expected[i])
		}
	}
}

func TestSSZPackUint64s(t *testing.T) {
	vals := []uint64{1, 2, 3, 4, 5}
	chunks := ssz.PackUint64s(vals)

	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks for 5 uint64s, got %d", len(chunks))
	}
}

func TestSSZMixInLength(t *testing.T) {
	var root [32]byte
	root[0] = 0x01
	mixed := ssz.MixInLength(root, 100)

	if mixed == root {
		t.Fatal("mixed root should differ from base root")
	}
}
