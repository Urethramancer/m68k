package assembler

import (
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

// NodeType defines the type of an assembly node.
type NodeType int

const (
	// NodeInstruction type.
	NodeInstruction NodeType = iota
	// NodeLabel type.
	NodeLabel
	// NodeDirective type.
	NodeDirective
)

// Node represents one parsed element from the assembly source.
type Node struct {
	Type     NodeType
	Label    string
	Mnemonic Mnemonic
	Operands []Operand
	Parts    []string
	Size     uint32
}

// GetSize calculates the final encoded size in bytes for the node.
func (n *Node) GetSize(asm *Assembler, pc uint32) (uint32, error) {
	switch n.Type {
	case NodeLabel:
		return 0, nil

	case NodeDirective:
		return asm.getDirectiveSize(n)

	case NodeInstruction:
		// --- Quick handling for special CPU registers (SR/CCR/USP) ---
		if len(n.Parts) > 1 {
			first := strings.ToLower(n.Parts[1])
			last := strings.ToLower(n.Parts[len(n.Parts)-1])

			if first == "sr" || last == "sr" || first == "ccr" || last == "ccr" {
				// Immediate to SR/CCR is 4 bytes
				if len(n.Parts) > 1 && strings.HasPrefix(n.Parts[1], "#") {
					return 4, nil
				}
				// Register or EA to/from SR/CCR
				op, err := parseOperand(n.Parts[1], asm)
				if err != nil {
					return 0, err
				}
				return 2 + uint32(len(op.ExtensionWords))*2, nil
			}

			if first == "usp" || last == "usp" {
				return 2, nil
			}
		}

		// --- Operand string extraction ---
		var opStrings []string
		if len(n.Parts) > 1 {
			opStrings = splitOperands(n.Parts[1])
		}

		// --- Instruction-specific adjustments ---
		switch strings.ToLower(n.Mnemonic.Value) {

		case "lea":
			// LEA label or displacement-based address: usually 4 bytes
			if len(n.Operands) == 2 {
				_, err := parseOperand(n.Operands[0].Raw, asm)
				if err != nil {
					return 4, nil
				}
			}

		case "movem":
			size := uint32(4) // opword + register mask
			opStr := n.Parts[1]
			if _, err := parseMovemList(opStr); err == nil && len(n.Parts) > 2 {
				opStr = n.Parts[2]
			}
			if op, err := parseOperand(opStr, asm); err == nil {
				size += uint32(len(op.ExtensionWords) * 2)
			}
			return size, nil

		case "movep":
			// MOVEP always has one word displacement
			return 4, nil

		case "chk":
			op, err := parseOperand(opStrings[0], asm)
			if err == nil {
				return 2 + uint32(len(op.ExtensionWords))*2, nil
			}
			return 2, nil

		case "move":
			if len(n.Operands) >= 2 && CanBeMoveq(n.Mnemonic, n.Operands[0], n.Operands[1], asm) {
				return 2, nil
			}

		case "add", "sub":
			if len(n.Operands) >= 1 && CanBeAddqSubq(n.Mnemonic, n.Operands[0], asm) {
				return 2, nil
			}

		case "jmp", "jsr":
			if len(n.Operands) == 1 {
				if _, err := parseOperand(n.Operands[0].Raw, asm); err != nil {
					return 6, nil
				}
			}

		// Branches
		case "bra", "bsr", "bhi", "bls", "bcc", "bcs", "bne", "beq", "bvc", "bvs", "bpl", "bmi", "bge", "blt", "bgt", "ble":
			return getSizeBra(n, asm, pc)

		// DBcc variants
		case "dbra", "dbhi", "dbls", "dbcc", "dbcs", "dbne", "dbeq", "dbvc", "dbvs", "dbpl", "dbmi", "dbge", "dblt", "dbgt", "dble":
			return 4, nil

		case "unlk", "rts", "rte", "nop", "reset", "exg", "addx", "subx", "trap", "illegal", "rtr", "trapv", "negx":
			return 2, nil

		case "stop":
			if len(opStrings) > 0 {
				return 4, nil
			}
			return 2, nil

		case "link":
			return 4, nil
		}

		// --- Default size calculation ---
		size := uint32(2) // base opword
		for _, s := range opStrings {
			op, err := parseOperand(s, asm)
			if err != nil {
				// unresolved label
				if _, ok := asm.labels[s]; ok {
					// label treated as an absolute long (4 bytes) if unresolved
					size += 4
					continue
				}
				return 0, err
			}
			// If this operand is a PC-relative type, count a single 16-bit extension.
			// (parseOperand no longer pre-fills ExtensionWords for label placeholders)
			if op.Mode == cpu.ModeOther && op.Register == cpu.ModePCRelative {
				size += 2
			} else {
				size += uint32(len(op.ExtensionWords) * 2)
			}
		}
		return size, nil
	}

	return 0, nil
}
