package assembler

import (
	"fmt"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

// assembleTrap handles TRAP and TRAPV instructions.
// (RTE is now handled in flow.go with other return instructions.)
func assembleTrap(mn Mnemonic, operands []Operand, asm *Assembler) ([]uint16, error) {
	switch strings.ToLower(mn.Value) {
	case "trap":
		return assembleTrapImmediate(operands, asm)
	case "trapv":
		return assembleTrapv(operands)
	default:
		return nil, fmt.Errorf("unknown trap instruction: %s", mn.Value)
	}
}

// assembleTrapImmediate assembles TRAP #<vector>
// Valid vectors are 0–15.
func assembleTrapImmediate(operands []Operand, asm *Assembler) ([]uint16, error) {
	if len(operands) != 1 {
		return nil, fmt.Errorf("TRAP requires 1 operand (an immediate vector number)")
	}

	src := operands[0]
	if !src.IsImmediate() {
		return nil, fmt.Errorf("TRAP vector must be immediate (e.g., TRAP #3)")
	}

	val, err := parseConstant(src.Raw, asm)
	if err != nil {
		return nil, fmt.Errorf("invalid TRAP vector: %v", err)
	}
	if val < 0 || val > 15 {
		return nil, fmt.Errorf("TRAP vector must be between 0 and 15 (got %d)", val)
	}

	opword := uint16(cpu.OPTRAP) | uint16(val)
	return []uint16{opword}, nil
}

// assembleTrapv assembles the TRAPV instruction (trap on overflow).
func assembleTrapv(operands []Operand) ([]uint16, error) {
	if len(operands) != 0 {
		return nil, fmt.Errorf("TRAPV takes no operands")
	}
	return []uint16{cpu.OPTRAPV}, nil
}
