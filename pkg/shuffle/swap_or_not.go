package shuffle

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
)

const (
	ShuffleRoundCount = 90
)

// ComputeShuffledIndex implements the Ethereum 2.0 Swap-or-Not shuffle algorithm.
// Maps index ∈ [0, indexCount) to its permuted position under seed.
func ComputeShuffledIndex(index uint64, indexCount uint64, seed [32]byte) (uint64, error) {
	if index >= indexCount {
		return 0, errors.New("index out of bounds")
	}
	if indexCount <= 1 {
		return index, nil
	}

	currIndex := index

	for round := uint8(0); round < ShuffleRoundCount; round++ {
		// Hash seed and round to determine pivot
		h := sha256.New()
		h.Write(seed[:])
		h.Write([]byte{round})
		pivotHash := h.Sum(nil)
		pivot := binary.LittleEndian.Uint64(pivotHash[:8]) % indexCount

		flip := (pivot + indexCount - currIndex) % indexCount
		position := currIndex
		if flip > position {
			position = flip
		}

		// Source hash for position bits
		var posBuf [4]byte
		binary.LittleEndian.PutUint32(posBuf[:], uint32(position/256))

		h2 := sha256.New()
		h2.Write(seed[:])
		h2.Write([]byte{round})
		h2.Write(posBuf[:])
		sourceHash := h2.Sum(nil)

		byteIndex := (position % 256) / 8
		bitIndex := position % 8
		bit := (sourceHash[byteIndex] >> bitIndex) & 1

		if bit == 1 {
			currIndex = flip
		}
	}

	return currIndex, nil
}

// ComputePermutation returns the fully shuffled array of indices from 0 to count-1.
func ComputePermutation(count uint64, seed [32]byte) ([]uint64, error) {
	shuffled := make([]uint64, count)
	for i := uint64(0); i < count; i++ {
		p, err := ComputeShuffledIndex(i, count, seed)
		if err != nil {
			return nil, err
		}
		shuffled[p] = i
	}
	return shuffled, nil
}

// ComputeCommittee extracts the committee members for a given slot and committee index.
func ComputeCommittee(activeIndices []uint64, seed [32]byte, committeeIndex uint64, totalCommittees uint64) ([]uint64, error) {
	totalActive := uint64(len(activeIndices))
	if totalCommittees == 0 || committeeIndex >= totalCommittees {
		return nil, errors.New("invalid committee parameters")
	}

	start := (totalActive * committeeIndex) / totalCommittees
	end := (totalActive * (committeeIndex + 1)) / totalCommittees

	committee := make([]uint64, end-start)
	for i := start; i < end; i++ {
		shuffledIdx, err := ComputeShuffledIndex(i, totalActive, seed)
		if err != nil {
			return nil, err
		}
		committee[i-start] = activeIndices[shuffledIdx]
	}

	return committee, nil
}
