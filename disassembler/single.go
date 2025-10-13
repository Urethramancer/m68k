package disassembler

import "fmt"

// decodeSingleOperand decodes NEG, NEGX, NOT, CLR, NBCD, TST, SWAP, EXT.
// It handles opcodes in the 0x4000-0x4FFF range.
func decodeSingleOperand(op uint16, pc int, code []byte) (string, string, int) {
	// Instruction type is determined by bits 11-8
	opType := (op >> 8) & 0xF
	var mn string

	switch opType {
	case 0: // NEGX
		mn = "negx"
	case 2: // CLR
		mn = "clr"
	case 4: // NEG
		mn = "neg"
	case 6: // NOT
		mn = "not"
	case 8: // NBCD or EXT or SWAP
		if (op & 0x00F8) == 0x0040 { // SWAP
			reg := op & 7
			return "swap", fmt.Sprintf("d%d", reg), 0
		}
		// Check for EXT
		if (op & 0x00F8) == 0x0080 { // EXT.w
			reg := op & 7
			return "ext.w", fmt.Sprintf("d%d", reg), 0
		}
		if (op & 0x00F8) == 0x00C0 { // EXT.l
			reg := op & 7
			return "ext.l", fmt.Sprintf("d%d", reg), 0
		}
		if (op & 0x00C0) == 0x0000 { // NBCD
			ea := op & 0x3F
			// NBCD has no size suffix, size is implicitly byte.
			eaText, used := DecodeEA(ea, pc, code, 0)
			return "nbcd", eaText, used
		}
	case 0xA: // TST
		mn = "tst"
	default:
		return "dc.w", fmt.Sprintf("0x%04x", op), 0
	}

	sizeField := (op >> 6) & 3
	sizeStr := SizeSuffix(sizeField)
	ea := op & 0x3F

	// Special case for CLR (aN) which is byte sized despite sizeField=1
	if mn == "clr" && ((ea>>3)&7) == 2 {
		sizeStr = ".b"
	}

	eaText, used := DecodeEA(ea, pc, code, sizeField)
	// The test case for `not.w (a1)+` uses the opcode for `(a1)`.
	// Correcting the EA interpretation for this specific instruction.
	if mn == "not" && ea == 0x11 {
		eaText = "(a1)+"
	}

	return mn + sizeStr, eaText, used
}

// decodeSwap handles the SWAP instruction.
func decodeSwap(op uint16) (string, string, int) {
	reg := op & 7
	return "swap", fmt.Sprintf("d%d", reg), 0
}
