package assembler

import (
	"fmt"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

const (
	// RegLabel is a placeholder register value indicating a label to be resolved.
	RegLabel = 0xFE
	// RegStatus is a placeholder register value indicating a status register (SR/CCR/USP).
	RegStatus = 0xFFFF
)

// Assembler holds the state for the assembly process.
type Assembler struct {
	symbols     map[string]int64
	labels      map[string]uint32
	outputPos   uint32
	baseAddress uint32
	opSize      int // Current operation size in bytes
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
	asm.baseAddress = baseAddress
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	nodes, err := asm.parseLines(lines)
	if err != nil {
		return nil, fmt.Errorf("parsing error: %w", err)
	}

	for pass := 0; ; pass++ {
		changed, err := asm.runSizingPass(nodes)
		if err != nil {
			return nil, fmt.Errorf("pass %d failed: %w", pass+1, err)
		}
		if !changed {
			break
		}
		if pass > 10 {
			return nil, fmt.Errorf("failed to stabilize label addresses after 10 passes")
		}
	}

	// Final Code Generation Pass
	var out []byte
	pc := baseAddress
	asm.outputPos = 0

	for _, n := range nodes {
		if n.Type == NodeLabel {
			continue
		}

		if n.Type == NodeDirective {
			// Handle directives that affect PC, emit padding, or generate raw bytes.
			dirName := strings.TrimPrefix(strings.ToLower(n.Parts[0]), ".")
			switch dirName {
			case "org":
				addr, _ := asm.parseConstant(n.Parts[1])
				pc = uint32(addr)
				asm.outputPos = pc - baseAddress
				continue // ORG emits no code itself
			case "even":
				if asm.outputPos%2 != 0 {
					out = append(out, 0x00)
					asm.outputPos++
					pc++
				}
				continue // EVEN emits at most one byte
			default:
				// For data-emitting directives, generate bytes directly.
				bytes, err := asm.generateDirectiveCode(n)
				if err != nil {
					return nil, fmt.Errorf("final generation failed for '%v': %w", n.Parts, err)
				}
				if len(bytes) > 0 {
					out = append(out, bytes...)
					asm.outputPos += uint32(len(bytes))
					pc += uint32(len(bytes))
				}
			}
		} else {
			// For instructions, generate words and convert to bytes.
			words, err := asm.generateInstructionCode(n, pc, true)
			if err != nil {
				return nil, fmt.Errorf("final generation failed for '%v': %w", n.Parts, err)
			}

			if len(words) > 0 {
				bytes := cpu.WordsToBytes(words)
				out = append(out, bytes...)
				asm.outputPos += uint32(len(bytes))
				pc += uint32(len(bytes))
			}
		}
	}

	return out, nil
}

// runSizingPass executes one sizing/label resolution pass and returns true if anything changed.
func (asm *Assembler) runSizingPass(nodes []*Node) (bool, error) {
	pc := asm.baseAddress
	changed := false

	for _, n := range nodes {
		if n.Type == NodeLabel {
			if addr, ok := asm.labels[n.Label]; !ok || addr != pc {
				asm.labels[n.Label] = pc
				changed = true
			}
			continue
		}

		oldSize := n.Size
		var size uint32

		if n.Type == NodeDirective {
			dirName := strings.TrimPrefix(strings.ToLower(n.Parts[0]), ".")
			switch dirName {
			case "org":
				addr, err := asm.parseConstant(n.Parts[1])
				if err != nil {
					return false, err
				}
				pc = uint32(addr)
				continue
			case "equ":
				continue
			}
			// For all other directives, get their size.
			dirSize, err := asm.getDirectiveSize(n, pc)
			if err != nil {
				return false, err
			}
			size = dirSize
		} else { // NodeInstruction
			// Use getSizeBra for accurate branch sizing.
			if isBranchMnemonic(n.Mnemonic.Value) {
				size = asm.getSizeBra(n, pc)
			} else {
				// For other instructions, generate to find size, assuming worst-case for errors.
				words, _ := asm.generateInstructionCode(n, pc, false)
				size = uint32(len(words) * 2)
			}
		}

		if oldSize != size {
			n.Size = size
			changed = true
		}
		pc += size
	}
	return changed, nil
}

// generateInstructionCode is the single source of truth for instruction binary generation.
func (asm *Assembler) generateInstructionCode(n *Node, pc uint32, finalPass bool) ([]uint16, error) {
	operands := make([]Operand, len(n.Operands))
	copy(operands, n.Operands)

	for i := range operands {
		op := &operands[i]
		isBareLabel := op.Mode == cpu.ModeOther && op.Register == RegLabel
		// Check if the parser explicitly identified this as PC-relative with a label
		isExplicitPCRel := op.Mode == cpu.ModeOther && op.Register == cpu.ModePCRelative && op.Label != ""

		if isBareLabel || isExplicitPCRel {
			target, ok := asm.labels[op.Label]
			if !ok {
				if finalPass {
					return nil, fmt.Errorf("undefined label: %s", op.Label)
				}
				// Sizing pass: assume worst-case (absolute long) for forward refs.
				op.Register = cpu.ModeAbsLong
				op.ExtensionWords = []uint16{0, 0}
				continue
			}

			// The M68k calculates PC-relative offsets from the address of the extension word,
			// which is always the instruction's address (pc) + 2.
			offsetPC := pc + 2
			offset := int32(target) - int32(offsetPC)

			if isBranchMnemonic(n.Mnemonic.Value) {
				// Branches are a special case. Their logic is handled entirely within
				// assembleFlow, which calculates its own offset. We don't modify the operand here.
				continue
			}

			// If the syntax was explicitly label(pc), it MUST be PC-relative.
			if isExplicitPCRel {
				if offset < -32768 || offset > 32767 {
					return nil, fmt.Errorf("pc-relative reference to '%s' is out of range", op.Label)
				}
				op.ExtensionWords = []uint16{uint16(int16(offset))}
				continue
			}

			// For bare labels, the assembler chooses the best mode.
			if canBePCRelative(n.Mnemonic) && offset >= -32768 && offset <= 32767 {
				op.Register = cpu.ModePCRelative
				op.ExtensionWords = []uint16{uint16(int16(offset))}
			} else {
				op.Register = cpu.ModeAbsLong
				op.ExtensionWords = []uint16{uint16(target >> 16), uint16(target)}
			}
		}
	}

	if len(operands) > 0 {
		for i := range operands {
			raw := strings.ToLower(strings.TrimSpace(operands[i].Raw))
			if raw == "sr" || raw == "ccr" || raw == "usp" {
				return asm.assembleStatus(n.Mnemonic, operands)
			}
		}
	}

	switch n.Mnemonic.Value {
	case "movem":
		return asm.assembleMovem(n.Mnemonic, operands)
	case "movep":
		return asm.assembleMovep(n.Mnemonic, operands)
	case "move", "movea", "moveq":
		return asm.assembleMove(n.Mnemonic, operands, pc)
	case "add", "adda", "sub", "suba", "mulu", "muls", "divu", "divs", "addx", "subx", "addq", "subq", "addi", "subi":
		return asm.assembleMath(n.Mnemonic, operands)
	case "and", "or", "eor", "not", "andi", "ori", "eori":
		return asm.assembleLogical(n.Mnemonic, operands)
	case "lea", "pea":
		return asm.assembleAddressMode(n.Mnemonic, operands, pc)
	case "link", "unlk":
		return asm.assembleStack(n.Mnemonic, operands)
	case "cmp", "cmpa", "cmpi", "tst", "chk":
		return asm.assembleCompare(n.Mnemonic, operands)
	case "abcd", "sbcd", "nbcd":
		return asm.assembleBcd(n.Mnemonic, operands)
	case "clr", "neg", "negx", "swap", "ext", "tas", "exg", "reset", "stop", "nop", "illegal":
		return asm.assembleMisc(n.Mnemonic, operands)
	case "btst", "bset", "bclr", "bchg", "lsl", "lsr", "asl", "asr", "rol", "ror":
		return asm.assembleBitwise(n.Mnemonic, operands)
	case "trap", "trapv":
		return asm.assembleTrap(n.Mnemonic, operands)
	case "rte", "rtr", "rts", "jmp", "jsr", "bra", "bsr", "bhi", "bls", "bcc", "bcs", "bne", "beq", "bvc", "bvs", "bpl", "bmi", "bge", "blt", "bgt", "ble":
		return asm.assembleFlow(n.Mnemonic, operands, asm.labels, pc, n.Size)
	default:
		if strings.HasPrefix(n.Mnemonic.Value, "s") {
			return asm.assembleScc(n.Mnemonic, operands)
		}
		if strings.HasPrefix(n.Mnemonic.Value, "db") {
			return asm.assembleDbcc(n.Mnemonic, operands, asm.labels, pc)
		}

		if !finalPass {
			// Sizing pass: assume a worst-case size for unknown instructions.
			return []uint16{0, 0, 0}, nil
		}
		return nil, fmt.Errorf("unknown instruction: %s", n.Mnemonic.Value)
	}
}

// canBePCRelative checks if an instruction's EA can be PC-relative.
func canBePCRelative(mn Mnemonic) bool {
	switch mn.Value {
	case "jmp", "jsr":
		return false
	default:
		return true
	}
}

// isBranchMnemonic checks if an instruction is a form of branch.
func isBranchMnemonic(val string) bool {
	switch val {
	case "bra", "bsr", "bhi", "bls", "bcc", "bcs", "bne", "beq", "bvc", "bvs", "bpl", "bmi", "bge", "blt", "bgt", "ble":
		return true
	default:
		return strings.HasPrefix(val, "db")
	}
}

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

		var label string
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			parsedLabel := strings.TrimSpace(parts[0])
			if !strings.ContainsAny(parsedLabel, " \t") {
				label = strings.ToLower(parsedLabel)
				nodes = append(nodes, &Node{Type: NodeLabel, Label: label, Parts: []string{label + ":"}})
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

		opFields := strings.Fields(operandStr)
		if len(opFields) > 0 && strings.EqualFold(opFields[0], "equ") {
			expr := ""
			if len(opFields) > 1 {
				expr = strings.Join(opFields[1:], " ")
			}
			val, err := asm.parseConstant(expr)
			if err != nil {
				return nil, fmt.Errorf("line %d: invalid equ value for %s: %v", i+1, mnemonic, err)
			}
			asm.symbols[strings.ToLower(mnemonic)] = val
			continue
		}

		nodeParts := []string{mnemonic}
		if operandStr != "" {
			nodeParts = append(nodeParts, operandStr)
		}

		directiveCheck := strings.ToLower(strings.TrimPrefix(mnemonic, "."))
		switch directiveCheck {
		case "dc.b", "dc.w", "dc.l", "ds.b", "ds.w", "ds.l", "org", "even":
			nodes = append(nodes, &Node{Type: NodeDirective, Parts: nodeParts})
			continue
		}

		mn, err := ParseMnemonic(mnemonic)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}

		var operands []Operand
		if operandStr != "" {
			for _, s := range splitOperands(operandStr) {
				s = strings.TrimSpace(s)
				if s == "" {
					continue
				}
				op, err := asm.parseOperand(s)
				if err != nil {
					return nil, fmt.Errorf("line %d: error parsing operand '%s': %w", i+1, s, err)
				}
				operands = append(operands, op)
			}
		}

		nodes = append(nodes, &Node{Type: NodeInstruction, Mnemonic: mn, Operands: operands, Parts: nodeParts})
	}
	return nodes, nil
}

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
