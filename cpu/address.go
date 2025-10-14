package cpu

import "fmt"

// GetOperand fetches a value using the specified addressing mode.
// This is the core of resolving the "source" part of an instruction.
func (c *CPU) GetOperand(mode, reg uint16, size Size) (uint32, error) {
	switch mode {
	case ModeData: // Data Register Direct
		val := c.D[reg]
		switch size {
		case SizeByte:
			return val & 0xFF, nil
		case SizeWord:
			return val & 0xFFFF, nil
		case SizeLong:
			return val, nil
		}
	case ModeAddr: // Address Register Direct
		val := c.A[reg]
		switch size {
		case SizeByte:
			return val & 0xFF, nil
		case SizeWord:
			return val & 0xFFFF, nil
		case SizeLong:
			return val, nil
		}
	case ModeAddrInd: // Address Register Indirect
		addr := c.A[reg]
		switch size {
		case SizeByte:
			return uint32(c.Mem[addr]), nil
		case SizeWord:
			return uint32(c.ReadU16(addr)), nil
		case SizeLong:
			return c.ReadU32(addr), nil
		}
	case ModeAddrPostInc: // Address Register Indirect with Postincrement
		addr := c.A[reg]
		increment := uint32(size.Bytes())
		// Byte operations on address registers (except A7) increment by 2
		if size == SizeByte && reg != 7 {
			increment = 2
		}
		c.A[reg] += increment

		switch size {
		case SizeByte:
			return uint32(c.Mem[addr]), nil
		case SizeWord:
			return uint32(c.ReadU16(addr)), nil
		case SizeLong:
			return c.ReadU32(addr), nil
		}
	case ModeAddrPreDec: // Address Register Indirect with Predecrement
		increment := uint32(size.Bytes())
		// Byte operations on address registers (except A7) increment by 2
		if size == SizeByte && reg != 7 {
			increment = 2
		}
		c.A[reg] -= increment
		addr := c.A[reg]

		switch size {
		case SizeByte:
			return uint32(c.Mem[addr]), nil
		case SizeWord:
			return uint32(c.ReadU16(addr)), nil
		case SizeLong:
			return c.ReadU32(addr), nil
		}
	case ModeAddrDisp: // Address Register Indirect with Displacement
		displacement := signExtend16(c.ReadU16(c.PC))
		c.PC += 2
		addr := uint32(int32(c.A[reg]) + displacement)
		switch size {
		case SizeByte:
			return uint32(c.Mem[addr]), nil
		case SizeWord:
			return uint32(c.ReadU16(addr)), nil
		case SizeLong:
			return c.ReadU32(addr), nil
		}
	case ModeOther: // Miscellaneous modes
		switch reg {
		case RegAbsShort: // Absolute Short
			addr := uint32(signExtend16(c.ReadU16(c.PC)))
			c.PC += 2
			switch size {
			case SizeByte:
				return uint32(c.Mem[addr]), nil
			case SizeWord:
				return uint32(c.ReadU16(addr)), nil
			case SizeLong:
				return c.ReadU32(addr), nil
			}
		case RegAbsLong: // Absolute Long
			addr := c.ReadU32(c.PC)
			c.PC += 4
			switch size {
			case SizeByte:
				return uint32(c.Mem[addr]), nil
			case SizeWord:
				return uint32(c.ReadU16(addr)), nil
			case SizeLong:
				return c.ReadU32(addr), nil
			}
		case RegImmediate: // Immediate
			var val uint32
			switch size {
			case SizeByte:
				// Byte immediates are stored as a word, high byte is ignored
				val = uint32(c.ReadU16(c.PC) & 0xFF)
				c.PC += 2
			case SizeWord:
				val = uint32(c.ReadU16(c.PC))
				c.PC += 2
			case SizeLong:
				val = c.ReadU32(c.PC)
				c.PC += 4
			}
			return val, nil
		default:
			return 0, fmt.Errorf("unimplemented source addressing sub-mode %d for mode %d", reg, mode)
		}
	default:
		return 0, fmt.Errorf("unimplemented source addressing mode %d", mode)
	}
	return 0, fmt.Errorf("invalid size for get operand")
}

// PutOperand writes a value using the specified addressing mode.
// This is the core of resolving the "destination" part of an instruction.
func (c *CPU) PutOperand(mode, reg uint16, size Size, value uint32) error {
	switch mode {
	case ModeData: // Data Register Direct
		switch size {
		case SizeByte:
			c.D[reg] = (c.D[reg] & 0xFFFFFF00) | (value & 0xFF)
		case SizeWord:
			c.D[reg] = (c.D[reg] & 0xFFFF0000) | (value & 0xFFFF)
		case SizeLong:
			c.D[reg] = value
		default:
			return fmt.Errorf("invalid size for put operand to D%d", reg)
		}
		return nil
	case ModeAddr: // Address Register Direct
		switch size {
		case SizeByte:
			return fmt.Errorf("invalid size .B for put operand to A%d", reg)
		case SizeWord:
			c.A[reg] = uint32(signExtend16(uint16(value)))
		case SizeLong:
			c.A[reg] = value
		default:
			return fmt.Errorf("invalid size for put operand to A%d", reg)
		}
		return nil
	case ModeAddrInd: // Address Register Indirect
		addr := c.A[reg]
		switch size {
		case SizeByte:
			c.Mem[addr] = byte(value & 0xFF)
		case SizeWord:
			c.WriteU16(addr, uint16(value&0xFFFF))
		case SizeLong:
			c.WriteU32(addr, value)
		default:
			return fmt.Errorf("invalid size for put operand to (A%d)", reg)
		}
		return nil
	case ModeAddrPostInc: // Address Register Indirect with Postincrement
		addr := c.A[reg]
		increment := uint32(size.Bytes())
		if size == SizeByte && reg != 7 {
			increment = 2
		}
		c.A[reg] += increment

		switch size {
		case SizeByte:
			c.Mem[addr] = byte(value & 0xFF)
		case SizeWord:
			c.WriteU16(addr, uint16(value&0xFFFF))
		case SizeLong:
			c.WriteU32(addr, value)
		default:
			return fmt.Errorf("invalid size for put operand to (A%d)+", reg)
		}
		return nil
	case ModeAddrPreDec: // Address Register Indirect with Predecrement
		increment := uint32(size.Bytes())
		if size == SizeByte && reg != 7 {
			increment = 2
		}
		c.A[reg] -= increment
		addr := c.A[reg]
		switch size {
		case SizeByte:
			c.Mem[addr] = byte(value & 0xFF)
		case SizeWord:
			c.WriteU16(addr, uint16(value&0xFFFF))
		case SizeLong:
			c.WriteU32(addr, value)
		default:
			return fmt.Errorf("invalid size for put operand to -(A%d)", reg)
		}
		return nil
	case ModeAddrDisp: // Address Register Indirect with Displacement
		// FIX: Do not advance PC here. It is handled by GetOperand.
		displacement := signExtend16(c.ReadU16(c.PC))
		addr := uint32(int32(c.A[reg]) + displacement)
		switch size {
		case SizeByte:
			c.Mem[addr] = byte(value & 0xFF)
		case SizeWord:
			c.WriteU16(addr, uint16(value&0xFFFF))
		case SizeLong:
			c.WriteU32(addr, value)
		default:
			return fmt.Errorf("invalid size for put operand to (d16,A%d)", reg)
		}
		return nil
	case ModeOther: // Miscellaneous modes
		switch reg {
		case RegAbsShort: // Absolute Short
			// FIX: Do not advance PC here.
			addr := uint32(signExtend16(c.ReadU16(c.PC)))
			switch size {
			case SizeByte:
				c.Mem[addr] = byte(value & 0xFF)
			case SizeWord:
				c.WriteU16(addr, uint16(value&0xFFFF))
			case SizeLong:
				c.WriteU32(addr, value)
			default:
				return fmt.Errorf("invalid size for put operand to (xxx).W")
			}
			return nil
		case RegAbsLong: // Absolute Long
			// FIX: Do not advance PC here.
			addr := c.ReadU32(c.PC)
			switch size {
			case SizeByte:
				c.Mem[addr] = byte(value & 0xFF)
			case SizeWord:
				c.WriteU16(addr, uint16(value&0xFFFF))
			case SizeLong:
				c.WriteU32(addr, value)
			default:
				return fmt.Errorf("invalid size for put operand to (xxx).L")
			}
			return nil
		default:
			return fmt.Errorf("invalid destination addressing sub-mode %d for mode %d", reg, mode)
		}
	default:
		return fmt.Errorf("unimplemented destination addressing mode %d", mode)
	}
}

// signExtend16 correctly sign-extends a 16-bit value to 32 bits.
func signExtend16(v uint16) int32 {
	return int32(int16(v))
}
