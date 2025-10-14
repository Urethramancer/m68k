package assembler

import (
	"fmt"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

// assembleAddressMode handles LEA and PEA instructions.
func (asm *Assembler) assembleAddressMode(mn Mnemonic, operands []Operand, pc uint32) ([]uint16, error) {
	switch strings.ToLower(mn.Value) {
	case "lea":
		return asm.assembleLea(operands)
	case "pea":
		return asm.assemblePea(operands)
	default:
		return nil, fmt.Errorf("unknown address mode instruction: %s", mn.Value)
	}
}

// assembleLea is now much simpler.
func (asm *Assembler) assembleLea(operands []Operand) ([]uint16, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("LEA requires 2 operands")
	}
	src, dst := operands[0], operands[1]

	if dst.Mode != cpu.ModeAddr {
		return nil, fmt.Errorf("destination of LEA must be an address register")
	}

	opword := uint16(cpu.OPLEA)
	opword |= (dst.Register << 9)

	eaBits, eaExt, err := asm.encodeEA(src, cpu.SizeLong)
	if err != nil {
		return nil, err
	}
	opword |= eaBits
	return append([]uint16{opword}, eaExt...), nil
}

// assemblePea is also simplified.
func (asm *Assembler) assemblePea(operands []Operand) ([]uint16, error) {
	if len(operands) != 1 {
		return nil, fmt.Errorf("PEA requires 1 operand")
	}
	src := operands[0]
	opword := uint16(cpu.OPPEA)

	eaBits, eaExt, err := asm.encodeEA(src, cpu.SizeLong)
	if err != nil {
		return nil, err
	}
	opword |= eaBits
	return append([]uint16{opword}, eaExt...), nil
}
