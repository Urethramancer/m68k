package assembler

import (
	"fmt"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

// assembleFlow dispatches to the correct flow-control assembly function.
func assembleFlow(mn Mnemonic, operands []Operand, labels map[string]uint32, pc uint32, size uint32) ([]uint16, error) {
	switch mn.Value {
	case "jmp", "jsr":
		return assembleJmpJsr(mn, operands, labels)
	case "rts":
		return assembleRts()
	case "rtr":
		return assembleRtr()
	case "rte":
		return assembleRte()
	case "bra", "bsr", "bhi", "bls", "bcc", "bcs", "bne", "beq", "bvc", "bvs", "bpl", "bmi", "bge", "blt", "bgt", "ble":
		return assembleBra(mn, operands, labels, pc, size)
	}
	return nil, fmt.Errorf("unknown flow instruction: %s", mn.Value)
}

// getSizeBra calculates the optimal size for a branch instruction during the sizing pass.
func getSizeBra(n *Node, asm *Assembler, pc uint32) uint32 {
	// If size is explicitly specified (e.g., bra.s), respect it.
	if n.Mnemonic.Size == cpu.SizeByte {
		return 2
	}
	if n.Mnemonic.Size == cpu.SizeWord {
		return 4
	}

	// If no operand, it's an error, but for sizing assume short.
	if len(n.Operands) == 0 {
		return 2
	}

	label := strings.ToLower(strings.TrimSpace(n.Operands[0].Raw))
	target, ok := asm.labels[label]
	if !ok {
		// Forward reference: assume long branch (worst case) to be safe.
		return 4
	}

	offset := int32(target) - int32(pc+2)
	if offset >= -128 && offset <= 127 {
		return 2 // Fits in a short branch.
	}
	return 4 // Requires a long branch.
}

// JMP / JSR

func assembleJmpJsr(mn Mnemonic, operands []Operand, labels map[string]uint32) ([]uint16, error) {
	if len(operands) != 1 {
		return nil, fmt.Errorf("%s requires 1 operand", strings.ToUpper(mn.Value))
	}
	src := operands[0]

	opword := uint16(cpu.OPJSR)
	if mn.Value == "jmp" {
		opword = cpu.OPJMP
	}

	// Label as absolute long
	if target, ok := labels[strings.ToLower(src.Raw)]; ok {
		if mn.Value == "jmp" {
			return []uint16{0x4EF9, uint16(target >> 16), uint16(target)}, nil
		}
		return []uint16{0x4EB9, uint16(target >> 16), uint16(target)}, nil
	}

	// Otherwise encode EA
	eaBits, eaExt, err := encodeEA(src)
	if err != nil {
		return nil, err
	}
	opword |= eaBits

	return append([]uint16{opword}, eaExt...), nil
}

// Branches (BRA/BSR/Bcc)

func assembleBra(mn Mnemonic, operands []Operand, labels map[string]uint32, pc uint32, size uint32) ([]uint16, error) {
	if len(operands) != 1 {
		return nil, fmt.Errorf("branch instruction requires 1 operand")
	}
	label := strings.ToLower(strings.TrimSpace(operands[0].Raw))

	baseOpcode, ok := cpu.BranchOpcodes[mn.Value]
	if !ok {
		return nil, fmt.Errorf("unknown branch type: %s", mn.Value)
	}

	target, ok := labels[label]
	if !ok {
		return nil, fmt.Errorf("undefined label: %s", label)
	}

	offset := int32(target) - int32(pc+2)
	if size == 2 {
		if offset < -128 || offset > 127 {
			return nil, fmt.Errorf("short branch to '%s' out of range (%d)", label, offset)
		}
		baseOpcode |= uint16(offset & 0xFF)
		return []uint16{baseOpcode}, nil
	}

	if offset < -32768 || offset > 32767 {
		return nil, fmt.Errorf("branch to '%s' out of range (%d)", label, offset)
	}
	return []uint16{baseOpcode, uint16(offset & 0xFFFF)}, nil
}

// Scc (Set Conditional)

func assembleScc(mn Mnemonic, operands []Operand) ([]uint16, error) {
	if len(operands) != 1 {
		return nil, fmt.Errorf("Scc requires 1 operand")
	}
	dst := operands[0]

	condStr := strings.ToLower(strings.TrimPrefix(mn.Value, "s"))
	condCode, ok := cpu.ConditionCodes[condStr]
	if !ok {
		return nil, fmt.Errorf("unknown condition code '%s' for Scc", condStr)
	}

	if dst.Mode == cpu.ModeAddr {
		return nil, fmt.Errorf("Scc destination cannot be an address register")
	}

	opword := uint16(cpu.OPScc)
	opword |= condCode << 8
	opword |= (dst.Mode << 3) | dst.Register

	return append([]uint16{opword}, dst.ExtensionWords...), nil
}

// DBcc (Decrement & Branch Conditional)

func assembleDbcc(mn Mnemonic, operands []Operand, labels map[string]uint32, pc uint32) ([]uint16, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("DBcc requires 2 operands (Dn, label)")
	}
	src, dst := operands[0], operands[1]

	if src.Mode != cpu.ModeData {
		return nil, fmt.Errorf("first operand of DBcc must be a data register")
	}

	condStr := strings.ToLower(strings.TrimPrefix(mn.Value, "db"))
	if condStr == "ra" {
		condStr = "f"
	}

	condCode, ok := cpu.ConditionCodes[condStr]
	if !ok {
		return nil, fmt.Errorf("unknown condition code '%s' for DBcc", condStr)
	}

	opword := uint16(cpu.OPDBcc)
	opword |= condCode << 8
	opword |= src.Register

	labelName := strings.ToLower(strings.TrimSpace(dst.Raw))
	target, ok := labels[labelName]
	if !ok {
		return nil, fmt.Errorf("undefined label '%s'", labelName)
	}

	offset := int32(target) - int32(pc+2)
	if offset < -32768 || offset > 32767 {
		return nil, fmt.Errorf("branch target out of range for DBcc")
	}

	return []uint16{opword, uint16(offset & 0xFFFF)}, nil
}

// Returns

func assembleRts() ([]uint16, error) { return []uint16{cpu.OPRTS}, nil }
func assembleRtr() ([]uint16, error) { return []uint16{cpu.OPRTR}, nil }
func assembleRte() ([]uint16, error) { return []uint16{cpu.OPRTE}, nil }
