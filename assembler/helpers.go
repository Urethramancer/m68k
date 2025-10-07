package assembler

import (
	"fmt"
	"math/bits"

	"github.com/Urethramancer/m68k/cpu"
)

//
// Size-bit lookup tables
//

var (
	commonSizeBits = map[cpu.Size]uint16{
		cpu.SizeByte: 0x0000,
		cpu.SizeWord: 0x0040,
		cpu.SizeLong: 0x0080,
	}

	// SizeBits is for dual-operand instructions (ADD, SUB, OR, AND, etc.)
	SizeBits = commonSizeBits
	// SizeBitsSingleOp is for single-operand instructions (CLR, NEG, NOT, etc.)
	SizeBitsSingleOp = commonSizeBits
	// SizeBitsTst is for the TST instruction (same pattern as dual-operand instructions).
	SizeBitsTst = commonSizeBits

	// SizeBitsAddr is for ADDA, SUBA, CMPA (address arithmetic)
	SizeBitsAddr = map[cpu.Size]uint16{
		cpu.SizeWord: 0x00C0,
		cpu.SizeLong: 0x01C0,
	}
)

// setOpwordSize applies the size field to an opcode
func setOpwordSize(opword uint16, size cpu.Size, sizeMap map[cpu.Size]uint16) (uint16, error) {
	if size == cpu.SizeInvalid {
		size = cpu.SizeWord
	}
	bits, ok := sizeMap[size]
	if !ok {
		return 0, fmt.Errorf("unsupported size for this instruction")
	}
	return opword | bits, nil
}

func encodeEA(op Operand) (uint16, []uint16, error) {
	var word uint16
	var exts []uint16

	switch op.Mode {
	case cpu.ModeData: // Dn
		word = (cpu.ModeData << 3) | op.Register

	case cpu.ModeAddr: // An
		word = (cpu.ModeAddr << 3) | op.Register

	case cpu.ModeAddrInd: // (An)
		word = (cpu.ModeAddrInd << 3) | op.Register

	case cpu.ModeAddrPostInc: // (An)+
		word = (cpu.ModeAddrPostInc << 3) | op.Register

	case cpu.ModeAddrPreDec: // -(An)
		word = (cpu.ModeAddrPreDec << 3) | op.Register

	case cpu.ModeAddrDisp: // (d16,An)
		word = (cpu.ModeAddrDisp << 3) | op.Register
		if len(op.ExtensionWords) > 0 {
			exts = append(exts, op.ExtensionWords...)
		}

	case cpu.ModeAddrIndex: // (d8,An,Xn)
		word = (cpu.ModeAddrIndex << 3) | op.Register
		if len(op.ExtensionWords) > 0 {
			exts = append(exts, op.ExtensionWords...)
		}

	case cpu.ModeOther:
		switch op.Register {
		case cpu.ModeAbsShort: // (xxx).W
			word = (cpu.ModeOther << 3) | cpu.ModeAbsShort
			exts = append(exts, op.ExtensionWords...)

		case cpu.ModeAbsLong: // (xxx).L
			word = (cpu.ModeOther << 3) | cpu.ModeAbsLong
			exts = append(exts, op.ExtensionWords...)

		case cpu.ModePCRelative: // (d16,PC)
			word = (cpu.ModeOther << 3) | cpu.ModePCRelative
			// PC-relative uses exactly one 16-bit displacement. Use label placeholder
			// if present; if not, safely append 0 so assembler can patch later.
			if len(op.ExtensionWords) > 0 {
				exts = append(exts, op.ExtensionWords[0])
			} else {
				exts = append(exts, 0)
			}

		case cpu.ModePCIndex: // (d8,PC,Xn)
			word = (cpu.ModeOther << 3) | 3 // 111 011
			exts = append(exts, op.ExtensionWords...)

		case cpu.ModeImmediate: // #<data>
			word = (cpu.ModeOther << 3) | cpu.ModeImmediate
			exts = append(exts, op.ExtensionWords...)

		default:
			return 0, nil, fmt.Errorf("invalid ModeOther subtype: %d", op.Register)
		}

	default:
		return 0, nil, fmt.Errorf("unsupported addressing mode: %d", op.Mode)
	}

	return word, exts, nil
}

// reverseMovemMask reverses MOVEM bit ordering correctly.
// Each 8-bit register group (D0–D7, A0–A7) is reversed independently.
func reverseMovemMask(mask uint16) uint16 {
	d := uint8(mask & 0xFF)
	a := uint8((mask >> 8) & 0xFF)
	d = bits.Reverse8(d)
	a = bits.Reverse8(a)
	return uint16(a)<<8 | uint16(d)
}
