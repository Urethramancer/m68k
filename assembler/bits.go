package assembler

import (
	"fmt"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

//
// Bitwise Instruction Tables
//

// ShiftRotateType contains shift/rotate opcode type bits (added to base opcode).
var ShiftRotateType = map[string]uint16{
	"asr": 0x0000, "asl": 0x0100,
	"lsr": 0x0008, "lsl": 0x0108,
	"ror": 0x0018, "rol": 0x0118,
	"roxr": 0x0010, "roxl": 0x0110, // Add these
}

// BitwiseSize contains size bits for shift/rotate register forms.
var BitwiseSize = map[cpu.Size]uint16{
	cpu.SizeByte: 0x0000,
	cpu.SizeWord: 0x0040,
	cpu.SizeLong: 0x0080,
}

//
// Bitwise Instruction Dispatcher
//

// assembleBitwise handles all shift, rotate, and bit manipulation instructions.
func (asm *Assembler) assembleBitwise(mn Mnemonic, operands []Operand) ([]uint16, error) {
	switch strings.ToLower(mn.Value) {
	case "asl", "asr", "lsl", "lsr", "rol", "ror":
		return asm.assembleShiftRotate(mn, operands)
	case "btst", "bset", "bclr", "bchg":
		return asm.assembleBitManipulation(mn, operands)
	default:
		return nil, fmt.Errorf("unknown bitwise instruction: %s", mn.Value)
	}
}

//
// Shift / Rotate
//

// assembleShiftRotate encodes ASL/ASR, LSL/LSR, ROL/ROR.
// Supports both register and memory forms:
//
//	Register form: <op> #imm,Dy  or  <op> Dx,Dy
//	Memory form:   <op> <ea>     (always word-sized)
func (asm *Assembler) assembleShiftRotate(mn Mnemonic, operands []Operand) ([]uint16, error) {
	opword := uint16(cpu.OPShiftRotateBase)
	opword |= ShiftRotateType[mn.Value]

	switch len(operands) {
	// Memory form
	case 1:
		if mn.Size != cpu.SizeWord && mn.Size != 0 {
			return nil, fmt.Errorf("%s on memory must be word-sized", mn.Value)
		}
		opword |= 0x00C0 // Set memory form bits
		dst := operands[0]

		eaBits, ext, err := asm.encodeEA(dst, cpu.SizeWord)
		if err != nil {
			return nil, err
		}
		opword |= eaBits
		return append([]uint16{opword}, ext...), nil

	// Register form
	case 2:
		src, dst := operands[0], operands[1]
		if dst.Mode != cpu.ModeData {
			return nil, fmt.Errorf("destination of %s must be a data register", mn.Value)
		}

		opword |= dst.Register // bits 2–0 = destination
		opword, err := setOpwordSize(opword, mn.Size, BitwiseSize)
		if err != nil {
			return nil, err
		}

		if src.IsImmediate() {
			count, _ := asm.parseConstant(src.Raw)
			if count < 1 || count > 8 {
				return nil, fmt.Errorf("immediate shift/rotate count must be between 1 and 8")
			}
			opword |= (uint16(count%8) << 9)
		} else if src.Mode == cpu.ModeData {
			opword |= 0x0020 // bit 5 = register count source
			opword |= (src.Register << 9)
		} else {
			return nil, fmt.Errorf("source of %s must be data register or immediate", mn.Value)
		}

		return []uint16{opword}, nil

	default:
		return nil, fmt.Errorf("%s requires 1 or 2 operands", mn.Value)
	}
}

//
// Bit Manipulation
//

// assembleBitManipulation handles BTST, BCHG, BCLR, BSET.
func (asm *Assembler) assembleBitManipulation(mn Mnemonic, operands []Operand) ([]uint16, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("%s requires 2 operands", mn.Value)
	}

	src, dst := operands[0], operands[1]
	mnLower := strings.ToLower(mn.Value)

	// Determine effective size for EA encoding and bit number modulo.
	// Data registers operate on longs (32 bits), memory on bytes (8 bits).
	eaSize := cpu.SizeByte
	bitModulo := uint64(8)
	if dst.Mode == cpu.ModeData {
		eaSize = cpu.SizeLong
		bitModulo = 32
	}

	// Immediate form: <op> #imm, <ea>
	if src.IsImmediate() {
		val, err := asm.parseConstant(src.Raw)
		if err != nil {
			return nil, err
		}

		opword := uint16(0x0800) // Base for immediate bit ops
		switch mnLower {
		case "btst":
			// 0x0800 is correct
		case "bchg":
			opword = 0x0840
		case "bclr":
			opword = 0x0880
		case "bset":
			opword = 0x08C0
		}

		eaBits, eaExt, err := asm.encodeEA(dst, eaSize)
		if err != nil {
			return nil, err
		}
		opword |= eaBits

		// Immediate bit number is an extension word.
		bitNum := uint16(uint64(val) % bitModulo)
		// The bit number is encoded in a single word, with the significant bits in the low byte.
		ext := []uint16{bitNum & 0x00FF}

		return append(append([]uint16{opword}, ext...), eaExt...), nil
	}

	// Register form: <op> Dn, <ea>
	if src.Mode != cpu.ModeData {
		return nil, fmt.Errorf("source of %s must be data register or immediate", mn.Value)
	}

	opword := uint16(cpu.OPBitManipulationBase)
	opword |= (src.Register << 9)

	eaBits, eaExt, err := asm.encodeEA(dst, eaSize)
	if err != nil {
		return nil, err
	}
	opword |= eaBits

	switch mnLower {
	case "btst":
		opword |= 0x0000
	case "bchg":
		opword |= 0x0040
	case "bclr":
		opword |= 0x0080
	case "bset":
		opword |= 0x00C0
	default:
		return nil, fmt.Errorf("invalid bit operation: %s", mn.Value)
	}

	return append([]uint16{opword}, eaExt...), nil
}
