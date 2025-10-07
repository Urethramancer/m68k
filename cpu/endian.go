package cpu

import (
	"encoding/binary"
)

// WordsToBytes converts a slice of 16-bit words into a byte slice in big-endian format.
// This is the standard serialization format for M68k machine code.
func WordsToBytes(words []uint16) []byte {
	b := make([]byte, len(words)*2)
	for i, w := range words {
		binary.BigEndian.PutUint16(b[i*2:], w)
	}
	return b
}

// BytesToWords converts a slice of bytes in big-endian format into a slice of 16-bit words.
// It assumes the byte slice has an even length.
func BytesToWords(b []byte) []uint16 {
	if len(b)&1 != 0 {
		b = append(b, 0)
	}
	n := len(b) / 2
	words := make([]uint16, n)
	for i := 0; i < n; i++ {
		words[i] = binary.BigEndian.Uint16(b[i*2:])
	}
	return words
}
