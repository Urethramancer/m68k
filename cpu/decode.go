package cpu

import "fmt"

// DecodedInstruction holds the parsed details of a single machine code instruction.
// It is the intermediate representation passed from the decoder to the executor.
type DecodedInstruction struct {
	// Handler is the function that will execute this instruction.
	Handler func(*CPU, *DecodedInstruction) error
	// Size is the operation size (.b, .w, .l).
	Size Size
	// SrcMode and SrcReg define the source effective address (EA).
	SrcMode, SrcReg uint16
	// DstMode and DstReg define the destination effective address (EA).
	DstMode, DstReg uint16
	// OpMode is used by some instructions (like ADD/SUB) for direction and size bits.
	OpMode uint16
}

// Decode parses a 16-bit opcode and returns a structured instruction.
func (c *CPU) Decode(opcode uint16) (*DecodedInstruction, error) {
	inst := &DecodedInstruction{}

	// Switch on the top 4 bits of the opcode, which is a common way
	// to group M68k instructions.
	switch opcode >> 12 {
	case 0b0001, 0b0010, 0b0011: // MOVE
		return c.decodeMove(opcode, inst)
	case 0b0101: // ADDQ, SUBQ
		return c.decodeAddqSubq(opcode, inst)
	case 0b0111: // MOVEQ
		return c.decodeMoveq(opcode, inst)
	case 0b1101: // ADD, ADDX
		return c.decodeAdd(opcode, inst)
	case 0b0100: // Miscellaneous group
		switch {
		case opcode&0xFFC0 == OPTRAP: // TRAP
			inst.Handler = (*CPU).opTRAP
			inst.DstReg = opcode & 0xF // The vector number is in the lower 4 bits.
			return inst, nil
		case opcode == OPRTS: // RTS
			inst.Handler = (*CPU).opRTS
			return inst, nil
		}
	}

	return nil, fmt.Errorf("unknown or unimplemented instruction: %04X", opcode)
}

// decodeMove handles the general MOVE and MOVEA instructions.
func (c *CPU) decodeMove(opcode uint16, inst *DecodedInstruction) (*DecodedInstruction, error) {
	sizeBits := (opcode >> 12) & 0b11
	switch sizeBits {
	case 0b01:
		inst.Size = SizeByte
	case 0b11:
		inst.Size = SizeWord
	case 0b10:
		inst.Size = SizeLong
	default:
		return nil, fmt.Errorf("invalid size bits in MOVE opcode %04X", opcode)
	}

	inst.DstMode = (opcode >> 6) & 0x7
	inst.DstReg = (opcode >> 9) & 0x7
	inst.SrcMode = (opcode >> 3) & 0x7
	inst.SrcReg = opcode & 0x7

	// A MOVE instruction with an address register as the destination is MOVEA.
	if inst.DstMode == ModeAddr {
		inst.Handler = (*CPU).opMOVEA
	} else {
		inst.Handler = (*CPU).opMOVE
	}
	return inst, nil
}

// decodeMoveq handles the MOVEQ instruction.
func (c *CPU) decodeMoveq(opcode uint16, inst *DecodedInstruction) (*DecodedInstruction, error) {
	inst.Handler = (*CPU).opMOVEQ
	inst.Size = SizeLong // MOVEQ is always a long operation.
	inst.DstReg = (opcode >> 9) & 0x7
	// The immediate 8-bit value is stored in the lower byte.
	// We'll pass it to the handler via the SrcReg field for convenience.
	inst.SrcReg = opcode & 0xFF
	return inst, nil
}

// decodeAdd handles the ADD and ADDX instructions.
func (c *CPU) decodeAdd(opcode uint16, inst *DecodedInstruction) (*DecodedInstruction, error) {
	inst.Handler = (*CPU).opADD
	inst.OpMode = (opcode >> 6) & 0b111 // Captures direction and size bits
	sizeBits := (opcode >> 6) & 0b11
	switch sizeBits {
	case 0b00:
		inst.Size = SizeByte
	case 0b01:
		inst.Size = SizeWord
	case 0b10:
		inst.Size = SizeLong
	default:
		return nil, fmt.Errorf("invalid size bits in ADD opcode %04X", sizeBits)
	}
	inst.DstReg = (opcode >> 9) & 0x7 // This is the Dn register for the operation
	inst.SrcMode = (opcode >> 3) & 0x7
	inst.SrcReg = opcode & 0x7
	return inst, nil
}

// decodeAddqSubq handles the ADDQ and SUBQ instructions.
func (c *CPU) decodeAddqSubq(opcode uint16, inst *DecodedInstruction) (*DecodedInstruction, error) {
	// Bit 8 determines ADDQ (0) or SUBQ (1)
	if (opcode>>8)&1 == 0 {
		inst.Handler = (*CPU).opADDQ
	} else {
		// inst.Handler = (*CPU).opSUBQ // Not yet implemented
		return nil, fmt.Errorf("unimplemented instruction SUBQ: %04X", opcode)
	}

	// The immediate data (1-8) is in bits 11-9. A value of 0 represents 8.
	data := (opcode >> 9) & 0x7
	if data == 0 {
		data = 8
	}
	inst.SrcReg = data // Pass data to handler via SrcReg

	sizeBits := (opcode >> 6) & 0b11
	switch sizeBits {
	case 0b00:
		inst.Size = SizeByte
	case 0b01:
		inst.Size = SizeWord
	case 0b10:
		inst.Size = SizeLong
	default:
		return nil, fmt.Errorf("invalid size bits for ADDQ/SUBQ: %04X", opcode)
	}
	inst.DstMode = (opcode >> 3) & 0x7
	inst.DstReg = opcode & 0x7
	return inst, nil
}
