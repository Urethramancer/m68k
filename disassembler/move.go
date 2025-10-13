package disassembler

import (
	"encoding/binary"
	"fmt"

	"github.com/Urethramancer/m68k/cpu"
)

// decodeMoveGeneral decodes MOVE and MOVEA instructions.
func decodeMoveGeneral(op uint16, pc int, code []byte) (string, string, int) {
	sizeBits := (op >> 12) & 0x3
	var mn string
	switch sizeBits {
	case 1:
		mn = "move.b"
	case 2:
		mn = "move.l"
	case 3:
		mn = "move.w"
	default:
		return "dc.w", fmt.Sprintf("0x%04x", op), 0
	}

	srcEA := uint16(((op>>3)&0x7)<<3) | uint16(op&0x7)
	dstEA := uint16(((op>>6)&0x7)<<3) | uint16((op>>9)&0x7)
	srcText, cons1 := DecodeEA(srcEA, pc, code, sizeBits)
	dstText, cons2 := DecodeEA(dstEA, pc+cons1, code, sizeBits)
	dstMode := (op >> 6) & 0x7

	if dstMode == 1 {
		if sizeBits == 2 {
			mn = "movea.l"
		} else {
			mn = "movea.w"
		}
	}
	return mn, fmt.Sprintf("%s,%s", srcText, dstText), cons1 + cons2
}

// decodeMovem decodes MOVEM (Move Multiple Registers) instructions.
func decodeMovem(op uint16, pc int, code []byte) (string, string, int) {
	isLoad := (op & 0x0400) != 0
	size := ".w"
	if (op & 0x0040) != 0 {
		size = ".l"
	}

	ea := op & 0x3F
	if pc+2 > len(code) {
		return "movem" + size, "?", 0
	}
	mask := binary.BigEndian.Uint16(code[pc:])

	eaText, used := DecodeEA(ea, pc+2, code, 0)
	regList := movemMaskToList(mask)

	if isLoad {
		// Memory → Registers: movem <ea>,<reglist>
		return "movem" + size, fmt.Sprintf("%s,%s", eaText, regList), used + 2
	}
	// Registers → Memory: movem <reglist>,<ea>
	return "movem" + size, fmt.Sprintf("%s,%s", regList, eaText), used + 2
}

// decodeMovep decodes the MOVEP (Move Peripheral) instruction.
func decodeMovep(op uint16, pc int, code []byte) (string, string, int) {
	dataReg, addrReg := (op>>9)&7, op&7
	opmode := (op >> 6) & 7

	var sizeStr string
	var isMemToReg bool
	switch opmode {
	case 4:
		sizeStr, isMemToReg = ".w", true
	case 5:
		sizeStr, isMemToReg = ".l", true
	case 6:
		sizeStr, isMemToReg = ".w", false
	case 7:
		sizeStr, isMemToReg = ".l", false
	default:
		return "dc.w", fmt.Sprintf("0x%04x", op), 0
	}

	if pc+2 > len(code) {
		return "movep" + sizeStr, "?", 0
	}
	disp := int16(binary.BigEndian.Uint16(code[pc:]))

	var ops string
	if isMemToReg {
		ops = fmt.Sprintf("(%d,a%d),d%d", disp, addrReg, dataReg)
	} else {
		ops = fmt.Sprintf("d%d,(%d,a%d)", dataReg, disp, addrReg)
	}
	return "movep" + sizeStr, ops, 2
}

// decodeMoveSystemRegister handles MOVE to/from SR, CCR, and USP.
func decodeMoveSystemRegister(op uint16, pc int, code []byte) (string, string, int) {
	if (op & 0xFFF0) == cpu.OPMOVEToUSP {
		reg := op & 7
		if (op & 0x0008) != 0 {
			return "move.l", fmt.Sprintf("usp,a%d", reg), 0
		}
		return "move.l", fmt.Sprintf("a%d,usp", reg), 0
	}

	ea := op & 0x3F
	eaText, used := DecodeEA(ea, pc, code, 1)
	switch op & 0xFFC0 {
	case cpu.OPMOVEFromSR:
		return "move", fmt.Sprintf("sr,%s", eaText), used
	case cpu.OPMOVEToCCR:
		return "move", fmt.Sprintf("%s,ccr", eaText), used
	case cpu.OPMOVEToSR:
		return "move", fmt.Sprintf("%s,sr", eaText), used
	}
	return "dc.w", fmt.Sprintf("$%04x", op), 0
}
