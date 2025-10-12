package cpu

import (
	"encoding/binary"
)

// IsLittleEndianHost checks if the host system is little-endian.
func IsLittleEndianHost() bool {
	var x uint16 = 1
	b := [2]byte{}
	binary.LittleEndian.PutUint16(b[:], x)
	return b[0] == 1
}

// WordsToBytes converts a slice of 16-bit words to a big-endian byte slice.
func WordsToBytes(words []uint16) []byte {
	out := make([]byte, len(words)*2)
	for i, w := range words {
		binary.BigEndian.PutUint16(out[i*2:], w)
	}
	return out
}

// BytesToWords interprets bytes as big-endian 16-bit words.
// If an odd number of bytes is passed, the final byte is padded with 0.
func BytesToWords(b []byte) []uint16 {
	if len(b)%2 != 0 {
		b = append(b, 0)
	}
	out := make([]uint16, len(b)/2)
	for i := range out {
		out[i] = binary.BigEndian.Uint16(b[i*2:])
	}
	return out
}
