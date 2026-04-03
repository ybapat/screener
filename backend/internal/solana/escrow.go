package solana

import (
	"crypto/sha256"
	"encoding/binary"
)

// anchorDiscriminator computes the 8-byte Anchor instruction discriminator
// for a given instruction name: sha256("global:<name>")[:8].
func anchorDiscriminator(name string) [8]byte {
	h := sha256.Sum256([]byte("global:" + name))
	var disc [8]byte
	copy(disc[:], h[:8])
	return disc
}

// putUint64LE writes a uint64 in little-endian format.
func putUint64LE(b []byte, v uint64) {
	binary.LittleEndian.PutUint64(b, v)
}

// BuildDepositData builds the instruction data for the escrow deposit instruction.
// Format: [8-byte discriminator][8-byte amount LE][16-byte dataset_id]
func BuildDepositData(amount uint64, datasetID [16]byte) []byte {
	data := make([]byte, 8+8+16)
	disc := anchorDiscriminator("deposit")
	copy(data[:8], disc[:])
	putUint64LE(data[8:16], amount)
	copy(data[16:32], datasetID[:])
	return data
}
