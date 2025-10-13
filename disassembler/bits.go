package disassembler

import "fmt"

// decodeBitManipulation decodes the BTST, BCHG, BCLR, and BSET instructions.
// These instructions have two forms: dynamic (bit number in a register)
// and immediate (bit number as an immediate value).
func decodeBitManipulation(op uint16, pc int, code []byte) (string, string, int) {
	var mn string
	var ea uint16
	var used int
	var ops string

	opType := (op >> 6) & 3
	mnBase := []string{"btst", "bchg", "bclr", "bset"}[opType]

	// Static (immediate) form: 0000 1000 ...
	if (op & 0xFF00) == 0x0800 {
		immText, immUsed := readImmediateBySize(code, pc, 0) // Bit number is always a byte
		used += immUsed
		ea = op & 0x3F

		var size uint16
		// Destination size is .l for data register, .b for memory.
		if (ea >> 3) == 0 { // Dn
			size = 2 // long
		} else {
			size = 0 // byte
		}
		mn = mnBase

		eaText, eaUsed := DecodeEA(ea, pc+used, code, size)
		used += eaUsed
		ops = fmt.Sprintf("%s,%s", immText, eaText)
		return mn, ops, used
	}

	// Dynamic (register) form: 0000 RRR 1 ...
	reg := (op >> 9) & 7
	ea = op & 0x3F

	var size uint16
	var sizeStr string
	// Destination size is .l for data register, .b for memory.
	if (ea >> 3) == 0 { // Dn
		size = 2 // long
		sizeStr = ".l"
	} else {
		size = 0 // byte
		sizeStr = ".b"
	}
	mn = mnBase + sizeStr

	eaText, eaUsed := DecodeEA(ea, pc, code, size)
	used += eaUsed
	ops = fmt.Sprintf("d%d,%s", reg, eaText)
	return mn, ops, used
}
