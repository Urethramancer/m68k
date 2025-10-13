package cpu

import "fmt"

// DecodedInstruction holds the parsed details of an M68k instruction.
// This struct is passed from the decoder to the executor.
type DecodedInstruction struct {
	Opcode   uint16
	Size     Size
	SrcMode  uint16
	SrcReg   uint16
	DestMode uint16
	DestReg  uint16
	Handler  func(*CPU, *DecodedInstruction) error
	// TODO: Add other fields as needed for different instructions,
	// such as immediate data or displacement values.
}

// Decode takes a 16-bit opcode word from the program counter and parses it
// into a structured format. It identifies the instruction to be run, its
// size, addressing modes, and registers.
func (c *CPU) Decode(opcode uint16) (*DecodedInstruction, error) {
	inst := &DecodedInstruction{Opcode: opcode}

	// MOVE instruction (opcode 00ssdddmmmmrrr)
	if opcode&0xC000 == OPMOVE {
		// The MOVE instruction's layout is: 00_ss_ddd_mmm_rrr
		// ss   = size
		// ddd  = destination register
		// mmm  = destination mode
		// rrr  = source register
		// Note: The size bits for MOVE are 01=byte, 11=word, 10=long. A mapping is needed.
		inst.Size = Size((opcode >> 12) & 0x3)
		inst.DestReg = (opcode >> 9) & 0x7
		inst.DestMode = (opcode >> 6) & 0x7
		inst.SrcMode = (opcode >> 3) & 0x7
		inst.SrcReg = opcode & 0x7
		// inst.Handler = (*CPU).opMOVE // Assign the handler function
		return inst, nil
	}

	// MOVEQ instruction (opcode 0111ddddiiiiiiii)
	if opcode&0xF000 == OPMOVEQ {
		inst.Size = SizeLong // MOVEQ is always long word
		inst.DestReg = (opcode >> 9) & 0x7
		// The source is immediate data stored in the lower 8 bits.
		inst.SrcMode = RegImmediate
		// The source "register" field contains the immediate data itself.
		inst.SrcReg = opcode & 0xFF
		// inst.Handler = (*CPU).opMOVEQ // Assign the handler function
		return inst, nil
	}

	// TODO: Implement decoding for other instructions.

	return nil, fmt.Errorf("unknown or unimplemented opcode: %04X", opcode)
}
