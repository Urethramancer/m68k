package assembler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

// assembleMovem assembles MOVEM instructions.
func assembleMovem(mn Mnemonic, operands []Operand) ([]uint16, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("MOVEM requires 2 operands")
	}

	src, dst := operands[0], operands[1]
	sz := mn.Size
	if sz == cpu.SizeInvalid {
		sz = cpu.SizeWord
	}
	if sz != cpu.SizeWord && sz != cpu.SizeLong {
		return nil, fmt.Errorf("MOVEM only supports .W or .L sizes")
	}

	// MOVEM <reglist>, <ea> — store
	if strings.Contains(src.Raw, "/") {
		return assembleMovemStore(src, dst, sz)
	}

	// MOVEM <ea>, <reglist> — load
	if strings.Contains(dst.Raw, "/") {
		return assembleMovemLoad(src, dst, sz)
	}

	return nil, fmt.Errorf("invalid MOVEM syntax: must include register list")
}

// Store form: MOVEM <reglist>, <ea>
func assembleMovemStore(src Operand, dst Operand, sz cpu.Size) ([]uint16, error) {
	regmask, err := parseMovemList(src.Raw)
	if err != nil {
		return nil, err
	}

	opword := uint16(cpu.OPMOVEM)
	if sz == cpu.SizeLong {
		opword |= 0x0040
	}

	dstEA, dstExt, err := encodeEA(dst)
	if err != nil {
		return nil, err
	}

	opword |= dstEA

	// Pre-decrement mode reverses register order.
	if dst.Mode == cpu.ModeAddrPreDec {
		// 68k stores from highest to lowest register: A7→D0
		regmask = reverseMovemMask(regmask)
	}

	return append([]uint16{opword, regmask}, dstExt...), nil
}

// Load form: MOVEM <ea>, <reglist>
func assembleMovemLoad(src Operand, dst Operand, sz cpu.Size) ([]uint16, error) {
	regmask, err := parseMovemList(dst.Raw)
	if err != nil {
		return nil, err
	}

	opword := uint16(cpu.OPMOVEM)
	if sz == cpu.SizeLong {
		opword |= 0x0040
	}

	srcEA, srcExt, err := encodeEA(src)
	if err != nil {
		return nil, err
	}

	opword |= 0x0400 // direction = memory → registers
	opword |= srcEA

	return append([]uint16{opword, regmask}, srcExt...), nil
}

// Parse register list (e.g. "d0-d3/a1/a3")
func parseMovemList(list string) (uint16, error) {
	var mask uint16
	parts := strings.Split(list, "/")

	for _, p := range parts {
		p = strings.TrimSpace(strings.ToLower(p))
		if p == "" {
			continue
		}

		if strings.Contains(p, "-") {
			rng := strings.Split(p, "-")
			if len(rng) != 2 {
				return 0, fmt.Errorf("invalid MOVEM range: %s", p)
			}

			if (rng[0][0] == 'd' && rng[1][0] == 'a') || (rng[0][0] == 'a' && rng[1][0] == 'd') {
				return 0, fmt.Errorf("MOVEM range cannot cross Dn/An groups: %s", p)
			}

			start, err1 := parseRegIndex(rng[0])
			end, err2 := parseRegIndex(rng[1])
			if err1 != nil || err2 != nil {
				return 0, fmt.Errorf("invalid MOVEM range: %s", p)
			}
			if start > end {
				start, end = end, start
			}
			for i := start; i <= end; i++ {
				mask |= 1 << i
			}
		} else {
			idx, err := parseRegIndex(p)
			if err != nil {
				return 0, fmt.Errorf("invalid MOVEM register: %s", p)
			}
			mask |= 1 << idx
		}
	}

	return mask, nil
}

// Parse register name (e.g. "d0", "a6")
func parseRegIndex(reg string) (int, error) {
	reg = strings.TrimSpace(strings.ToLower(reg))
	if len(reg) < 2 {
		return 0, fmt.Errorf("invalid register name: %s", reg)
	}

	numStr := reg[1:]
	num, err := strconv.Atoi(numStr)
	if err != nil || num < 0 || num > 7 {
		return 0, fmt.Errorf("invalid register number: %s", reg)
	}

	switch reg[0] {
	case 'd':
		return num, nil
	case 'a':
		return num + 8, nil
	default:
		return 0, fmt.Errorf("invalid register type: %s", reg)
	}
}
