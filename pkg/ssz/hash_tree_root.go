package ssz

import (
	"crypto/sha256"
	"encoding/binary"
	"math/bits"
)

const (
	BytesPerChunk = 32
)

// MerkleizeChunks hashes a list of 32-byte chunks into a single Merkle root.
// If limit is specified and greater than len(chunks), it pads with zero chunks up to the limit.
func MerkleizeChunks(chunks [][32]byte, limit uint64) [32]byte {
	count := uint64(len(chunks))
	leafCount := count
	if limit > leafCount {
		leafCount = limit
	}

	if leafCount == 0 {
		return [32]byte{}
	}

	// Calculate next power of 2
	depth := 0
	if leafCount > 1 {
		depth = 64 - bits.LeadingZeros64(leafCount-1)
	}
	targetLeaves := uint64(1) << depth

	tree := make([][32]byte, targetLeaves*2)

	// Fill bottom leaves
	for i := uint64(0); i < count; i++ {
		tree[targetLeaves+i] = chunks[i]
	}
	// Zero padding is automatically [32]byte{} for indices count to targetLeaves-1

	// Compute internal node hashes
	for i := targetLeaves - 1; i >= 1; i-- {
		h := sha256.New()
		h.Write(tree[2*i][:])
		h.Write(tree[2*i+1][:])
		copy(tree[i][:], h.Sum(nil))
	}

	return tree[1]
}

// PackUint64s serializes uint64 values into little-endian 32-byte chunks (4 uint64s per chunk).
func PackUint64s(values []uint64) [][32]byte {
	if len(values) == 0 {
		return nil
	}

	chunkCount := (len(values) + 3) / 4
	chunks := make([][32]byte, chunkCount)

	for i, val := range values {
		chunkIdx := i / 4
		offset := (i % 4) * 8
		binary.LittleEndian.PutUint64(chunks[chunkIdx][offset:offset+8], val)
	}

	return chunks
}

// PackBytes serializes raw byte slice into 32-byte chunks.
func PackBytes(data []byte) [][32]byte {
	if len(data) == 0 {
		return nil
	}

	chunkCount := (len(data) + BytesPerChunk - 1) / BytesPerChunk
	chunks := make([][32]byte, chunkCount)

	for i := 0; i < len(data); i += BytesPerChunk {
		end := i + BytesPerChunk
		if end > len(data) {
			end = len(data)
		}
		copy(chunks[i/BytesPerChunk][:], data[i:end])
	}

	return chunks
}

// MixInLength mixes the length of a dynamic list with its Merkle root.
// Returns sha256(root || uint64_to_bytes32_le(length)).
func MixInLength(root [32]byte, length uint64) [32]byte {
	var lenChunk [32]byte
	binary.LittleEndian.PutUint64(lenChunk[:8], length)

	h := sha256.New()
	h.Write(root[:])
	h.Write(lenChunk[:])

	var res [32]byte
	copy(res[:], h.Sum(nil))
	return res
}

// HashUint64 converts a uint64 into a 32-byte chunk with little-endian encoding.
func HashUint64(val uint64) [32]byte {
	var chunk [32]byte
	binary.LittleEndian.PutUint64(chunk[:8], val)
	return chunk
}
