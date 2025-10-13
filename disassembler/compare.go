package disassembler

import "fmt"

// CMP / EOR
func decodeCmp(op uint16, pc int, code []byte) (string, string, int) {
	opmode := (op >> 6) & 7
	reg := (op >> 9) & 7
	var size uint16
	var sizeStr string
	var mn string

	// Bit 8 (0x0100) distinguishes CMP/CMPA from EOR.
	if (op & 0x0100) != 0 {
		// This is an EOR instruction. Size is encoded in the 3-bit opmode field (bits 8-6).
		mn = "eor"
		switch opmode {
		case 4: // 100
			size = 0 // byte
			sizeStr = ".b"
		case 5: // 101
			size = 1 // word
			sizeStr = ".w"
		case 6: // 110
			size = 2 // long
			sizeStr = ".l"
		default:
			// Invalid opmode for EOR Dn,<ea>
			return "dc.w", fmt.Sprintf("0x%04x", op), 0
		}
		ea := op & 0x3F
		eaText, used := DecodeEA(ea, pc, code, size)
		// EOR is always Dn -> <ea>
		return mn + sizeStr, fmt.Sprintf("d%d,%s", reg, eaText), used
	}

	// This is a CMP or CMPA instruction. Size is encoded differently.
	mn = "cmp"
	sizeField := (op >> 6) & 3
	switch opmode {
	case 3: // 011 - CMPA.W
		mn = "cmpa"
		size = 1
		sizeStr = ".w"
	case 7: // 111 - CMPA.L
		mn = "cmpa"
		size = 2
		sizeStr = ".l"
	default: // Standard CMP uses the 2-bit size field
		size = sizeField
		sizeStr = SizeSuffix(sizeField)
	}

	ea := op & 0x3F
	eaText, used := DecodeEA(ea, pc, code, size)

	if mn == "cmpa" {
		return mn + sizeStr, fmt.Sprintf("%s,a%d", eaText, reg), used
	}
	return mn + sizeStr, fmt.Sprintf("%s,d%d", eaText, reg), used
}

// decodeChk decodes the CHK instruction.
func decodeChk(op uint16, pc int, code []byte) (string, string, int) {
	reg := (op >> 9) & 7
	ea := op & 0x3F
	// Size is always .w for CHK on MC68000
	eaText, used := DecodeEA(ea, pc, code, 1)
	return "chk.w", fmt.Sprintf("%s,d%d", eaText, reg), used
}

// decodeCmpm decodes the CMPM (Compare Memory) instruction.
// Format: CMPM (Ay)+,(Ax)+
func decodeCmpm(op uint16) (string, string, int) {
	sizeField := (op >> 6) & 3
	sizeStr := SizeSuffix(sizeField)

	regX := (op >> 9) & 7 // Ax
	regY := op & 7        // Ay

	return "cmpm" + sizeStr, fmt.Sprintf("(a%d)+,(a%d)+", regY, regX), 0
}

// decodeTas decodes the TAS (Test and Set) instruction.
// The opcode format is 0100 1010 11 <ea>.
// TAS is always byte-sized and the <ea> cannot be an address register direct,
// PC-relative, or immediate mode.
func decodeTas(op uint16, pc int, code []byte) (string, string, int) {
	// The size is implicitly byte.
	ea := op & 0x3F
	eaText, used := DecodeEA(ea, pc, code, 0)
	return "tas", eaText, used
}
