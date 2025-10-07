package assembler

import (
	"fmt"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

func assembleMisc(mn Mnemonic, operands []Operand) ([]uint16, error) {
	switch strings.ToLower(mn.Value) {
	case "exg":
		return assembleExg(operands)
	case "stop":
		return assembleStop(operands)
	case "clr", "neg", "negx", "swap", "ext", "tas":
		return assembleMiscOneOp(mn, operands)
	case "reset", "nop", "illegal":
		return assembleMiscNoOp(mn, operands)
	default:
		return nil, fmt.Errorf("unknown misc instruction: %s", mn.Value)
	}
}

// --- STOP ---
func assembleStop(operands []Operand) ([]uint16, error) {
	if len(operands) != 1 {
		return nil, fmt.Errorf("STOP requires one immediate operand")
	}
	src := operands[0]
	if !src.IsImmediate() {
		return nil, fmt.Errorf("STOP operand must be immediate")
	}
	if len(src.ExtensionWords) == 0 {
		return nil, fmt.Errorf("missing 16-bit immediate for STOP")
	}
	return []uint16{cpu.OPSTOP, src.ExtensionWords[0]}, nil
}

// --- RESET / NOP / ILLEGAL ---
func assembleMiscNoOp(mn Mnemonic, operands []Operand) ([]uint16, error) {
	if len(operands) != 0 {
		return nil, fmt.Errorf("%s requires no operands", strings.ToUpper(mn.Value))
	}
	switch mn.Value {
	case "reset":
		return []uint16{cpu.OPRESET}, nil
	case "nop":
		return []uint16{cpu.OPNOP}, nil
	case "illegal":
		return []uint16{cpu.OPILLEGAL}, nil
	default:
		return nil, fmt.Errorf("unknown zero-operand misc instruction: %s", mn.Value)
	}
}

// --- EXG ---
func assembleExg(operands []Operand) ([]uint16, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("EXG requires 2 operands")
	}
	op1, op2 := operands[0], operands[1]
	opword := uint16(cpu.OPEXG)

	switch {
	case op1.Mode == cpu.ModeData && op2.Mode == cpu.ModeData:
		opword |= 0x0040
	case op1.Mode == cpu.ModeAddr && op2.Mode == cpu.ModeAddr:
		opword |= 0x0048
	case op1.Mode == cpu.ModeData && op2.Mode == cpu.ModeAddr:
		opword |= 0x0088
	case op1.Mode == cpu.ModeAddr && op2.Mode == cpu.ModeData:
		opword |= 0x0088
		op1, op2 = op2, op1
	default:
		return nil, fmt.Errorf("invalid operand combination for EXG")
	}
	opword |= (op1.Register << 9) | op2.Register
	return []uint16{opword}, nil
}

// --- One-operand instructions ---
func assembleMiscOneOp(mn Mnemonic, operands []Operand) ([]uint16, error) {
	if len(operands) != 1 {
		return nil, fmt.Errorf("%s requires 1 operand", strings.ToUpper(mn.Value))
	}
	dst := operands[0]
	var opword uint16
	var err error

	switch strings.ToLower(mn.Value) {
	case "clr":
		opword, err = setOpwordSize(cpu.OPCLR, mn.Size, SizeBitsSingleOp)
	case "neg":
		opword, err = setOpwordSize(cpu.OPNEG, mn.Size, SizeBitsSingleOp)
	case "negx":
		opword, err = setOpwordSize(cpu.OPNEGX, mn.Size, SizeBitsSingleOp)
	case "swap":
		if dst.Mode != cpu.ModeData {
			return nil, fmt.Errorf("SWAP requires a data register")
		}
		opword = cpu.OPSWAP | dst.Register
	case "ext":
		if dst.Mode != cpu.ModeData {
			return nil, fmt.Errorf("EXT requires a data register")
		}
		opword = cpu.OPEXT
		switch mn.Size {
		case cpu.SizeWord, cpu.SizeInvalid:
			opword |= 0x0080 // Sign-extend byte → word
		case cpu.SizeLong:
			opword |= 0x00C0 // Sign-extend word → long
		default:
			return nil, fmt.Errorf("EXT only supports .w and .l sizes")
		}
		opword |= dst.Register
	case "tas":
		opword = cpu.OPTAS
	default:
		return nil, fmt.Errorf("unsupported misc one-op instruction: %s", mn.Value)
	}

	if err != nil {
		return nil, err
	}

	eaBits, extWords, err := encodeEA(dst)
	if err != nil {
		return nil, fmt.Errorf("invalid addressing mode for %s: %v", mn.Value, err)
	}

	opword |= eaBits
	return append([]uint16{opword}, extWords...), nil
}
