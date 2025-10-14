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
