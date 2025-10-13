package disassembler

import "fmt"

// decodeShiftRotateGeneric decodes LSL/LSR/ASL/ASR/ROL/ROR/ROXL/ROXR instructions.
//
// 68000 shift/rotate instructions use bits:
//
//	15–12: 1110 (0xE)
//	11–9 : <register/count>
//	8    : 0 = register shift, 1 = memory shift (not handled here)
//	7    : 0 = register count, 1 = immediate count (in register form)
//	6–7  : size bits (00=byte, 01=word, 10=long)
//	5–3  : direction + type of shift
//	2–0  : destination register
//
// The instruction families are divided into right and left variants:
//
//	Right: ASR, LSR, ROXR, ROR
//	Left : ASL, LSL, ROXL, ROL
//
// Example encodings:
//
//	0xE008 → LSR.W D0
//	0xE018 → ROR.W D0
//	0xE108 → LSL.W D0
//	0xE118 → ROL.W D0
func decodeShiftRotateGeneric(op uint16) (string, string, int) {
	// Bit 8 (0x0100): 0 = right shift/rotate, 1 = left shift/rotate
	isLeft := (op & 0x0100) != 0

	// Operation index bits 5–3
	opType := (op >> 3) & 3 // 0..3 for AS/LS/ROX/ROR base
	if isLeft {
		opType += 4 // add 4 to select ASL/LSL/ROXL/ROL
	}

	mnBases := []string{"asr", "lsr", "roxr", "ror", "asl", "lsl", "roxl", "rol"}
	mn := mnBases[opType]

	// Bits 7–6 encode size: 00=b, 01=w, 10=l
	switch (op >> 6) & 3 {
	case 0:
		mn += ".b"
	case 1:
		mn += ".w"
	case 2:
		mn += ".l"
	}

	// Bit 5 (0x0020) distinguishes register-count (0) vs immediate-count (1) forms
	isRegForm := (op & 0x0020) == 0
	if isRegForm {
		cntReg := (op >> 9) & 7
		dstReg := op & 7
		return mn, fmt.Sprintf("d%d,d%d", cntReg, dstReg), 0
	}

	// Immediate count (bits 11–9, 0 means 8)
	cnt := (op >> 9) & 7
	if cnt == 0 {
		cnt = 8
	}
	dstReg := op & 7
	return mn, fmt.Sprintf("#%d,d%d", cnt, dstReg), 0
}
