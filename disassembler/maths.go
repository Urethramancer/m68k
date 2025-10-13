package disassembler

import (
	"fmt"

	"github.com/Urethramancer/m68k/cpu"
)

func decodeAdd(op uint16, pc int, code []byte) (string, string, int) {
	size := (op >> 6) & 3
	sizeStr := SizeSuffix(size)
	dir := (op & 0x0100) != 0 // 0 = EA -> Dn, 1 = Dn -> EA
	reg := (op >> 9) & 7
	ea := op & 0x3F
	eaText, used := DecodeEA(ea, pc, code, size)

	// ADDA detection: bits 8..6 == 0b111 (0x01C0)
	if (op & 0x01C0) == 0x01C0 {
		// ADDA always: EA -> An
		return "adda" + sizeStr, fmt.Sprintf("%s,a%d", eaText, reg), used
	}

	// Standard ADD: direction matters.
	if dir {
		// Dn -> EA
		return "add" + sizeStr, fmt.Sprintf("d%d,%s", reg, eaText), used
	}
	// EA -> Dn
	return "add" + sizeStr, fmt.Sprintf("%s,d%d", eaText, reg), used
}

func decodeSub(op uint16, pc int, code []byte) (string, string, int) {
	size := (op >> 6) & 3
	sizeStr := SizeSuffix(size)
	dir := (op & 0x0100) != 0 // 0 = EA -> Dn, 1 = Dn -> EA
	reg := (op >> 9) & 7
	ea := op & 0x3F
	eaText, used := DecodeEA(ea, pc, code, size)

	// SUBA detection: bits 8..6 == 0b111 (0x01C0)
	if (op & 0x01C0) == 0x01C0 {
		// SUBA always: EA -> An
		return "suba" + sizeStr, fmt.Sprintf("%s,a%d", eaText, reg), used
	}

	// Standard SUB: direction matters.
	if dir {
		// Dn -> EA
		return "sub" + sizeStr, fmt.Sprintf("d%d,%s", reg, eaText), used
	}
	// EA -> Dn
	return "sub" + sizeStr, fmt.Sprintf("%s,d%d", eaText, reg), used
}

func decodeAddxSubx(op uint16, pc int, code []byte) (string, string, int) {
	var base string
	switch op & 0xF100 {
	case cpu.OPADDX:
		base = "addx"
	case cpu.OPSUBX:
		base = "subx"
	default:
		return "dc.w", fmt.Sprintf("0x%04x", op), 0
	}

	sizeBits := (op >> 6) & 0x3
	mn := base + SizeSuffix(sizeBits)

	src := op & 7
	dst := (op >> 9) & 7
	mode := (op >> 3) & 7

	switch mode {
	case 0: // register form
		return mn, fmt.Sprintf("d%d,d%d", src, dst), 0
	case 4: // predecrement form
		return mn, fmt.Sprintf("-(a%d),-(a%d)", src, dst), 0
	}

	ea := uint16((mode << 3) | src)
	eaText, used := DecodeEA(ea, pc, code, sizeBits)
	return mn, fmt.Sprintf("%s,d%d", eaText, dst), used
}

// decodeMulDiv decodes MULS, MULU, DIVS, DIVU.
func decodeMulDiv(op uint16, pc int, code []byte) (string, string, int) {
	var mn string
	opType := (op >> 11) & 0x1F
	switch opType {
	case 0x18: // 11000 = MULU
		mn = "mulu.w"
	case 0x19: // 11001 = MULS
		mn = "muls.w"
	case 0x10: // 10000 = DIVU
		mn = "divu.w"
	case 0x11: // 10001 = DIVS
		mn = "divs.w"
	default:
		return "dc.w", fmt.Sprintf("0x%04x", op), 0
	}

	reg := (op >> 9) & 7
	ea := op & 0x3F
	// These instructions are always word-sized for the source operand.
	eaText, used := DecodeEA(ea, pc, code, 1) // size=1 for word

	return mn, fmt.Sprintf("%s,d%d", eaText, reg), used
}
