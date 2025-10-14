package cpu

import (
	"fmt"
)

// DecodedInstruction holds the parsed details of an M68k instruction.
type DecodedInstruction struct {
	Handler func(*CPU, *DecodedInstruction) error
	Size    Size
	SrcMode uint16
	SrcReg  uint16
	DstMode uint16
	DstReg  uint16
	OpMode  uint16
}

// Decode takes a 16-bit opcode and returns a structured DecodedInstruction.
// This is the heart of the CPU, determining what each instruction means.
func (c *CPU) Decode(opcode uint16) (*DecodedInstruction, error) {
	inst := &DecodedInstruction{}

	// Switch on the top 4 bits of the opcode for efficient decoding.
	switch opcode >> 12 {

	// MOVE instructions are patterns 0001 (byte), 0011 (word), 0010 (long)
	case 1, 2, 3:
		switch (opcode >> 12) & 3 {
		case 1:
			inst.Size = SizeByte
		case 3:
			inst.Size = SizeWord
		case 2:
			inst.Size = SizeLong
		}

		inst.DstMode = (opcode >> 6) & 7
		inst.DstReg = (opcode >> 9) & 7
		inst.SrcMode = (opcode >> 3) & 7
		inst.SrcReg = opcode & 7

		// MOVEA is a special case of MOVE where the destination is an address register.
		if inst.DstMode == ModeAddr {
			inst.Handler = (*CPU).opMOVEA
		} else {
			inst.Handler = (*CPU).opMOVE
		}
		return inst, nil

	// ADDQ/SUBQ instruction pattern is 0101
	case 5:
		// Scc and DBcc have the pattern 0101 <cond> 11 <ea>, which we exclude.
		if opcode&0x00C0 == 0x00C0 {
			break // Fall through to unknown
		}

		data := (opcode >> 9) & 7
		if data == 0 {
			data = 8 // A value of 0 in the data field means 8
		}
		inst.SrcReg = uint16(data) // Pass immediate value via SrcReg

		switch (opcode >> 6) & 3 {
		case 0:
			inst.Size = SizeByte
		case 1:
			inst.Size = SizeWord
		case 2:
			inst.Size = SizeLong
		}

		inst.DstMode = (opcode >> 3) & 7
		inst.DstReg = opcode & 7

		if (opcode>>8)&1 == 0 {
			inst.Handler = (*CPU).opADDQ
		} else {
			// inst.Handler = (*CPU).opSUBQ // To be implemented
		}
		return inst, nil

	// MOVEQ instruction pattern is 0111
	case 7:
		inst.Handler = (*CPU).opMOVEQ
		inst.Size = SizeLong // MOVEQ is always long
		inst.DstReg = (opcode >> 9) & 7
		// The immediate value is stored in the low 8 bits. We pass it via SrcReg.
		inst.SrcReg = opcode & 0xFF
		return inst, nil

	// ADD/ADDA instruction pattern is 1101
	case 13: // 0xD
		inst.Handler = (*CPU).opADD
		inst.OpMode = opcode & 0x01C0 // Contains direction and size bits
		inst.DstReg = (opcode >> 9) & 7
		inst.SrcMode = (opcode >> 3) & 7
		inst.SrcReg = opcode & 7

		switch (opcode >> 6) & 3 {
		case 0:
			inst.Size = SizeByte
		case 1:
			inst.Size = SizeWord
		case 2:
			inst.Size = SizeLong
		}
		return inst, nil

	} // end switch

	return nil, fmt.Errorf("unknown or unimplemented instruction: %04X", opcode)
}
