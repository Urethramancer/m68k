package assembler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

// Mnemonic represents a parsed instruction mnemonic.
type Mnemonic struct {
	Value string
	Size  cpu.Size
}

// Operand represents a parsed instruction operand.
type Operand struct {
	Mode           uint16
	Register       uint16
	ExtensionWords []uint16
	Raw            string
	Label          string
}

// IsImmediate returns true if this operand is an immediate constant.
func (o *Operand) IsImmediate() bool {
	return o.Mode == cpu.ModeOther && o.Register == cpu.RegImmediate
}

var (
	reDataRegister       = regexp.MustCompile(`(?i)^d([0-7])$`)
	reAddressRegister    = regexp.MustCompile(`(?i)^a([0-7])$`)
	reAddressIndirect    = regexp.MustCompile(`(?i)^\(a([0-7])\)$`)
	reAddressPostInc     = regexp.MustCompile(`(?i)^\(a([0-7])\)\+$`)
	reAddressPreDec      = regexp.MustCompile(`(?i)^-\(a([0-7])\)$`)
	reAddressDisp        = regexp.MustCompile(`(?i)^([a-fA-F0-9\$\-%]+)\(a([0-7])\)$`)
	reImmediate          = regexp.MustCompile(`(?i)^#(.+)$`)
	reAbsoluteParenShort = regexp.MustCompile(`(?i)^\(([a-fA-F0-9\$\-%]+)\)\.w$`)
	reAbsoluteParenLong  = regexp.MustCompile(`(?i)^\(([a-fA-F0-9\$\-%]+)\)\.l$`)
	reAbsoluteDollarSize = regexp.MustCompile(`(?i)^\$([a-fA-F0-9]+)\.(w|l)$`)
	reAddressIndex       = regexp.MustCompile(`(?i)^([a-fA-F0-9\$\-%]*)\(a([0-7]),(d|a)([0-7])\.(w|l)\)$`)
	rePCRelDispParen     = regexp.MustCompile(`(?i)^\(([a-fA-F0-9\$\-%]+),\s*pc\)$`)
	rePCRelDisp          = regexp.MustCompile(`(?i)^([a-zA-Z0-9_\$\-%]+)\(pc\)$`)
	rePCRelIndex         = regexp.MustCompile(`(?i)^([a-fA-F0-9\$\-%]*)\(pc,(d|a)([0-7])\.(w|l)\)$`)
	reAbsoluteSimple     = regexp.MustCompile(`(?i)^\$[a-fA-F0-9]+$`)
	reLabel              = regexp.MustCompile(`(?i)^[a-z_][a-z0-9_]*$`)
)

// ParseMnemonic splits an instruction like "MOVE.W" → ("move", SizeWord).
func ParseMnemonic(s string) (Mnemonic, error) {
	parts := strings.Split(strings.ToLower(s), ".")
	mn := Mnemonic{Value: parts[0], Size: cpu.SizeInvalid}
	if len(parts) > 1 {
		switch parts[1] {
		case "b", "s":
			mn.Size = cpu.SizeByte
		case "w":
			mn.Size = cpu.SizeWord
		case "l":
			mn.Size = cpu.SizeLong
		default:
			return mn, fmt.Errorf("invalid size suffix: %s", parts[1])
		}
	}
	return mn, nil
}

// parseOperand converts an operand string into a structured Operand.
func parseOperand(s string, asm *Assembler) (Operand, error) {
	s = strings.TrimSpace(s)
	lcs := strings.ToLower(s)

	if lcs == "sr" || lcs == "ccr" || lcs == "usp" {
		op := Operand{Raw: s}
		// Special value to identify these registers later
		op.Mode = cpu.ModeOther
		op.Register = 0xFFFF
		return op, nil
	}

	op := Operand{Raw: s}

	// Indexed and PC-relative index modes
	if m := reAddressIndex.FindStringSubmatch(s); m != nil {
		return parseAddressIndex(m, asm)
	}
	if m := rePCRelIndex.FindStringSubmatch(s); m != nil {
		return parsePCRelIndex(m, asm)
	}

	// Parenthesized PC-relative: (disp,pc) or (label,pc)
	if m := rePCRelDispParen.FindStringSubmatch(s); m != nil {
		inner := m[1]
		// numeric displacement?
		if val, err := parseConstant(inner, asm); err == nil {
			op.Mode = cpu.ModeOther
			op.Register = cpu.ModePCRelative
			op.ExtensionWords = []uint16{uint16(int16(val))}
			return op, nil
		}
		// otherwise treat as a label to be resolved later
		op.Mode = cpu.ModeOther
		op.Register = cpu.ModePCRelative
		op.Label = strings.ToLower(inner)
		return op, nil
	}

	// PC relative displacement (label(pc) or $hex(pc))
	if m := rePCRelDisp.FindStringSubmatch(s); m != nil {
		op.Mode = cpu.ModeOther
		op.Register = cpu.ModePCRelative // Mark as explicit PC-relative
		inner := m[1]
		if val, err := parseConstant(inner, asm); err == nil {
			op.ExtensionWords = []uint16{uint16(int16(val))}
		} else {
			op.Label = strings.ToLower(inner)
		}
		return op, nil
	}

	// Absolute short and long — parenthesized forms ( ($val).w / ($val).l )
	if m := reAbsoluteParenShort.FindStringSubmatch(s); m != nil {
		val, err := parseConstant(m[1], asm)
		if err != nil {
			return op, err
		}
		op.Mode = cpu.ModeOther
		op.Register = cpu.RegAbsShort
		op.ExtensionWords = []uint16{uint16(val)}
		return op, nil
	}

	if m := reAbsoluteParenLong.FindStringSubmatch(s); m != nil {
		val, err := parseConstant(m[1], asm)
		if err != nil {
			return op, err
		}
		op.Mode = cpu.ModeOther
		op.Register = cpu.RegAbsLong
		op.ExtensionWords = []uint16{uint16(val >> 16), uint16(val)}
		return op, nil
	}

	// Absolute forms like $xxxx.w / $xxxx.l (support $hex.w / $hex.l)
	if m := reAbsoluteDollarSize.FindStringSubmatch(s); m != nil {
		valStr := m[1]
		size := strings.ToLower(m[2])
		val, err := strconv.ParseInt(valStr, 16, 64)
		if err != nil {
			return op, fmt.Errorf("invalid hex constant: %s", valStr)
		}
		op.Mode = cpu.ModeOther
		if size == "w" {
			op.Register = cpu.RegAbsShort
			op.ExtensionWords = []uint16{uint16(val)}
		} else {
			op.Register = cpu.RegAbsLong
			op.ExtensionWords = []uint16{uint16(val >> 16), uint16(val)}
		}
		return op, nil
	}

	// Registers and address modes
	if m := reDataRegister.FindStringSubmatch(s); m != nil {
		reg, _ := strconv.Atoi(m[1])
		op.Mode = cpu.ModeData
		op.Register = uint16(reg)
		return op, nil
	}

	if m := reAddressRegister.FindStringSubmatch(s); m != nil {
		reg, _ := strconv.Atoi(m[1])
		op.Mode = cpu.ModeAddr
		op.Register = uint16(reg)
		return op, nil
	}

	if m := reAddressIndirect.FindStringSubmatch(s); m != nil {
		reg, _ := strconv.Atoi(m[1])
		op.Mode = cpu.ModeAddrInd
		op.Register = uint16(reg)
		return op, nil
	}

	if m := reAddressPostInc.FindStringSubmatch(s); m != nil {
		reg, _ := strconv.Atoi(m[1])
		op.Mode = cpu.ModeAddrPostInc
		op.Register = uint16(reg)
		return op, nil
	}

	if m := reAddressPreDec.FindStringSubmatch(s); m != nil {
		reg, _ := strconv.Atoi(m[1])
		op.Mode = cpu.ModeAddrPreDec
		op.Register = uint16(reg)
		return op, nil
	}

	if m := reAddressDisp.FindStringSubmatch(s); m != nil {
		disp, err := parseConstant(m[1], asm)
		if err != nil {
			return op, err
		}
		reg, _ := strconv.Atoi(m[2])
		op.Mode = cpu.ModeAddrDisp
		op.Register = uint16(reg)
		// Use signed 16-bit displacement stored as a single extension word
		op.ExtensionWords = []uint16{uint16(int16(disp))}
		return op, nil
	}

	// Immediate (#value)
	if m := reImmediate.FindStringSubmatch(s); m != nil {
		val, err := parseConstant(m[1], asm)
		if err != nil {
			return op, err
		}
		op.Mode = cpu.ModeOther
		op.Register = cpu.RegImmediate
		if val > 0xFFFF || val < -32768 {
			op.ExtensionWords = []uint16{uint16(val >> 16), uint16(val)}
		} else {
			op.ExtensionWords = []uint16{uint16(val)}
		}
		return op, nil
	}

	// Absolute numeric without explicit size (e.g. $DEAD)
	if m := reAbsoluteSimple.FindStringSubmatch(s); m != nil {
		val, err := parseConstant(m[0], asm)
		if err != nil {
			return op, err
		}
		if val <= 0xFFFF {
			op.Mode = cpu.ModeOther
			op.Register = cpu.RegAbsShort
			op.ExtensionWords = []uint16{uint16(val)}
		} else {
			op.Mode = cpu.ModeOther
			op.Register = cpu.RegAbsLong
			op.ExtensionWords = []uint16{uint16(val >> 16), uint16(val)}
		}
		return op, nil
	}

	// Bare label (checked after all other patterns fail)
	if reLabel.MatchString(s) {
		op.Mode = cpu.ModeOther
		op.Register = RegLabel // Use the placeholder flag
		op.Label = strings.ToLower(s)
		return op, nil
	}

	return op, fmt.Errorf("unknown operand format: %s", s)
}

// parseAddressIndex handles (d8,An,Xn)
func parseAddressIndex(m []string, asm *Assembler) (Operand, error) {
	op := Operand{Raw: m[0], Mode: cpu.ModeAddrIndex}
	var ext uint16

	var disp int64
	if m[1] != "" {
		var err error
		disp, err = parseConstant(m[1], asm)
		if err != nil {
			return op, err
		}
	}
	// signed 8-bit displacement
	ext |= uint16(uint8(int8(disp)))

	an, _ := strconv.Atoi(m[2])
	op.Register = uint16(an)

	xnType := strings.ToLower(m[3])
	xn, _ := strconv.Atoi(m[4])
	ext |= (uint16(xn) << 12)
	if xnType == "a" {
		ext |= 0x8000
	}
	if strings.ToLower(m[5]) == "l" {
		ext |= 0x0800
	}

	op.ExtensionWords = []uint16{ext}
	return op, nil
}

// parsePCRelIndex handles (d8,PC,Xn)
func parsePCRelIndex(m []string, asm *Assembler) (Operand, error) {
	op := Operand{Raw: m[0], Mode: cpu.ModeOther, Register: cpu.RegPCIndex}
	var ext uint16

	var disp int64
	if m[1] != "" {
		var err error
		disp, err = parseConstant(m[1], asm)
		if err != nil {
			return op, err
		}
	}
	ext |= uint16(uint8(int8(disp)))

	xnType := strings.ToLower(m[2])
	xn, _ := strconv.Atoi(m[3])
	ext |= (uint16(xn) << 12)
	if xnType == "a" {
		ext |= 0x8000
	}
	if strings.ToLower(m[4]) == "l" {
		ext |= 0x0800
	}

	op.ExtensionWords = []uint16{ext}
	return op, nil
}

// parseConstant converts numeric or symbolic expressions to int64.
func parseConstant(s string, asm *Assembler) (int64, error) {
	s = strings.TrimSpace(strings.TrimPrefix(s, "#"))

	// Character literal ('A')
	if len(s) >= 3 && s[0] == '\'' && s[len(s)-1] == '\'' {
		return int64(s[1]), nil
	}

	// Symbol lookup
	if asm != nil {
		if val, ok := asm.symbols[strings.ToLower(s)]; ok {
			return val, nil
		}
	}

	base := 10
	switch {
	case strings.HasPrefix(s, "$"):
		s = s[1:]
		base = 16
	case strings.HasPrefix(strings.ToLower(s), "0x"):
		s = s[2:]
		base = 16
	case strings.HasPrefix(s, "%"):
		s = s[1:]
		base = 2
	}

	val, err := strconv.ParseInt(s, base, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number format: %s", s)
	}
	return val, nil
}
