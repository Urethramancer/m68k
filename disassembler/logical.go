package disassembler

import (
	"fmt"

	"github.com/Urethramancer/m68k/cpu"
)

// Immediate logical/arithmetic (ORI, ANDI, ADDI, SUBI, EORI, CMPI)
func decodeImmediateLogical(op uint16, pc int, code []byte) (string, string, int) {
	sizeBits := (op >> 6) & 0x3
	var mn string

	switch op & 0xFF00 {
	case 0x0000:
		mn = "ori"
	case 0x0200:
		mn = "andi"
	case 0x0400:
		mn = "subi"
	case 0x0600:
		mn = "addi"
	case 0x0A00:
		mn = "eori"
	case 0x0C00:
		mn = "cmpi"
	default:
		return "dc.w", fmt.Sprintf("0x%04x", op), 0
	}

	immText, immUsed := readImmediateBySize(code, pc, sizeBits)
	ea := op & 0x3F
	eaText, eaUsed := DecodeEA(ea, pc+immUsed, code, sizeBits)

	return mn + SizeSuffix(sizeBits), fmt.Sprintf("%s,%s", immText, eaText), immUsed + eaUsed
}

// decodeLogical handles AND, OR, and EOR instructions (non-immediate forms).
func decodeLogical(op uint16, pc int, code []byte) (string, string, int) {
	var mn string
	switch op & 0xF000 {
	case cpu.OPAND:
		mn = "and"
	case cpu.OPOR:
		mn = "or"
	case 0xB000: // EOR is in this range
		mn = "eor"
	default:
		// This case should ideally not be reached if called correctly.
		return "dc.w", fmt.Sprintf("0x%04x", op), 0
	}

	size := (op >> 6) & 3
	sizeStr := SizeSuffix(size)
	dir := (op & 0x0100) != 0 // 0 = EA -> Dn, 1 = Dn -> EA
	reg := (op >> 9) & 7
	ea := op & 0x3F
	eaText, used := DecodeEA(ea, pc, code, size)

	// Direction 1 is Dn -> EA.
	if dir {
		// Dn -> EA
		return mn + sizeStr, fmt.Sprintf("d%d,%s", reg, eaText), used
	}
	// EA -> Dn
	return mn + sizeStr, fmt.Sprintf("%s,d%d", eaText, reg), used
}

// decodeExg decodes the EXG (Exchange Registers) instruction.
func decodeExg(op uint16) (string, string, int) {
	regX := (op >> 9) & 7
	regY := op & 7
	opmode := (op >> 3) & 0x1F

	var ops string
	switch opmode {
	case 0b01000: // EXG Dx, Dy
		ops = fmt.Sprintf("d%d,d%d", regX, regY)
	case 0b01001: // EXG Ax, Ay
		ops = fmt.Sprintf("a%d,a%d", regX, regY)
	case 0b10001: // EXG Dx, Ay
		ops = fmt.Sprintf("d%d,a%d", regX, regY)
	default:
		// This path should not be reachable if dispatched correctly.
		return "dc.w", fmt.Sprintf("0x%04x", op), 0
	}

	return "exg", ops, 0
}

// decodeImmediateToSystemRegister decodes ANDI, ORI, and EORI to CCR/SR.
func decodeImmediateToSystemRegister(op uint16, pc int, code []byte) (string, string, int) {
	var mn, reg string
	var size uint16

	switch op {
	case cpu.OPANDItoCCR:
		mn = "andi"
		reg = "ccr"
		size = 0 // byte
	case cpu.OPORItoCCR:
		mn = "ori"
		reg = "ccr"
		size = 0 // byte
	case cpu.OPEORItoCCR:
		mn = "eori"
		reg = "ccr"
		size = 0 // byte
	case cpu.OPANDItoSR:
		mn = "andi"
		reg = "sr"
		size = 1 // word
	case cpu.OPORItoSR:
		mn = "ori"
		reg = "sr"
		size = 1 // word
	case cpu.OPEORItoSR:
		mn = "eori"
		reg = "sr"
		size = 1 // word
	default:
		return "dc.w", fmt.Sprintf("$%04x", op), 0
	}

	immText, used := readImmediateBySize(code, pc, size)
	return mn, fmt.Sprintf("%s,%s", immText, reg), used
}
