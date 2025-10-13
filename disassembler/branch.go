package disassembler

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

// isBranchMnemonic checks if an instruction is a form of branch.
func isBranchMnemonic(val string) bool {
	switch val {
	case "bra", "bsr", "bhi", "bls", "bcc", "bcs", "bne", "beq", "bvc", "bvs", "bpl", "bmi", "bge", "blt", "bgt", "ble":
		return true
	default:
		return strings.HasPrefix(val, "db")
	}
}

// decodeBranch decodes all branch and conditional branch opcodes.
func decodeBranch(op uint16, code []byte, pc int) (string, string, int) {
	cond := uint16((op >> 8) & 0xF)
	var name string
	switch cond {
	case 0x0:
		name = "bra"
	case 0x1:
		name = "bsr"
	default:
		name = "b" + condName(cond)
	}

	disp8 := uint8(op & 0xFF)
	// short (8-bit) displacement
	if disp8 != 0x00 && disp8 != 0xFF {
		val := int8(disp8)
		return name, formatDisp(int64(val)), 0
	}

	// word displacement
	if disp8 == 0x00 {
		if pc+2 > len(code) {
			return name, "?", 0
		}
		w := int16(binary.BigEndian.Uint16(code[pc:]))
		return name, formatDisp(int64(w)), 2
	}

	// long displacement (0xFF)
	if pc+4 > len(code) {
		return name, "?", 0
	}
	l := int32(binary.BigEndian.Uint32(code[pc:]))
	return name, formatDisp(int64(l)), 4
}

func condName(cond uint16) string {
	names := []string{"t", "f", "hi", "ls", "cc", "cs", "ne", "eq",
		"vc", "vs", "pl", "mi", "ge", "lt", "gt", "le"}
	if int(cond) < len(names) {
		return names[cond]
	}
	return "??"
}

// parseBranchOffset is more robust than naive fmt.Sscanf.
func parseBranchOffset(tok string) int32 {
	tok = strings.TrimSpace(tok)
	if tok == "" {
		return 0
	}
	if strings.HasPrefix(tok, "loc_") {
		return 0
	}
	if tok[0] == '+' {
		tok = tok[1:]
	}
	i, err := strconv.ParseInt(tok, 0, 64)
	if err != nil {
		return 0
	}
	return int32(i)
}

// decodeJmpJsr decodes the JMP and JSR instructions.
func decodeJmpJsr(op uint16, pc int, code []byte) (string, string, int) {
	var mn string
	if (op & 0x0040) == 0 {
		mn = "jsr"
	} else {
		mn = "jmp"
	}

	ea := op & 0x3F
	var size uint16 = 1 // Default to word
	if (ea >> 3) == 7 { // Mode 7
		if (ea & 7) == 1 { // Long
			size = 2
		}
	}

	eaText, used := DecodeEA(ea, pc, code, size)
	return mn, eaText, used
}

// decodeScc decodes the Scc (Set on condition) instruction.
func decodeScc(op uint16, pc int, code []byte) (string, string, int) {
	cond := (op >> 8) & 0xF
	mn := "s" + condName(cond)

	ea := op & 0x3F
	eaText, used := DecodeEA(ea, pc, code, 0)
	return mn, eaText, used
}

// decodeDbcc decodes the DBcc (Decrement and branch on condition) instruction.
func decodeDbcc(op uint16, pc int, code []byte) (string, string, int) {
	cond := (op >> 8) & 0xF
	reg := op & 7
	mn := "db" + condName(cond)

	if pc+1 >= len(code) {
		return mn, fmt.Sprintf("d%d,?", reg), 0
	}
	disp := int16(binary.BigEndian.Uint16(code[pc:]))

	return mn, fmt.Sprintf("d%d,%s", reg, formatDisp(int64(disp))), 2
}
