package assembler

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

// assembleMove handles MOVE, MOVEA, and MOVEQ instructions.
func assembleMove(mn Mnemonic, operands []Operand, asm *Assembler, pc uint32) ([]uint16, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("MOVE requires 2 operands")
	}
	src, dst := operands[0], operands[1]

	// Label → absolute long addressing.
	if target, ok := asm.labels[src.Raw]; ok && !src.IsImmediate() {
		src.Mode = cpu.ModeOther
		src.Register = cpu.ModeAbsLong
		src.ExtensionWords = []uint16{uint16(target >> 16), uint16(target)}
	}
	if target, ok := asm.labels[dst.Raw]; ok && !dst.IsImmediate() {
		dst.Mode = cpu.ModeOther
		dst.Register = cpu.ModeAbsLong
		dst.ExtensionWords = []uint16{uint16(target >> 16), uint16(target)}
	}

	// --- MOVEQ ---
	if CanBeMoveq(mn, src, dst, asm) {
		val, _ := parseConstant(src.Raw, asm)
		if mn.Size == cpu.SizeWord || mn.Size == cpu.SizeByte {
			return nil, fmt.Errorf("MOVEQ only supports .L size")
		}
		opword := uint16(cpu.OPMOVEQ)
		opword |= (dst.Register << 9)
		opword |= uint16(val) & 0x00FF
		return []uint16{opword}, nil
	}

	// --- MOVEA ---
	if dst.Mode == cpu.ModeAddr {
		var opword uint16
		switch mn.Size {
		case cpu.SizeWord:
			opword = 0x3040
		case cpu.SizeLong:
			opword = 0x2040
		default:
			return nil, fmt.Errorf("MOVEA only supports .W or .L sizes")
		}

		srcBits, srcExt, err := encodeEA(src)
		if err != nil {
			return nil, err
		}

		opword |= (dst.Register << 9)
		opword |= srcBits

		return append([]uint16{opword}, srcExt...), nil
	}

	// --- General MOVE ---
	opword := uint16(cpu.OPMOVE)
	switch mn.Size {
	case cpu.SizeByte:
		opword |= 0x1000
	case cpu.SizeWord:
		opword |= 0x3000
	case cpu.SizeLong:
		opword |= 0x2000
	default:
		return nil, fmt.Errorf("unsupported MOVE size")
	}

	srcBits, srcExt, err := encodeEA(src)
	if err != nil {
		return nil, err
	}
	dstBits, dstExt, err := encodeEA(dst)
	if err != nil {
		return nil, err
	}

	opword |= (dstBits << 6)
	opword |= srcBits

	// Fix PC-relative labels like label(PC)
	if src.Mode == cpu.ModeOther && src.Register == cpu.ModePCRelative {
		re := regexp.MustCompile(`(?i)^([a-zA-Z_][a-zA-Z0-9_]*)\(pc\)$`)
		if m := re.FindStringSubmatch(src.Raw); m != nil {
			label := m[1]
			if target, ok := asm.labels[label]; ok {
				offset := int32(target) - int32(pc) - 2
				if len(srcExt) > 0 {
					srcExt[0] = uint16(offset)
				}
			}
		}
	}

	words := []uint16{opword}
	words = append(words, srcExt...)
	words = append(words, dstExt...)
	return words, nil
}

// CanBeMoveq checks if the instruction can be encoded as MOVEQ.
func CanBeMoveq(mn Mnemonic, src Operand, dst Operand, asm *Assembler) bool {
	name := strings.ToLower(mn.Value)
	if name != "move" && name != "moveq" {
		return false
	}

	if dst.Mode == cpu.ModeData && src.IsImmediate() {
		val, err := parseConstant(src.Raw, asm)
		if err != nil {
			return false
		}
		return val >= -128 && val <= 127
	}
	return false
}
