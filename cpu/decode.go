package cpu

import (
	"fmt"
)

// DecodedInstruction holds the parsed details of an M68k instruction.
type DecodedInstruction struct {
	// Handler is the function that executes this instruction.
	Handler func(*CPU, *DecodedInstruction) error
	// Size of the operation (Byte, Word, Long).
	Size Size
	// SrcMode is the source addressing mode.
	SrcMode uint16
	// SrcReg is the source register number.
	SrcReg uint16
	// DstMode is the destination addressing mode.
	DstMode uint16
	// DstReg is the destination register number.
	DstReg uint16
	// Other fields for specific instructions (e.g., immediate values, displacements).
	// We will add these as needed.
}

// Decode takes a 16-bit opcode and returns a structured DecodedInstruction.
// This is the heart of the CPU, determining what each instruction means.
func (c *CPU) Decode(opcode uint16) (*DecodedInstruction, error) {
	inst := &DecodedInstruction{}

	// The order of these checks is important. More specific opcodes must be
	// checked before more general ones to avoid misidentification.

	// Line 7xxx: MOVEQ (Move Quick)
	// Format: 0111 <reg> 0 <8-bit data>
	if opcode&0xF000 == OPMOVEQ {
		inst.Handler = (*CPU).opMOVEQ
		inst.Size = SizeLong // MOVEQ is always a long word operation
		inst.DstReg = (opcode >> 9) & 0x7
		// We store the immediate data in SrcReg for the handler's convenience.
		inst.SrcReg = opcode & 0xFF
		return inst, nil
	}

	// Lines 0, 1, 2, 3: MOVE, MOVEA
	// Format: 00xx <dst reg> <dst mode> <src mode> <src reg>
	// The top two bits (15, 14) being 00 signifies a MOVE instruction.
	if opcode&0xC000 == 0x0000 {
		// Extract size from bits 13 and 12
		switch (opcode >> 12) & 0x3 {
		case 1:
			inst.Size = SizeByte
		case 3:
			inst.Size = SizeWord
		case 2:
			inst.Size = SizeLong
		default:
			return nil, fmt.Errorf("invalid size for MOVE instruction")
		}

		// Extract addressing modes and registers
		inst.DstMode = (opcode >> 6) & 0x7
		inst.DstReg = (opcode >> 9) & 0x7
		inst.SrcMode = (opcode >> 3) & 0x7
		inst.SrcReg = opcode & 0x7

		// Distinguish between MOVE and MOVEA.
		// MOVEA is a MOVE where the destination is an address register (ModeAddr).
		// It has a different handler because it doesn't affect status flags.
		if inst.DstMode == ModeAddr {
			inst.Handler = (*CPU).opMOVEA
			// MOVEA can only be Word or Long
			if inst.Size == SizeByte {
				return nil, fmt.Errorf("MOVEA cannot have byte size")
			}
		} else {
			inst.Handler = (*CPU).opMOVE
		}
		return inst, nil
	}

	// --- Placeholder blocks for other instruction families ---

	// Line 4xxx: Miscellaneous instructions (NEG, NOT, TST, CLR, JMP, JSR, etc.)
	if opcode&0xF000 == 0x4000 {
		// Future logic for this group goes here.
	}

	// Line 5xxx: ADDQ, SUBQ, Scc, DBcc
	if opcode&0xF000 == 0x5000 {
		// Future logic for this group goes here.
	}

	// Line 6xxx: Branch instructions (BRA, BSR, Bcc)
	if opcode&0xF000 == 0x6000 {
		// Future logic for this group goes here.
	}

	// Line 8xxx: OR, DIV, SBCD
	if opcode&0xF000 == 0x8000 {
		// Future logic for this group goes here.
	}

	// Line 9xxx: SUB, SUBA, SUBX
	if opcode&0xF000 == 0x9000 {
		// Future logic for this group goes here.
	}

	// Line Bxxx: CMP, CMPA, EOR
	if opcode&0xF000 == 0xB000 {
		// Future logic for this group goes here.
	}

	// Line Cxxx: AND, MUL, ABCD, EXG
	if opcode&0xF000 == 0xC000 {
		// Future logic for this group goes here.
	}

	// Line Dxxx: ADD, ADDA, ADDX
	if opcode&0xF000 == 0xD000 {
		// Future logic for this group goes here.
	}

	// Line Exxx: Shift and Rotate instructions
	if opcode&0xF000 == 0xE000 {
		// Future logic for this group goes here.
	}

	// If no handler has been assigned, the instruction is unknown or unimplemented.
	if inst.Handler == nil {
		return nil, fmt.Errorf("unknown or unimplemented instruction: %04X", opcode)
	}

	return inst, nil
}
