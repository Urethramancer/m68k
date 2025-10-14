package assembler

import (
	"fmt"

	"github.com/Urethramancer/m68k/cpu"
)

// assembleMovep assembles the MOVEP instruction.
// Syntax:
//
//	MOVEP Dx, d(An)   ; Register → Memory
//	MOVEP d(An), Dx   ; Memory → Register
//
// Only supports address-displacement modes (d16,An).
func (asm *Assembler) assembleMovep(mn Mnemonic, operands []Operand) ([]uint16, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("MOVEP requires 2 operands")
	}

	src, dst := operands[0], operands[1]
	opword := uint16(cpu.OPMOVEP)
	var dataReg, addrReg, disp uint16

	switch {
	// Register → Memory
	case src.Mode == cpu.ModeData && dst.Mode == cpu.ModeAddrDisp:
		opword |= 0x0180 // direction bit set (to memory)
		dataReg = src.Register
		addrReg = dst.Register
		if len(dst.ExtensionWords) == 0 {
			return nil, fmt.Errorf("missing displacement for MOVEP destination")
		}
		disp = dst.ExtensionWords[0]

	// Memory → Register
	case src.Mode == cpu.ModeAddrDisp && dst.Mode == cpu.ModeData:
		opword |= 0x0100 // direction bit clear (to register)
		dataReg = dst.Register
		addrReg = src.Register
		if len(src.ExtensionWords) == 0 {
			return nil, fmt.Errorf("missing displacement for MOVEP source")
		}
		disp = src.ExtensionWords[0]

	default:
		return nil, fmt.Errorf("invalid operand combination for MOVEP — must be Dx,d(An) or d(An),Dx")
	}

	// Size (.W or .L)
	switch mn.Size {
	case cpu.SizeWord:
		// default
	case cpu.SizeLong:
		opword |= 0x0040
	default:
		return nil, fmt.Errorf("MOVEP only supports .W or .L sizes")
	}

	// Register fields
	opword |= (dataReg << 9)
	opword |= addrReg

	// Ensure displacement is treated as signed 16-bit
	disp = uint16(int16(disp))

	return []uint16{opword, disp}, nil
}
