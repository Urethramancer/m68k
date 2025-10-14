package cpu

import "encoding/binary"

// ReadU16 reads a big-endian 16-bit word from memory at the given address.
func (c *CPU) ReadU16(addr uint32) uint16 {
	return binary.BigEndian.Uint16(c.Mem[addr:])
}

// WriteU16 writes a 16-bit word to memory at the given address in big-endian format.
func (c *CPU) WriteU16(addr uint32, val uint16) {
	binary.BigEndian.PutUint16(c.Mem[addr:], val)
}

// ReadU32 reads a big-endian 32-bit long word from memory at the given address.
func (c *CPU) ReadU32(addr uint32) uint32 {
	return binary.BigEndian.Uint32(c.Mem[addr:])
}

// WriteU32 writes a 32-bit long word to memory at the given address in big-endian format.
func (c *CPU) WriteU32(addr uint32, val uint32) {
	binary.BigEndian.PutUint32(c.Mem[addr:], val)
}

// setNZ updates the N and Z flags in the SR based on a value and operation size.
func (c *CPU) setNZ(value uint32, size Size) {
	// Clear N and Z flags
	c.SR &^= (SRN | SRZ)

	var isZero bool
	var isNegative bool

	switch size {
	case SizeByte:
		val := int8(value)
		isZero = (val == 0)
		isNegative = (val < 0)
	case SizeWord:
		val := int16(value)
		isZero = (val == 0)
		isNegative = (val < 0)
	case SizeLong:
		val := int32(value)
		isZero = (val == 0)
		isNegative = (val < 0)
	}

	if isZero {
		c.SR |= SRZ
	}
	if isNegative {
		c.SR |= SRN
	}
}

// setFlagsArith sets the C, V, N, Z, and X flags based on an arithmetic operation.
// This is one of the most complex parts of CPU emulation.
func (c *CPU) setFlagsArith(src, dst, result uint32, size Size) {
	// Clear the flags first
	c.SR &^= (SRX | SRN | SRZ | SRV | SRC)

	var msbMask uint32
	var signMask uint32

	switch size {
	case SizeByte:
		msbMask = 0x80
		signMask = 0xFF
	case SizeWord:
		msbMask = 0x8000
		signMask = 0xFFFF
	case SizeLong:
		msbMask = 0x80000000
		signMask = 0xFFFFFFFF
	}

	// Sign bits of operands and result
	s := src & msbMask
	d := dst & msbMask
	r := result & msbMask

	// Zero flag (Z): Set if the result is zero.
	if (result & signMask) == 0 {
		c.SR |= SRZ
	}

	// Negative flag (N): Set if the most significant bit of the result is set.
	if r != 0 {
		c.SR |= SRN
	}

	// Carry flag (C): Set if a carry was generated from the most significant bit.
	// This happens if (src AND dst) OR (NOT result AND src) OR (NOT result AND dst) has its MSB set.
	if (s&d)|(^r&s)|(^r&d) != 0 {
		c.SR |= SRC
		c.SR |= SRX // Extend flag is always set with Carry
	}

	// Overflow flag (V): Set if the sign of the result is incorrect.
	// This happens if the source and destination signs are the same, but the result sign is different.
	if (s == d) && (s != r) {
		c.SR |= SRV
	}
}
