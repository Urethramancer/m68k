package assembler

import (
	"fmt"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

// assembleAddressMode handles LEA and PEA instructions.
func assembleAddressMode(mn Mnemonic, operands []Operand, asm *Assembler, pc uint32) ([]uint16, error) {
	switch strings.ToLower(mn.Value) {
	case "lea":
		return assembleLea(operands, asm, pc)
	case "pea":
		return assemblePea(operands, asm, pc)
	default:
		return nil, fmt.Errorf("unknown address mode instruction: %s", mn.Value)
	}
}

func assembleLea(operands []Operand, asm *Assembler, pc uint32) ([]uint16, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("LEA requires 2 operands")
	}
	src, dst := operands[0], operands[1]

	if dst.Mode != cpu.ModeAddr {
		return nil, fmt.Errorf("destination of LEA must be an address register")
	}

	opword := uint16(cpu.OPLEA)
	opword |= (dst.Register << 9)

	// Prefer PC-relative for label operands
	if words, ok := tryResolveLabelEA(src, asm, pc, opword); ok {
		return words, nil
	}

	eaBits, eaExt, err := encodeEA(src)
	if err != nil {
		return nil, err
	}
	opword |= eaBits
	return append([]uint16{opword}, eaExt...), nil
}

func assemblePea(operands []Operand, asm *Assembler, pc uint32) ([]uint16, error) {
	if len(operands) != 1 {
		return nil, fmt.Errorf("PEA requires 1 operand")
	}
	src := operands[0]
	opword := uint16(cpu.OPPEA)

	if words, ok := tryResolveLabelEA(src, asm, pc, opword); ok {
		return words, nil
	}

	eaBits, eaExt, err := encodeEA(src)
	if err != nil {
		return nil, err
	}
	opword |= eaBits
	return append([]uint16{opword}, eaExt...), nil
}

// tryResolveLabelEA attempts to resolve a label operand to either a PC-relative
// or absolute long addressing mode. It returns the assembled words and true if
// successful, or nil and false if not.
func tryResolveLabelEA(src Operand, asm *Assembler, pc uint32, opword uint16) ([]uint16, bool) {
	if src.Mode != cpu.ModeOther {
		return nil, false
	}

	key := strings.ToLower(strings.TrimSpace(src.Label))
	if key == "" {
		key = strings.ToLower(strings.TrimSpace(src.Raw))
	}
	target, ok := asm.labels[key]
	if !ok {
		return nil, false
	}

	offset := int32(target) - int32(pc) - 2
	if offset >= -32768 && offset <= 32767 {
		eaBits := (cpu.ModeOther << 3) | cpu.ModePCRelative
		opword |= eaBits
		return []uint16{opword, uint16(int16(offset))}, true
	}

	eaBits := (cpu.ModeOther << 3) | cpu.ModeAbsLong
	opword |= eaBits
	return []uint16{opword, uint16(target >> 16), uint16(target)}, true
}
