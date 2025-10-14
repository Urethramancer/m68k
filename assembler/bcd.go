package assembler

import (
	"fmt"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

// assembleBcd handles ABCD, SBCD, and NBCD instructions.
func (asm *Assembler) assembleBcd(mn Mnemonic, operands []Operand) ([]uint16, error) {
	switch strings.ToLower(mn.Value) {
	case "abcd":
		return asm.assembleAbcdSbcd(true, operands)
	case "sbcd":
		return asm.assembleAbcdSbcd(false, operands)
	case "nbcd":
		return asm.assembleNbcd(operands)
	}
	return nil, fmt.Errorf("unknown BCD instruction: %s", mn.Value)
}

// assembleAbcdSbcd assembles ABCD and SBCD instructions.
//
// Encoding:
//   - Register-to-register:  1100|Dst|1000|000|Src   (ABCD Dx,Dy / SBCD Dx,Dy)
//   - Memory (predecrement): 1100|Dst|1000|001|Src   (ABCD -(Ax),-(Ay) / SBCD -(Ax),-(Ay))
func (asm *Assembler) assembleAbcdSbcd(isAdd bool, operands []Operand) ([]uint16, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("ABCD/SBCD require 2 operands")
	}
	src, dst := operands[0], operands[1]

	var opword uint16
	if isAdd {
		opword = 0xC100 // Base opcode for ABCD
	} else {
		opword = 0x8100 // Base opcode for SBCD
	}

	switch {
	// Register-to-register form (Dx, Dy)
	case src.Mode == cpu.ModeData && dst.Mode == cpu.ModeData:
		opword |= (dst.Register << 9) | src.Register

	// Memory predecrement form (-(Ax), -(Ay))
	case src.Mode == cpu.ModeAddrPreDec && dst.Mode == cpu.ModeAddrPreDec:
		opword |= (dst.Register << 9) | (1 << 3) | src.Register

	default:
		return nil, fmt.Errorf("invalid operand combination for ABCD/SBCD: %s, %s", src.Raw, dst.Raw)
	}

	return []uint16{opword}, nil
}

// assembleNbcd assembles the NBCD instruction.
//
// Encoding:
//
//	0100 1000 00 | <EA>
//	- Only supports memory destination or data register.
func (asm *Assembler) assembleNbcd(operands []Operand) ([]uint16, error) {
	if len(operands) != 1 {
		return nil, fmt.Errorf("NBCD requires 1 operand")
	}

	dst := operands[0]
	opword := uint16(cpu.OPNBCD)

	eaBits, eaExt, err := asm.encodeEA(dst, cpu.SizeByte)
	if err != nil {
		return nil, err
	}
	opword |= eaBits

	return append([]uint16{opword}, eaExt...), nil
}
