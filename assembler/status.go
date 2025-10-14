package assembler

import (
	"fmt"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

// assembleStatus handles instructions involving SR, CCR, and USP.
func (asm *Assembler) assembleStatus(mn Mnemonic, operands []Operand) ([]uint16, error) {
	if len(operands) == 0 {
		return nil, fmt.Errorf("%s requires at least one operand", strings.ToUpper(mn.Value))
	}

	op1 := operands[0]
	var op2 Operand
	if len(operands) > 1 {
		op2 = operands[1]
	}

	switch strings.ToLower(mn.Value) {

	// MOVE SR/CCR/USP variants
	case "move":
		switch {
		// MOVE <ea>, SR
		case strings.EqualFold(op2.Raw, "sr"):
			return asm.assembleMoveToSr(op1)

		// MOVE <ea>, CCR
		case strings.EqualFold(op2.Raw, "ccr"):
			return asm.assembleMoveToCcr(op1)

		// MOVE SR, <ea>
		case strings.EqualFold(op1.Raw, "sr"):
			return asm.assembleMoveFromSr(op2)

		case strings.EqualFold(op1.Raw, "ccr"):
			// Assembles identically to MOVE from SR, the user just needs to mask the bits
			return asm.assembleMoveFromSr(op2)

		// MOVE <ea>, USP
		case strings.EqualFold(op2.Raw, "usp"):
			return assembleMoveToUsp(op1)

		// MOVE USP, <ea>
		case strings.EqualFold(op1.Raw, "usp"):
			return assembleMoveFromUsp(op2)

		default:
			return nil, fmt.Errorf("invalid MOVE combination for status registers")
		}

	// ANDI to SR/CCR
	case "andi":
		if strings.EqualFold(op2.Raw, "sr") {
			return assembleLogicImmediateToSr(cpu.OPANDItoSR, op1, "ANDI")
		}
		return assembleLogicImmediateToSr(cpu.OPANDItoCCR, op1, "ANDI")

	// ORI to SR/CCR
	case "ori":
		if strings.EqualFold(op2.Raw, "sr") {
			return assembleLogicImmediateToSr(cpu.OPORItoSR, op1, "ORI")
		}
		return assembleLogicImmediateToSr(cpu.OPORItoCCR, op1, "ORI")

	// EORI to SR/CCR
	case "eori":
		if strings.EqualFold(op2.Raw, "sr") {
			return assembleLogicImmediateToSr(cpu.OPEORItoSR, op1, "EORI")
		}
		return assembleLogicImmediateToSr(cpu.OPEORItoCCR, op1, "EORI")
	}

	return nil, fmt.Errorf("unknown status register instruction: %s", mn.Value)
}

// MOVE <ea>, SR
func (asm *Assembler) assembleMoveToSr(src Operand) ([]uint16, error) {
	eaBits, eaExt, err := asm.encodeEA(src, cpu.SizeWord)
	if err != nil {
		return nil, err
	}
	opword := uint16(cpu.OPMOVEToSR)
	opword |= eaBits
	return append([]uint16{opword}, eaExt...), nil
}

// MOVE <ea>, CCR
func (asm *Assembler) assembleMoveToCcr(src Operand) ([]uint16, error) {
	eaBits, eaExt, err := asm.encodeEA(src, cpu.SizeWord)
	if err != nil {
		return nil, err
	}
	opword := uint16(cpu.OPMOVEToCCR)
	opword |= eaBits
	return append([]uint16{opword}, eaExt...), nil
}

// MOVE SR, <ea>
func (asm *Assembler) assembleMoveFromSr(dst Operand) ([]uint16, error) {
	eaBits, eaExt, err := asm.encodeEA(dst, cpu.SizeWord)
	if err != nil {
		return nil, err
	}
	opword := uint16(cpu.OPMOVEFromSR)
	opword |= eaBits
	return append([]uint16{opword}, eaExt...), nil
}

// MOVE An, USP
func assembleMoveToUsp(src Operand) ([]uint16, error) {
	if src.Mode != cpu.ModeAddr {
		return nil, fmt.Errorf("source for MOVE to USP must be an address register (An)")
	}
	opword := uint16(cpu.OPMOVEToUSP)
	opword |= src.Register
	return []uint16{opword}, nil
}

// MOVE USP, An
func assembleMoveFromUsp(dst Operand) ([]uint16, error) {
	if dst.Mode != cpu.ModeAddr {
		return nil, fmt.Errorf("destination for MOVE from USP must be an address register (An)")
	}
	opword := uint16(cpu.OPMOVEFromUSP)
	opword |= dst.Register
	return []uint16{opword}, nil
}

// ANDI/ORI/EORI to SR or CCR
// These instructions operate only with an immediate source operand.
// e.g.  ANDI #$2700,SR  or  EORI #$FF,CCR
func assembleLogicImmediateToSr(baseOpcode uint16, src Operand, opname string) ([]uint16, error) {
	if !src.IsImmediate() {
		return nil, fmt.Errorf("%s requires an immediate source operand", opname)
	}

	if len(src.ExtensionWords) == 0 {
		return nil, fmt.Errorf("%s missing immediate data", opname)
	}

	// Build final word sequence: [opcode][immediate extension(s)]
	return append([]uint16{baseOpcode}, src.ExtensionWords...), nil
}
