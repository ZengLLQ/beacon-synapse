package state

import (
	"github.com/ZengLLQ/beacon-synapse/pkg/ssz"
)

const (
	GweiPerEth                   uint64 = 1_000_000_000
	MaxEffectiveBalance          uint64 = 32 * GweiPerEth
	EffectiveBalanceIncrement    uint64 = 1 * GweiPerEth
	HysteresisQuotient           uint64 = 4
	HysteresisDownwardMultiplier uint64 = 1
	HysteresisUpwardMultiplier   uint64 = 5
	FarFutureEpoch               uint64 = ^uint64(0)
)

// Validator represents an Ethereum PoS validator record.
type Validator struct {
	Pubkey                     [48]byte
	WithdrawalCredentials      [32]byte
	EffectiveBalance           uint64
	Slashed                    bool
	ActivationEligibilityEpoch uint64
	ActivationEpoch            uint64
	ExitEpoch                  uint64
	WithdrawableEpoch          uint64
}

// NewValidator initializes an active validator with specified credentials.
func NewValidator(pubkey [48]byte, withdrawalCreds [32]byte, balance uint64) *Validator {
	eff := balance - (balance % EffectiveBalanceIncrement)
	if eff > MaxEffectiveBalance {
		eff = MaxEffectiveBalance
	}

	return &Validator{
		Pubkey:                     pubkey,
		WithdrawalCredentials:      withdrawalCreds,
		EffectiveBalance:           eff,
		Slashed:                    false,
		ActivationEligibilityEpoch: 0,
		ActivationEpoch:            0,
		ExitEpoch:                  FarFutureEpoch,
		WithdrawableEpoch:          FarFutureEpoch,
	}
}

// UpdateEffectiveBalance implements the hysteresis rule from the Ethereum Consensus spec.
func (v *Validator) UpdateEffectiveBalance(rawBalance uint64) {
	hysteresisIncrement := EffectiveBalanceIncrement / HysteresisQuotient
	downwardThreshold := hysteresisIncrement * HysteresisDownwardMultiplier
	upwardThreshold := hysteresisIncrement * HysteresisUpwardMultiplier

	if rawBalance+downwardThreshold < v.EffectiveBalance || v.EffectiveBalance+upwardThreshold < rawBalance {
		newEff := rawBalance - (rawBalance % EffectiveBalanceIncrement)
		if newEff > MaxEffectiveBalance {
			newEff = MaxEffectiveBalance
		}
		v.EffectiveBalance = newEff
	}
}

// HashTreeRoot computes the SSZ Merkle tree root for the Validator container (8 fields).
func (v *Validator) HashTreeRoot() [32]byte {
	chunks := make([][32]byte, 8)

	// 1. Pubkey: 48 bytes -> padded to 64 bytes (2 chunks), hashed to 1 chunk
	pkChunks := ssz.PackBytes(v.Pubkey[:])
	chunks[0] = ssz.MerkleizeChunks(pkChunks, 2)

	// 2. WithdrawalCredentials: 32 bytes (1 chunk)
	chunks[1] = v.WithdrawalCredentials

	// 3. EffectiveBalance: uint64 (1 chunk)
	chunks[2] = ssz.HashUint64(v.EffectiveBalance)

	// 4. Slashed: boolean (1 chunk)
	if v.Slashed {
		chunks[3][0] = 1
	}

	// 5. ActivationEligibilityEpoch: uint64
	chunks[4] = ssz.HashUint64(v.ActivationEligibilityEpoch)

	// 6. ActivationEpoch: uint64
	chunks[5] = ssz.HashUint64(v.ActivationEpoch)

	// 7. ExitEpoch: uint64
	chunks[6] = ssz.HashUint64(v.ExitEpoch)

	// 8. WithdrawableEpoch: uint64
	chunks[7] = ssz.HashUint64(v.WithdrawableEpoch)

	// Container Merkleization: 8 fields -> 8 leaves
	return ssz.MerkleizeChunks(chunks, 8)
}

// ComputeValidatorRegistryRoot computes the Merkle root of an entire validator registry.
func ComputeValidatorRegistryRoot(validators []*Validator) [32]byte {
	count := uint64(len(validators))
	if count == 0 {
		return [32]byte{}
	}

	chunks := make([][32]byte, count)
	for i, v := range validators {
		chunks[i] = v.HashTreeRoot()
	}

	// 1099511627776 is the historical max validator limit in Eth2 spec (2^40)
	treeRoot := ssz.MerkleizeChunks(chunks, 0)
	return ssz.MixInLength(treeRoot, count)
}
