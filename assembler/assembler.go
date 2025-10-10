package assembler

import (
	"fmt"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

// Assembler holds the state for the assembly process.
type Assembler struct {
	symbols map[string]int64
	labels  map[string]uint32
}

// New creates a new Assembler instance.
func New() *Assembler {
	return &Assembler{
		symbols: make(map[string]int64),
		labels:  make(map[string]uint32),
	}
}

// Assemble takes M68k assembly code and returns the machine code.
func (asm *Assembler) Assemble(src string, baseAddress uint32) ([]byte, error) {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")

	nodes, err := asm.parseLines(lines)
	if err != nil {
		return nil, fmt.Errorf("parsing error: %w", err)
	}

	// Pass: resolve label addresses and node sizes until stable.
	for {
		pc := baseAddress
		changed := false
		for _, n := range nodes {
			if n.Type == NodeLabel {
				if addr, ok := asm.labels[n.Label]; !ok || addr != pc {
					asm.labels[n.Label] = pc
					changed = true
				}
				continue
			}
			if n.Type == NodeDirective && len(n.Parts) > 1 && n.Parts[0] == ".org" {
				addr, err := parseConstant(n.Parts[1], asm)
				if err != nil {
					return nil, err
				}
				pc = uint32(addr)
				continue
			}

			oldSize := n.Size
			size, err := n.GetSize(asm, pc)
			if err != nil {
				return nil, fmt.Errorf("error calculating size for '%v': %w", n.Parts, err)
			}
			if oldSize != size {
				changed = true
			}
			n.Size = size
			pc += size
		}
		if !changed {
			break
		}
	}

	// Generate machine code.
	var machineCode []uint16
	pc := baseAddress
	for _, n := range nodes {
		var code []uint16
		var err error

		switch n.Type {
		case NodeLabel:
			// Labels do not emit code.
			continue
		case NodeDirective:
			code, err = asm.generateDirectiveCode(n)
			if len(n.Parts) > 1 && n.Parts[0] == ".org" {
				addr, _ := parseConstant(n.Parts[1], asm)
				pc = uint32(addr)
			}
		case NodeInstruction:
			code, err = asm.generateInstructionCode(n, pc)
		}

		if err != nil {
			return nil, fmt.Errorf("error generating code for '%v': %w", n.Parts, err)
		}
		machineCode = append(machineCode, code...)
		pc += n.Size
	}

	return cpu.WordsToBytes(machineCode), nil
}

// parseLines converts raw source lines into a slice of Node objects.
func (asm *Assembler) parseLines(lines []string) ([]*Node, error) {
	var nodes []*Node
	for i, line := range lines {
		if commentIndex := strings.IndexRune(line, ';'); commentIndex != -1 {
			line = line[:commentIndex]
		}
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "*") {
			continue
		}

		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			label := strings.TrimSpace(parts[0])
			if !strings.ContainsAny(label, " \t") {
				nodes = append(nodes, &Node{Type: NodeLabel, Label: strings.ToLower(label), Parts: []string{label + ":"}})
				line = strings.TrimSpace(parts[1])
			}
		}

		if line == "" {
			continue
		}

		var mnemonic, operandStr string
		firstSpace := strings.IndexAny(line, " \t")
		if firstSpace == -1 {
			mnemonic = line
		} else {
			mnemonic = line[:firstSpace]
			operandStr = strings.TrimSpace(line[firstSpace:])
		}

		nodeParts := []string{mnemonic}
		if operandStr != "" {
			nodeParts = append(nodeParts, operandStr)
		}

		directiveCheck := strings.ToLower(mnemonic)
		directiveCheck = strings.TrimPrefix(directiveCheck, ".")
		switch directiveCheck {
		case "dc.b", "dc.w", "dc.l", "ds.b", "ds.w", "ds.l", "org", "equ":
			nodes = append(nodes, &Node{Type: NodeDirective, Parts: nodeParts})
			continue
		}

		mn, err := ParseMnemonic(mnemonic)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}

		var operands []Operand
		if operandStr != "" {
			opStrings := splitOperands(operandStr)
			for _, s := range opStrings {
				s = strings.TrimSpace(s)
				op, err := parseOperand(s, asm)
				if err != nil {
					// Keep raw placeholder to allow later patching (labels, unknown EAs).
					operands = append(operands, Operand{Raw: s})
				} else {
					op.Raw = s
					operands = append(operands, op)
				}
			}
		}
		nodes = append(nodes, &Node{Type: NodeInstruction, Mnemonic: mn, Operands: operands, Parts: nodeParts})
	}
	return nodes, nil
}

// generateInstructionCode resolves operands (including forward labels), patches PC-relative
// candidates, and dispatches to the appropriate instruction assembler.
func (asm *Assembler) generateInstructionCode(n *Node, pc uint32) ([]uint16, error) {
	// Resolve operands; preserve raw/label if parsing fails (forward label).
	operands := make([]Operand, 0, len(n.Operands))
	for _, op := range n.Operands {
		resolved, err := parseOperand(op.Raw, asm)
		if err != nil {
			// keep original (unresolved) operand
			operands = append(operands, op)
			continue
		}
		resolved.Raw = op.Raw
		if resolved.Label == "" && op.Label != "" {
			resolved.Label = op.Label
		}
		operands = append(operands, resolved)
	}

	// Patch known labels into PC-relative or absolute-long encoding when appropriate.
	for i := range operands {
		op := &operands[i]
		if op.Mode != cpu.ModeOther {
			continue
		}

		// derive label name from op.Label (preferred) or op.Raw
		labelName := op.Label
		if labelName == "" {
			raw := strings.ToLower(strings.TrimSpace(op.Raw))
			if strings.HasSuffix(raw, "(pc)") {
				labelName = strings.TrimSpace(strings.TrimSuffix(raw, "(pc)"))
			} else {
				labelName = strings.Trim(raw, "() ")
			}
		}
		if labelName == "" {
			continue
		}

		target, ok := asm.labels[strings.ToLower(labelName)]
		if !ok {
			continue
		}
		offset := int32(target) - int32(pc) - 2

		// If parser created a PC-relative placeholder, fill displacement.
		if op.Register == cpu.ModePCRelative {
			op.ExtensionWords = []uint16{uint16(int16(offset))}
			continue
		}

		// If parser left an absolute-long placeholder (2 ext words), convert to PC-relative
		// when offset fits; otherwise ensure extension words contain the actual absolute address.
		if len(op.ExtensionWords) == 2 {
			if offset >= -32768 && offset <= 32767 {
				op.Register = cpu.ModePCRelative
				op.ExtensionWords = []uint16{uint16(int16(offset))}
			} else {
				op.ExtensionWords = []uint16{uint16(target >> 16), uint16(target)}
				op.Register = cpu.RegAbsLong
			}
			continue
		}

		// If Reg says AbsLong but ext words empty (bare label case), fill appropriately.
		if op.Register == cpu.RegAbsLong && len(op.ExtensionWords) == 0 {
			if offset >= -32768 && offset <= 32767 {
				op.Register = cpu.ModePCRelative
				op.ExtensionWords = []uint16{uint16(int16(offset))}
			} else {
				op.ExtensionWords = []uint16{uint16(target >> 16), uint16(target)}
			}
		}
	}

	// write back patched operands
	n.Operands = operands

	// Special-case: operations involving SR/CCR/USP must go through assembleStatus.
	// e.g. MOVE <ea>, SR  ; MOVE SR, <ea> ; ANDI #<val>, SR ; MOVE <ea>, USP ; MOVE USP, <ea>
	if len(operands) > 0 {
		for i := range operands {
			raw := strings.ToLower(strings.TrimSpace(operands[i].Raw))
			if raw == "sr" || raw == "ccr" || raw == "usp" {
				return assembleStatus(n.Mnemonic, n.Operands, asm)
			}
		}
	}

	// Dispatch to assembler functions.
	switch n.Mnemonic.Value {
	case "movem":
		return assembleMovem(n.Mnemonic, n.Operands)
	case "movep":
		return assembleMovep(n.Mnemonic, n.Operands, asm)
	case "move", "movea", "moveq":
		return assembleMove(n.Mnemonic, n.Operands, asm, pc)
	case "add", "adda", "sub", "suba", "mulu", "muls", "divu", "divs",
		"addx", "subx", "addq", "subq", "addi", "subi":
		return assembleMath(n.Mnemonic, n.Operands, asm)
	case "and", "or", "eor", "not", "andi", "ori", "eori":
		return assembleLogical(n.Mnemonic, n.Operands, asm)
	case "lea", "pea":
		return assembleAddressMode(n.Mnemonic, n.Operands, asm, pc)
	case "link", "unlk":
		return assembleStack(n.Mnemonic, n.Operands, asm)
	case "cmp", "cmpa", "cmpi", "tst", "chk":
		return assembleCompare(n.Mnemonic, n.Operands, asm)
	case "abcd", "sbcd", "nbcd":
		return assembleBcd(n.Mnemonic, n.Operands)
	case "clr", "neg", "negx", "swap", "ext", "tas", "exg", "reset", "stop", "nop", "illegal":
		return assembleMisc(n.Mnemonic, n.Operands)
	case "btst", "bset", "bclr", "bchg", "lsl", "lsr", "asl", "asr", "rol", "ror":
		return assembleBitwise(n.Mnemonic, n.Operands, asm)
	case "trap", "trapv":
		return assembleTrap(n.Mnemonic, n.Operands, asm)
	case "rte", "rtr", "rts", "jmp", "jsr", "bra", "bsr", "bhi", "bls", "bcc", "bcs", "bne", "beq", "bvc", "bvs", "bpl", "bmi", "bge", "blt", "bgt", "ble":
		return assembleFlow(n.Mnemonic, n.Operands, asm.labels, pc, n.Size)
	default:
		if strings.HasPrefix(n.Mnemonic.Value, "s") {
			return assembleScc(n.Mnemonic, n.Operands)
		}
		if strings.HasPrefix(n.Mnemonic.Value, "db") {
			return assembleDbcc(n.Mnemonic, n.Operands, asm.labels, pc)
		}
		return nil, fmt.Errorf("unknown instruction: %s", n.Mnemonic.Value)
	}
}

// splitOperands splits an operand string by commas, but ignores commas inside parentheses.
func splitOperands(s string) []string {
	var result []string
	parenLevel := 0
	last := 0
	for i, r := range s {
		switch r {
		case '(':
			parenLevel++
		case ')':
			parenLevel--
		case ',':
			if parenLevel == 0 {
				result = append(result, strings.TrimSpace(s[last:i]))
				last = i + 1
			}
		}
	}
	result = append(result, strings.TrimSpace(s[last:]))
	return result
}
