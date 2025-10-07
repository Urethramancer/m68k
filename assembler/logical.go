package assembler

import (
	"fmt"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

// assembleLogical handles AND, OR, EOR, and NOT instructions.
func assembleLogical(mn Mnemonic, operands []Operand, asm *Assembler) ([]uint16, error) {
	switch strings.ToLower(mn.Value) {
	case "and", "andi":
		return assembleAnd(mn, operands, asm)
	case "or", "ori":
		return assembleOr(mn, operands, asm)
	case "eor", "eori":
		return assembleEor(mn, operands, asm)
	case "not":
		return assembleNot(mn, operands)
	}
	return nil, fmt.Errorf("unknown logical instruction: %s", mn.Value)
}

func assembleAnd(mn Mnemonic, operands []Operand, asm *Assembler) ([]uint16, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("AND requires 2 operands")
	}
	src, dst := operands[0], operands[1]

	// Immediate variant: ANDI #imm, <ea>
	if src.IsImmediate() {
		opword := uint16(cpu.OPANDI)
		var err error
		opword, err = setOpwordSize(opword, mn.Size, SizeBitsSingleOp)
		if err != nil {
			return nil, err
		}

		eaBits, eaExt, err := encodeEA(dst)
		if err != nil {
			return nil, err
		}
		opword |= eaBits

		words := []uint16{opword}
		if len(src.ExtensionWords) > 0 {
			words = append(words, src.ExtensionWords...)
		}
		if len(eaExt) > 0 {
			words = append(words, eaExt...)
		}
		return words, nil
	}

	// Non-immediate AND Dn,<ea> or <ea>,Dn
	opword := uint16(cpu.OPAND)
	var err error
	opword, err = setOpwordSize(opword, mn.Size, SizeBits)
	if err != nil {
		return nil, err
	}

	var eaBits uint16
	var eaExt []uint16

	if dst.Mode == cpu.ModeData {
		// Direction: <ea> -> Dn
		opword |= (dst.Register << 9)
		eaBits, eaExt, err = encodeEA(src)
	} else {
		// Direction: Dn -> <ea>
		opword |= 0x0100
		opword |= (src.Register << 9)
		eaBits, eaExt, err = encodeEA(dst)
	}
	if err != nil {
		return nil, err
	}
	opword |= eaBits

	words := []uint16{opword}
	if len(eaExt) > 0 {
		words = append(words, eaExt...)
	}
	return words, nil
}

func assembleOr(mn Mnemonic, operands []Operand, asm *Assembler) ([]uint16, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("OR requires 2 operands")
	}
	src, dst := operands[0], operands[1]

	// Immediate variant: ORI #imm, <ea>
	if src.IsImmediate() {
		opword := uint16(cpu.OPORI)
		var err error
		opword, err = setOpwordSize(opword, mn.Size, SizeBitsSingleOp)
		if err != nil {
			return nil, err
		}

		eaBits, eaExt, err := encodeEA(dst)
		if err != nil {
			return nil, err
		}
		opword |= eaBits

		words := []uint16{opword}
		if len(src.ExtensionWords) > 0 {
			words = append(words, src.ExtensionWords...)
		}
		if len(eaExt) > 0 {
			words = append(words, eaExt...)
		}
		return words, nil
	}

	// Non-immediate OR Dn,<ea> or <ea>,Dn
	opword := uint16(cpu.OPOR)
	var err error
	opword, err = setOpwordSize(opword, mn.Size, SizeBits)
	if err != nil {
		return nil, err
	}

	var eaBits uint16
	var eaExt []uint16

	if dst.Mode == cpu.ModeData {
		// Direction: <ea> -> Dn
		opword |= (dst.Register << 9)
		eaBits, eaExt, err = encodeEA(src)
	} else {
		// Direction: Dn -> <ea>
		opword |= 0x0100
		opword |= (src.Register << 9)
		eaBits, eaExt, err = encodeEA(dst)
	}
	if err != nil {
		return nil, err
	}
	opword |= eaBits

	words := []uint16{opword}
	if len(eaExt) > 0 {
		words = append(words, eaExt...)
	}
	return words, nil
}

func assembleEor(mn Mnemonic, operands []Operand, asm *Assembler) ([]uint16, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("EOR requires 2 operands")
	}
	src, dst := operands[0], operands[1]

	// Immediate variant: EORI #imm, <ea>
	if src.IsImmediate() {
		opword := uint16(cpu.OPEORI)
		var err error
		opword, err = setOpwordSize(opword, mn.Size, SizeBitsSingleOp)
		if err != nil {
			return nil, err
		}

		eaBits, eaExt, err := encodeEA(dst)
		if err != nil {
			return nil, err
		}
		opword |= eaBits

		words := []uint16{opword}
		if len(src.ExtensionWords) > 0 {
			words = append(words, src.ExtensionWords...)
		}
		if len(eaExt) > 0 {
			words = append(words, eaExt...)
		}
		return words, nil
	}

	// Non-immediate EOR: must be Dn -> <ea>
	if src.Mode != cpu.ModeData {
		return nil, fmt.Errorf("source of EOR must be a data register")
	}

	opword := uint16(cpu.OPEOR)
	// EOR uses its own size encoding (not the same helper map)
	sz := mn.Size
	if sz == cpu.SizeInvalid {
		sz = cpu.SizeWord
	}
	switch sz {
	case cpu.SizeByte:
		opword |= 0x0000
	case cpu.SizeWord:
		opword |= 0x0040
	case cpu.SizeLong:
		opword |= 0x0080
	default:
		return nil, fmt.Errorf("unsupported size for EOR")
	}

	opword |= (src.Register << 9)

	eaBits, eaExt, err := encodeEA(dst)
	if err != nil {
		return nil, err
	}
	opword |= eaBits

	words := []uint16{opword}
	if len(eaExt) > 0 {
		words = append(words, eaExt...)
	}
	return words, nil
}

func assembleNot(mn Mnemonic, operands []Operand) ([]uint16, error) {
	if len(operands) != 1 {
		return nil, fmt.Errorf("NOT requires 1 operand")
	}
	dst := operands[0]

	opword := uint16(cpu.OPNOT)
	var err error
	opword, err = setOpwordSize(opword, mn.Size, SizeBitsSingleOp)
	if err != nil {
		return nil, err
	}

	eaBits, eaExt, err := encodeEA(dst)
	if err != nil {
		return nil, err
	}
	opword |= eaBits

	words := []uint16{opword}
	if len(eaExt) > 0 {
		words = append(words, eaExt...)
	}
	return words, nil
}
