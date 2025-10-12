package assembler

import (
	"fmt"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

// assembleStack handles LINK and UNLK stack operations.
func assembleStack(mn Mnemonic, operands []Operand, asm *Assembler) ([]uint16, error) {
	switch strings.ToLower(mn.Value) {
	case "link":
		return assembleLink(operands)
	case "unlk":
		return assembleUnlk(operands)
	default:
		return nil, fmt.Errorf("unknown stack instruction: %s", mn.Value)
	}
}

// LINK

// assembleLink assembles the LINK instruction.
// Syntax: LINK An, #<displacement>
// Example: LINK A6, #-32
// Opcode: 0100 1110 0101 0 | An
func assembleLink(operands []Operand) ([]uint16, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("LINK requires 2 operands: (An, #<displacement>)")
	}

	regOp, immOp := operands[0], operands[1]

	if regOp.Mode != cpu.ModeAddr {
		return nil, fmt.Errorf("first operand of LINK must be an address register (An)")
	}
	if !immOp.IsImmediate() {
		return nil, fmt.Errorf("second operand of LINK must be an immediate displacement")
	}
	if len(immOp.ExtensionWords) != 1 {
		return nil, fmt.Errorf("LINK displacement must be a 16-bit value")
	}

	opword := uint16(cpu.OPLINK)
	opword |= regOp.Register

	return []uint16{opword, immOp.ExtensionWords[0]}, nil
}

// UNLK

// assembleUnlk assembles the UNLK instruction.
// Syntax: UNLK An
// Example: UNLK A6
// Opcode: 0100 1110 0101 1 | An
func assembleUnlk(operands []Operand) ([]uint16, error) {
	if len(operands) != 1 {
		return nil, fmt.Errorf("UNLK requires 1 operand: (An)")
	}

	regOp := operands[0]
	if regOp.Mode != cpu.ModeAddr {
		return nil, fmt.Errorf("operand of UNLK must be an address register (An)")
	}

	opword := uint16(cpu.OPUNLK)
	opword |= regOp.Register

	return []uint16{opword}, nil
}
