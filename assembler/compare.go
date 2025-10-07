package assembler

import (
	"fmt"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

// assembleCompare handles CMP, CMPA, CMPI, TST, and CHK instructions.
func assembleCompare(mn Mnemonic, operands []Operand, asm *Assembler) ([]uint16, error) {
	switch strings.ToLower(mn.Value) {
	case "cmp", "cmpa", "cmpi":
		return assembleCmpFamily(mn, operands, asm)
	case "tst":
		return assembleTst(mn, operands, asm)
	case "chk":
		return assembleChk(operands, asm)
	default:
		return nil, fmt.Errorf("unknown compare instruction: %s", mn.Value)
	}
}

// --- CMP / CMPA / CMPI ---

func assembleCmpFamily(mn Mnemonic, operands []Operand, asm *Assembler) ([]uint16, error) {
	name := strings.ToLower(mn.Value)

	if len(operands) != 2 {
		return nil, fmt.Errorf("%s requires 2 operands", strings.ToUpper(name))
	}
	src, dst := operands[0], operands[1]

	switch name {
	case "cmp":
		return assembleCmp(mn, src, dst, asm)
	case "cmpa":
		return assembleCmpa(mn, src, dst, asm)
	case "cmpi":
		return assembleCmpi(mn, src, dst, asm)
	default:
		return nil, fmt.Errorf("unhandled compare type: %s", name)
	}
}

// CMP: 1011 Dn Sz <ea>
func assembleCmp(mn Mnemonic, src, dst Operand, asm *Assembler) ([]uint16, error) {
	if dst.Mode != cpu.ModeData {
		return nil, fmt.Errorf("CMP destination must be a data register")
	}

	opword := uint16(cpu.OPCMP)
	opword, err := setOpwordSize(opword, mn.Size, SizeBits)
	if err != nil {
		return nil, err
	}

	opword |= dst.Register << 9

	eaBits, ext, err := encodeEA(src)
	if err != nil {
		return nil, err
	}
	opword |= eaBits

	return append([]uint16{opword}, ext...), nil
}

// CMPA: 1011 An 11 Sz <ea>
func assembleCmpa(mn Mnemonic, src, dst Operand, asm *Assembler) ([]uint16, error) {
	if dst.Mode != cpu.ModeAddr {
		return nil, fmt.Errorf("CMPA destination must be an address register")
	}

	opword := uint16(cpu.OPCMPA)
	opword, err := setOpwordSize(opword, mn.Size, SizeBitsAddr)
	if err != nil {
		return nil, err
	}

	opword |= dst.Register << 9

	eaBits, ext, err := encodeEA(src)
	if err != nil {
		return nil, err
	}
	opword |= eaBits

	return append([]uint16{opword}, ext...), nil
}

// CMPI: 0000 1100 Sz <ea>
func assembleCmpi(mn Mnemonic, src, dst Operand, asm *Assembler) ([]uint16, error) {
	if !src.IsImmediate() {
		return nil, fmt.Errorf("CMPI source must be immediate")
	}

	opword := uint16(cpu.OPCMPI)
	opword, err := setOpwordSize(opword, mn.Size, SizeBitsSingleOp)
	if err != nil {
		return nil, err
	}

	eaBits, eaExt, err := encodeEA(dst)
	if err != nil {
		return nil, err
	}
	opword |= eaBits

	// Combine: opcode + immediate + EA extensions
	words := []uint16{opword}
	words = append(words, src.ExtensionWords...)
	words = append(words, eaExt...)

	return words, nil
}

// TST: 0100 1010 Sz <ea>
func assembleTst(mn Mnemonic, operands []Operand, asm *Assembler) ([]uint16, error) {
	if len(operands) != 1 {
		return nil, fmt.Errorf("TST requires 1 operand")
	}
	op := operands[0]

	if op.Mode == cpu.ModeAddr {
		return nil, fmt.Errorf("TST cannot test an address register directly")
	}

	opword := uint16(cpu.OPTST)
	opword, err := setOpwordSize(opword, mn.Size, SizeBits)
	if err != nil {
		return nil, err
	}

	eaBits, ext, err := encodeEA(op)
	if err != nil {
		return nil, err
	}
	opword |= eaBits

	return append([]uint16{opword}, ext...), nil
}

// CHK: 0100 1000 <ea>, Dn
func assembleChk(operands []Operand, asm *Assembler) ([]uint16, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("CHK requires 2 operands (<ea>,Dn)")
	}
	src, dst := operands[0], operands[1]

	if dst.Mode != cpu.ModeData {
		return nil, fmt.Errorf("CHK destination must be a data register")
	}

	opword := uint16(cpu.OPCHK)
	opword |= dst.Register << 9

	eaBits, ext, err := encodeEA(src)
	if err != nil {
		return nil, err
	}
	opword |= eaBits

	return append([]uint16{opword}, ext...), nil
}
