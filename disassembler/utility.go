package disassembler

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// SizeSuffix returns the canonical size suffix (.b, .w, .l).
func SizeSuffix(bits uint16) string {
	switch bits {
	case 0:
		return ".b"
	case 1:
		return ".w"
	case 2:
		return ".l"
	default:
		return ""
	}
}

// movemMaskToList converts a register mask into a canonical, human-readable string list (e.g., "d0-d3/a0/a6").
func movemMaskToList(mask uint16) string {
	dRegs := make([]int, 0, 8)
	aRegs := make([]int, 0, 8)

	// The mask is always encoded in the same canonical order:
	// Bits 0-7 -> D0-D7
	// Bits 8-15 -> A0-A7
	for i := 0; i < 8; i++ {
		if (mask & (1 << i)) != 0 {
			dRegs = append(dRegs, i)
		}
		if (mask & (1 << (i + 8))) != 0 {
			aRegs = append(aRegs, i)
		}
	}

	var parts []string
	if len(dRegs) > 0 {
		parts = append(parts, formatRegRange("d", dRegs)...)
	}
	if len(aRegs) > 0 {
		parts = append(parts, formatRegRange("a", aRegs)...)
	}

	return strings.Join(parts, "/")
}

// formatRegRange is a helper to turn a list of register numbers into ranges.
func formatRegRange(prefix string, regs []int) []string {
	if len(regs) == 0 {
		return nil
	}
	var parts []string
	start, end := regs[0], regs[0]

	for i := 1; i < len(regs); i++ {
		if regs[i] == end+1 {
			end = regs[i]
		} else {
			if start == end {
				parts = append(parts, fmt.Sprintf("%s%d", prefix, start))
			} else {
				parts = append(parts, fmt.Sprintf("%s%d-%s%d", prefix, start, prefix, end))
			}
			start, end = regs[i], regs[i]
		}
	}
	if start == end {
		parts = append(parts, fmt.Sprintf("%s%d", prefix, start))
	} else {
		parts = append(parts, fmt.Sprintf("%s%d-%s%d", prefix, start, prefix, end))
	}
	return parts
}

// (The rest of the file is unchanged)

// DecodeEA decodes the effective address field.
func DecodeEA(ea uint16, pc int, code []byte, size uint16) (string, int) {
	mode := (ea >> 3) & 7
	reg := ea & 7

	switch mode {
	case 0:
		return fmt.Sprintf("d%d", reg), 0
	case 1:
		return fmt.Sprintf("a%d", reg), 0
	case 2:
		return fmt.Sprintf("(a%d)", reg), 0
	case 3:
		return fmt.Sprintf("(a%d)+", reg), 0
	case 4:
		return fmt.Sprintf("-(a%d)", reg), 0
	case 5:
		if pc+2 > len(code) {
			return fmt.Sprintf("(?,a%d)", reg), 0
		}
		disp := int16(binary.BigEndian.Uint16(code[pc:]))
		return fmt.Sprintf("(%s,a%d)", formatDisp16(disp), reg), 2
	case 6:
		if pc+2 > len(code) {
			return fmt.Sprintf("(?,a%d,x?)", reg), 0
		}
		ext := binary.BigEndian.Uint16(code[pc:])
		disp := int8(ext & 0xFF)
		idx := (ext >> 12) & 7
		sizeChar := "w"
		if (ext & 0x0800) != 0 {
			sizeChar = "l"
		}
		regType := "d"
		if (ext & 0x8000) != 0 {
			regType = "a"
		}
		return fmt.Sprintf("(%s,a%d,%s%d.%s)", formatDisp8(disp), reg, regType, idx, sizeChar), 2
	case 7:
		switch reg {
		case 0:
			if pc+2 > len(code) {
				return "(?.w)", 0
			}
			addr := binary.BigEndian.Uint16(code[pc:])
			return fmt.Sprintf("$%x.w", addr), 2
		case 1:
			if pc+4 > len(code) {
				return "(?.l)", 0
			}
			addr := binary.BigEndian.Uint32(code[pc:])
			return fmt.Sprintf("$%x.l", addr), 4
		case 2:
			if pc+2 > len(code) {
				return "(?,pc)", 0
			}
			disp := int16(binary.BigEndian.Uint16(code[pc:]))
			return fmt.Sprintf("(%s,pc)", formatDisp16(disp)), 2
		case 3:
			if pc+2 > len(code) {
				return "(?,pc,xn)", 0
			}
			ext := binary.BigEndian.Uint16(code[pc:])
			disp := int8(ext & 0xFF)
			idx := (ext >> 12) & 7
			sizeChar := "w"
			if (ext & 0x0800) != 0 {
				sizeChar = "l"
			}
			regType := "d"
			if (ext & 0x8000) != 0 {
				regType = "a"
			}
			return fmt.Sprintf("(%s,pc,%s%d.%s)", formatDisp8(disp), regType, idx, sizeChar), 2
		case 4:
			return readImmediateBySize(code, pc, size)
		}
	}
	return fmt.Sprintf("(ea mode=%d reg=%d)", mode, reg), 0
}

// readImmediateBySize reads immediate data based on the size field.
func readImmediateBySize(code []byte, pc int, size uint16) (string, int) {
	n := len(code)
	switch size {
	case 0:
		if pc+2 > n {
			return "#<trunc>", 0
		}
		val := int8(code[pc+1])
		return fmt.Sprintf("#%d", val), 2
	case 1:
		if pc+2 > n {
			return "#<trunc>", 0
		}
		w := int16(binary.BigEndian.Uint16(code[pc:]))
		if w >= 0 && w <= 255 {
			return fmt.Sprintf("#%d", w), 2
		}
		return fmt.Sprintf("#$%x", uint16(w)), 2
	case 2:
		if pc+4 > n {
			return "#<trunc>", 0
		}
		l := binary.BigEndian.Uint32(code[pc:])
		return fmt.Sprintf("#$%x", l), 4
	}
	return "#?", 0
}

// TestableDecode is a wrapper around decode for testing purposes.
func TestableDecode(op uint16, pc int, code []byte) (string, string, int) {
	return decode(op, pc, code)
}

func formatDisp8(v int8) string {
	if v >= -9 && v <= 9 {
		return fmt.Sprintf("%d", v)
	}
	return fmt.Sprintf("$%x", uint8(v))
}

func formatDisp16(v int16) string {
	if v >= -9 && v <= 9 {
		return fmt.Sprintf("%d", v)
	}
	return fmt.Sprintf("$%x", uint16(v))
}

func formatDisp(v int64) string {
	if v >= 0 {
		return fmt.Sprintf("+%d", v)
	}
	return fmt.Sprintf("%d", v)
}

// labelName generates a label string based on the address and its context.
func labelName(addr uint32, labelType LabelType) string {
	prefix := "loc_"
	switch labelType {
	case SubroutineEntry:
		prefix = "sub_"
	}
	return fmt.Sprintf("%s%04X", prefix, addr)
}
