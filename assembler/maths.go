package assembler

import (
	"fmt"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

// CanBeAddqSubq checks if an ADD/SUB instruction can be optimized to ADDQ/SUBQ.
func CanBeAddqSubq(mn Mnemonic, src Operand, asm *Assembler) bool {
	m := strings.ToLower(mn.Value)
	if m != "addq" && m != "subq" {
		return false
	}
	if !src.IsImmediate() {
		return false
	}
	val, err := parseConstant(src.Raw, asm)
	return err == nil && val >= 1 && val <= 8
}

// assembleMath handles all integer arithmetic instructions.
func assembleMath(mn Mnemonic, operands []Operand, asm *Assembler) ([]uint16, error) {
	switch strings.ToLower(mn.Value) {
	case "add", "adda", "addq", "addi":
		return assembleAdd(mn, operands, asm)
	case "sub", "suba", "subq", "subi":
		return assembleSub(mn, operands, asm)
	case "addx", "subx":
		return assembleAddxSubx(mn, operands)
	case "muls", "mulu":
		return assembleMul(mn, operands)
	case "divs", "divu":
		return assembleDiv(mn, operands)
	case "neg", "negx":
		return assembleMisc(mn, operands)
	}
	return nil, fmt.Errorf("unknown math instruction: %s", mn.Value)
}

// ---------------- ADD ----------------

func assembleAdd(mn Mnemonic, operands []Operand, asm *Assembler) ([]uint16, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("ADD requires 2 operands")
	}
	src, dst := operands[0], operands[1]

	// --- ADDQ optimization ---
	if CanBeAddqSubq(mn, src, asm) {
		opword := uint16(cpu.OPADDQ)
		val, _ := parseConstant(src.Raw, asm)
		data := uint16(val)
		if val == 8 {
			data = 0
		}
		opword |= (data << 9)

		var err error
		opword, err = setOpwordSize(opword, mn.Size, SizeBits)
		if err != nil {
			return nil, err
		}

		eaBits, ext, err := encodeEA(dst)
		if err != nil {
			return nil, err
		}
		opword |= eaBits
		return append([]uint16{opword}, ext...), nil
	}

	// --- ADDI (immediate source) ---
	if src.IsImmediate() {
		opword := uint16(cpu.OPADDI)
		var err error
		opword, err = setOpwordSize(opword, mn.Size, SizeBitsSingleOp)
		if err != nil {
			return nil, err
		}

		eaBits, ext, err := encodeEA(dst)
		if err != nil {
			return nil, err
		}
		opword |= eaBits

		// Build immediate extension based on instruction size, not value size.
		val, err := parseConstant(src.Raw, asm)
		if err != nil {
			return nil, err
		}

		var immExt []uint16
		switch mn.Size {
		case cpu.SizeLong:
			immExt = []uint16{uint16(val >> 16), uint16(val)}
		default: // .b or .w
			immExt = []uint16{uint16(val)}
		}

		return append([]uint16{opword}, append(immExt, ext...)...), nil
	}

	// --- ADDA (destination is address register) ---
	if dst.Mode == cpu.ModeAddr {
		opword := uint16(cpu.OPADDA)
		var err error
		opword, err = setOpwordSize(opword, mn.Size, SizeBitsAddr)
		if err != nil {
			return nil, err
		}
		opword |= (dst.Register << 9)

		eaBits, ext, err := encodeEA(src)
		if err != nil {
			return nil, err
		}
		opword |= eaBits
		return append([]uint16{opword}, ext...), nil
	}

	// --- Standard ADD (register or memory destination) ---
	opword := uint16(cpu.OPADD)
	var err error
	opword, err = setOpwordSize(opword, mn.Size, SizeBits)
	if err != nil {
		return nil, err
	}

	var eaBits uint16
	var ext []uint16

	if dst.Mode == cpu.ModeData {
		opword |= (dst.Register << 9)
		eaBits, ext, err = encodeEA(src)
	} else {
		opword |= 0x0100 // direction bit: Dn to EA
		opword |= (src.Register << 9)
		eaBits, ext, err = encodeEA(dst)
	}
	if err != nil {
		return nil, err
	}

	opword |= eaBits
	return append([]uint16{opword}, ext...), nil
}

// ---------------- SUB ----------------

func assembleSub(mn Mnemonic, operands []Operand, asm *Assembler) ([]uint16, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("SUB requires 2 operands")
	}
	src, dst := operands[0], operands[1]

	// --- SUBQ optimization ---
	if CanBeAddqSubq(mn, src, asm) {
		opword := uint16(cpu.OPSUBQ)
		val, _ := parseConstant(src.Raw, asm)
		data := uint16(val)
		if val == 8 {
			data = 0
		}
		opword |= (data << 9)

		var err error
		opword, err = setOpwordSize(opword, mn.Size, SizeBits)
		if err != nil {
			return nil, err
		}

		eaBits, ext, err := encodeEA(dst)
		if err != nil {
			return nil, err
		}
		opword |= eaBits
		return append([]uint16{opword}, ext...), nil
	}

	// --- SUBI (immediate source) ---
	if src.IsImmediate() {
		opword := uint16(cpu.OPSUBI)
		var err error
		opword, err = setOpwordSize(opword, mn.Size, SizeBitsSingleOp)
		if err != nil {
			return nil, err
		}

		eaBits, ext, err := encodeEA(dst)
		if err != nil {
			return nil, err
		}
		opword |= eaBits

		// Build immediate extension based on instruction size, not value size.
		val, err := parseConstant(src.Raw, asm)
		if err != nil {
			return nil, err
		}

		var immExt []uint16
		switch mn.Size {
		case cpu.SizeLong:
			immExt = []uint16{uint16(val >> 16), uint16(val)}
		default: // .b or .w
			immExt = []uint16{uint16(val)}
		}

		return append([]uint16{opword}, append(immExt, ext...)...), nil
	}

	// --- SUBA ---
	if dst.Mode == cpu.ModeAddr {
		opword := uint16(cpu.OPSUBA)
		var err error
		opword, err = setOpwordSize(opword, mn.Size, SizeBitsAddr)
		if err != nil {
			return nil, err
		}
		opword |= (dst.Register << 9)

		eaBits, ext, err := encodeEA(src)
		if err != nil {
			return nil, err
		}
		opword |= eaBits
		return append([]uint16{opword}, ext...), nil
	}

	// --- Standard SUB ---
	opword := uint16(cpu.OPSUB)
	var err error
	opword, err = setOpwordSize(opword, mn.Size, SizeBits)
	if err != nil {
		return nil, err
	}

	var eaBits uint16
	var ext []uint16

	if dst.Mode == cpu.ModeData {
		opword |= (dst.Register << 9)
		eaBits, ext, err = encodeEA(src)
	} else {
		opword |= 0x0100
		opword |= (src.Register << 9)
		eaBits, ext, err = encodeEA(dst)
	}
	if err != nil {
		return nil, err
	}

	opword |= eaBits
	return append([]uint16{opword}, ext...), nil
}

// ---------------- ADDX / SUBX ----------------

func assembleAddxSubx(mn Mnemonic, operands []Operand) ([]uint16, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("%s requires 2 operands", strings.ToUpper(mn.Value))
	}
	src, dst := operands[0], operands[1]

	var opword uint16
	if mn.Value == "addx" {
		opword = uint16(cpu.OPADDX)
	} else {
		opword = uint16(cpu.OPSUBX)
	}

	var err error
	opword, err = setOpwordSize(opword, mn.Size, SizeBits)
	if err != nil {
		return nil, err
	}

	if src.Mode == cpu.ModeData && dst.Mode == cpu.ModeData {
		opword |= (dst.Register << 9) | src.Register
	} else if src.Mode == cpu.ModeAddrPreDec && dst.Mode == cpu.ModeAddrPreDec {
		opword |= 0x0008
		opword |= (dst.Register << 9) | src.Register
	} else {
		return nil, fmt.Errorf("invalid operand combination for %s", strings.ToUpper(mn.Value))
	}
	return []uint16{opword}, nil
}

// ---------------- MUL / DIV ----------------

func assembleMul(mn Mnemonic, operands []Operand) ([]uint16, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("MUL requires 2 operands (<ea>, Dn)")
	}
	src, dst := operands[0], operands[1]

	if dst.Mode != cpu.ModeData {
		return nil, fmt.Errorf("destination of MUL must be a data register")
	}
	if mn.Size != cpu.SizeWord && mn.Size != cpu.SizeInvalid {
		return nil, fmt.Errorf("MUL only supports word size (.w)")
	}

	var opword uint16
	if mn.Value == "muls" {
		opword = uint16(cpu.OPMULS)
	} else {
		opword = uint16(cpu.OPMULU)
	}
	opword |= (dst.Register << 9)

	eaBits, ext, err := encodeEA(src)
	if err != nil {
		return nil, err
	}
	opword |= eaBits
	return append([]uint16{opword}, ext...), nil
}

func assembleDiv(mn Mnemonic, operands []Operand) ([]uint16, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("DIV requires 2 operands (<ea>, Dn)")
	}
	src, dst := operands[0], operands[1]

	if dst.Mode != cpu.ModeData {
		return nil, fmt.Errorf("destination of DIV must be a data register")
	}
	if mn.Size != cpu.SizeWord && mn.Size != cpu.SizeInvalid {
		return nil, fmt.Errorf("DIV only supports word size (.w)")
	}

	var opword uint16
	if mn.Value == "divs" {
		opword = uint16(cpu.OPDIVS)
	} else {
		opword = uint16(cpu.OPDIVU)
	}
	opword |= (dst.Register << 9)

	eaBits, ext, err := encodeEA(src)
	if err != nil {
		return nil, err
	}
	opword |= eaBits
	return append([]uint16{opword}, ext...), nil
}
