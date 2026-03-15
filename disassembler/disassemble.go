package disassembler

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

// LabelType defines the context of a label.
type LabelType int

const (
	// JumpTarget is for a simple branch (BRA, BNE, etc.).
	JumpTarget LabelType = iota
	// SubroutineEntry is for a JSR or BSR target.
	SubroutineEntry
)

// Instruction represents a single decoded instruction at a specific address.
type Instruction struct {
	Address  uint32
	Op       uint16
	Mnemonic string
	Operands string
	Size     uint32
	IsCode   bool // Flag to mark as reachable code
}

// Disassemble performs a robust, multi-stage disassembly.
func Disassemble(code []byte) (string, error) {
	if len(code) == 0 {
		return "", nil
	}

	// --- STAGE 1: Linear Sweep ---
	instructions := make(map[uint32]*Instruction)
	for pc := 0; pc+1 < len(code); {
		addr := uint32(pc)
		op := binary.BigEndian.Uint16(code[pc:])
		var extensions []byte
		if pc+2 < len(code) {
			extensions = code[pc+2:]
		}
		mn, ops, used := decode(op, 0, extensions)
		inst := &Instruction{
			Address:  addr,
			Op:       op,
			Mnemonic: mn,
			Operands: ops,
			Size:     uint32(2 + used),
		}
		instructions[addr] = inst
		pc += int(2 + used)
	}

	// --- STAGE 2: Control Flow Analysis ---
	labelTargets := make(map[uint32]LabelType)
	q := newQueue()
	q.push(0)

	// Perform control-flow analysis to find label targets (best-effort).
	for {
		addr, ok := q.pop()
		if !ok {
			break
		}

		// Heuristic: if this address points into a long printable ASCII run, it's
		// likely data. Skip marking it as code so the renderer will emit it as
		// data instead of decoding instructions out of text.
		if int(addr) < len(code) {
			printables := 0
			for i := int(addr); i < len(code) && i < int(addr)+32; i++ {
				b := code[i]
				if b >= 32 && b <= 126 {
					printables++
				} else {
					if printables >= 6 {
						break
					}
					printables = 0
				}
			}
			if printables >= 6 {
				continue
			}
		}

		inst, exists := instructions[addr]
		if !exists {
			continue
		}
		// Mark this instruction as reachable code.
		inst.IsCode = true

		if !isTerminal(inst.Mnemonic) {
			q.push(addr + inst.Size)
		}

		isSubroutineCall := inst.Mnemonic == "jsr" || inst.Mnemonic == "bsr"
		if isBranchMnemonic(inst.Mnemonic) || isSubroutineCall {
			offsetPC := inst.Address + 2
			var target int64 = -1

			if isBranchMnemonic(inst.Mnemonic) {
				offset := parseBranchOffset(inst.Operands)
				target = int64(offsetPC) + int64(offset)
			}
			if addr := parseAbsoluteAddress(inst.Operands); addr >= 0 {
				target = int64(addr)
			}

			if target >= 0 {
				targetAddr := uint32(target)
				q.push(targetAddr)
				if isSubroutineCall {
					labelTargets[targetAddr] = SubroutineEntry
				} else if _, exists := labelTargets[targetAddr]; !exists {
					labelTargets[targetAddr] = JumpTarget
				}
			}
		}
	}

	// --- STAGE 3: Render Final Output ---
	var out strings.Builder
	stringCounter := 1
	pc := uint32(0)
	totalLen := uint32(len(code))
	printedLabels := make(map[string]bool)

	for pc < totalLen {
		// If there's a label at this address, print it even if it's data.
		if labelType, exists := labelTargets[pc]; exists {
			ln := labelName(pc, labelType)
			if !printedLabels[ln] {
				fmt.Fprintf(&out, "%s:\n", ln)
				printedLabels[ln] = true
			}
		}

		// If the current address does not have a decoded instruction, or the
		// instruction isn't marked as reachable code, treat it as data.
		inst, exists := instructions[pc]
		if !exists || (exists && !inst.IsCode) {
			dataStart := pc
			dataEnd := dataStart
			for dataEnd < totalLen {
				if _, exists := instructions[dataEnd]; exists {
					break
				}
				dataEnd++
			}
			// If any labelTargets fall inside this data range, split the data region
			// so labels can be emitted at their exact addresses. This preserves
			// layout while guaranteeing forward/backward label definitions exist.
			for {
				// Find the nearest label target within [dataStart+1, dataEnd)
				var splitAt uint32 = 0
				for addr := range labelTargets {
					if addr > dataStart && addr < dataEnd {
						if splitAt == 0 || addr < splitAt {
							splitAt = addr
						}
					}
				}
				if splitAt == 0 {
					// No labels inside the remaining data region, emit it and break.
					out.WriteString(analyzeAndFormatData(code[dataStart:dataEnd], dataStart, &stringCounter))
					pc = dataEnd
					break
				}
				// Debug: report that we will split data at this label.
				// Emit data up to splitAt, then emit the label, then loop to handle
				// remaining data after the label.
				out.WriteString(analyzeAndFormatData(code[dataStart:splitAt], dataStart, &stringCounter))
				lt := labelTargets[splitAt]
				ln := labelName(splitAt, lt)
				if !printedLabels[ln] {
					fmt.Fprintf(&out, "%s:\n", ln)
					printedLabels[ln] = true
				}
				// Advance dataStart to the label address and continue splitting.
				dataStart = splitAt
			}
			continue
		}

		// Get the instruction and print it.
		finalOperands := inst.Operands
		if isBranchMnemonic(inst.Mnemonic) || inst.Mnemonic == "jsr" {
			offsetPC := inst.Address + 2
			var target int64 = -1
			if isBranchMnemonic(inst.Mnemonic) {
				offset := parseBranchOffset(inst.Operands)
				target = int64(offsetPC) + int64(offset)
			}
			if addr := parseAbsoluteAddress(inst.Operands); addr >= 0 {
				target = int64(addr)
			}
			if target >= 0 {
				labelType := JumpTarget
				targetAddr := uint32(target)
				// Use the labelTargets map (populated during control-flow analysis)
				// to determine the label type and render a label reference. This
				// ensures branches reference emitted labels even if the target is in
				// a data region.
				if lt, exists := labelTargets[targetAddr]; exists {
					labelType = lt
				} else {
					labelTargets[targetAddr] = labelType
				}
				finalOperands = labelName(targetAddr, labelType)
				// If the target is the current PC, emit the label now so the
				// assembler will see it when re-parsing this disassembly.
				if targetAddr == pc {
					ln := labelName(pc, labelType)
					if !printedLabels[ln] {
						fmt.Fprintf(&out, "%s:\n", ln)
						printedLabels[ln] = true
					}
				}
			}
		}

		// Normalize some operand forms for assembly compatibility. The decode
		// routines return canonical forms used by unit tests; however the
		// assembler parser is picky about certain EA syntaxes (notably MOVEP).
		// Apply a final normalization pass just before printing to maximize
		// round-trip success while keeping TestableDecode outputs unchanged.
		if strings.HasPrefix(inst.Mnemonic, "movep") {
			finalOperands = normalizeMovepForAssembly(finalOperands)
		}
		if finalOperands != "" {
			fmt.Fprintf(&out, "    %-8s %s\n", inst.Mnemonic, finalOperands)
		} else {
			fmt.Fprintf(&out, "    %s\n", inst.Mnemonic)
		}

		// Advance PC by the size of this single instruction.
		pc += inst.Size
	}

	// Emit any label targets that lie at or beyond the rendered range as
	// trailing label declarations so forward references do not fail during
	// re-assembly. This is conservative and does not change earlier offsets.
	for addr, lt := range labelTargets {
		if addr >= totalLen {
			ln := labelName(addr, lt)
			if !printedLabels[ln] {
				fmt.Fprintf(&out, "%s:\n", ln)
				printedLabels[ln] = true
			}
		}
	}

	// Debug: print which labels were emitted.
	if len(printedLabels) > 0 {
		fmt.Println("DEBUG: printedLabels:")
		for ln := range printedLabels {
			fmt.Printf("  %s\n", ln)
		}
	}

	return out.String(), nil
}

// isTerminal checks if an instruction unconditionally stops linear execution.
func isTerminal(mn string) bool {
	return mn == "rts" || mn == "rte" || mn == "rtr" || mn == "jmp" || mn == "bra"
}

// decode returns mnemonic, operand string, and number of extra bytes consumed.
func decode(op uint16, pc int, code []byte) (string, string, int) {
	// Handle dense 0x4E00 opcode space first with specific, ordered checks
	if (op & 0xFF00) == 0x4E00 {
		if (op&0xFFF0) == cpu.OPMOVEToUSP || (op&0xFFF0) == cpu.OPMOVEFromUSP {
			return decodeMoveSystemRegister(op, pc, code)
		}
		switch op {
		case cpu.OPNOP:
			return "nop", "", 0
		case cpu.OPRTS:
			return "rts", "", 0
		case cpu.OPRTR:
			return "rtr", "", 0
		case cpu.OPRTE:
			return "rte", "", 0
		case cpu.OPRESET:
			return "reset", "", 0
		case cpu.OPTRAPV:
			return "trapv", "", 0
		case cpu.OPSTOP:
			imm, used := readImmediateBySize(code, pc, 1)
			return "stop", imm, used
		}
		if (op & 0xFFF8) == cpu.OPLINK {
			reg := op & 7
			disp, used := readImmediateBySize(code, pc, 1)
			return "link", fmt.Sprintf("a%d,%s", reg, disp), used
		}
		if (op & 0xFFF8) == cpu.OPUNLK {
			reg := op & 7
			return "unlk", fmt.Sprintf("a%d", reg), 0
		}
		if (op & 0xFFF0) == cpu.OPTRAP {
			vec := op & 0xF
			return "trap", fmt.Sprintf("#%d", vec), 0
		}
		if (op & 0xFFC0) == cpu.OPJSR {
			return decodeJmpJsr(op, pc, code)
		}
		if (op & 0xFFC0) == cpu.OPJMP {
			return decodeJmpJsr(op, pc, code)
		}
	}

	switch op {
	case cpu.OPILLEGAL:
		return "illegal", "", 0
	case cpu.OPANDItoCCR, cpu.OPORItoCCR, cpu.OPEORItoCCR,
		cpu.OPANDItoSR, cpu.OPORItoSR, cpu.OPEORItoSR:
		return decodeImmediateToSystemRegister(op, pc, code)
	}

	if (op & 0xF138) == 0x0108 {
		return decodeMovep(op, pc, code)
	}

	if (op&0xFF00) == cpu.OPORI ||
		(op&0xFF00) == cpu.OPANDI ||
		(op&0xFF00) == cpu.OPSUBI ||
		(op&0xFF00) == cpu.OPADDI ||
		(op&0xFF00) == cpu.OPEORI ||
		(op&0xFF00) == cpu.OPCMPI {
		return decodeImmediateLogical(op, pc, code)
	}

	if (op & 0xFF00) == 0x0800 {
		return decodeBitManipulation(op, pc, code)
	}
	if (op&0xF000) == 0 && (op&0x0100) != 0 {
		return decodeBitManipulation(op, pc, code)
	}

	hi := op & 0xF000
	switch {
	case (op & 0xF0C8) == cpu.OPDBcc:
		return decodeDbcc(op, pc, code)
	case (op & 0xF0C0) == cpu.OPScc:
		return decodeScc(op, pc, code)
	case hi == cpu.OPMOVEQ:
		reg := (op >> 9) & 7
		imm8 := int8(op & 0xFF)
		return "moveq", fmt.Sprintf("#%d,d%d", imm8, reg), 0
	case (op & 0xC000) == cpu.OPMOVE:
		return decodeMoveGeneral(op, pc, code)
	case hi == cpu.OPBRA:
		return decodeBranch(op, code, pc)
	case hi == cpu.OPADDQ:
		imm := int((op >> 9) & 7)
		if imm == 0 {
			imm = 8
		}
		size := (op >> 6) & 3
		sizeStr := SizeSuffix(size)
		ea := op & 0x3F
		eaText, used := DecodeEA(ea, pc, code, size)
		if (op & 0x0100) != 0 {
			return "subq" + sizeStr, fmt.Sprintf("#%d,%s", imm, eaText), used
		}
		return "addq" + sizeStr, fmt.Sprintf("#%d,%s", imm, eaText), used
	case (op & 0xF000) == cpu.OPAND:
		if (op & 0xF100) == 0xC100 {
			opmode := (op >> 3) & 0x1F
			if opmode == 0b01001 || opmode == 0b10001 {
				return decodeExg(op)
			}
			if opmode == 0b01000 {
				regX := (op >> 9) & 7
				regY := op & 7
				if regX == regY {
					return decodeExg(op)
				}
			}
		}
		if (op&0xF0C0) == cpu.OPMULU || (op&0xF0C0) == cpu.OPMULS {
			return decodeMulDiv(op, pc, code)
		}
		return decodeLogical(op, pc, code)
	case (op & 0xF000) == cpu.OPOR:
		if (op&0xF0C0) == cpu.OPDIVU || (op&0xF0C0) == cpu.OPDIVS {
			return decodeMulDiv(op, pc, code)
		}
		return decodeLogical(op, pc, code)
	case (op & 0xF000) == 0xD000:
		return decodeAdd(op, pc, code)
	case (op & 0xF000) == 0x9000:
		return decodeSub(op, pc, code)
	case (op & 0xF000) == 0xB000:
		if (op & 0xF138) == 0xB108 {
			return decodeCmpm(op)
		}
		if (op&0x0100) == 0 && (op&0x00C0) != 0 {
			if (op & 0x01F8) == 0x0180 {
				return decodeChk(op, pc, code)
			}
		}
		return decodeCmp(op, pc, code)
	case (op & 0xFFC0) == cpu.OPMOVEFromSR,
		(op & 0xFFC0) == cpu.OPMOVEToCCR,
		(op & 0xFFC0) == cpu.OPMOVEToSR:
		return decodeMoveSystemRegister(op, pc, code)
	case (op & 0xFF00) == cpu.OPNEGX,
		(op & 0xFF00) == cpu.OPCLR,
		(op & 0xFF00) == cpu.OPNEG,
		(op & 0xFF00) == cpu.OPNOT:
		return decodeSingleOperand(op, pc, code)
	case (op & 0xFFC0) == cpu.OPTAS:
		return decodeTas(op, pc, code)
	case (op&0xFF00) == cpu.OPTST && (op&0xFFC0) != 0x4AC0:
		return decodeSingleOperand(op, pc, code)
	case (op & 0xFFC0) == cpu.OPNBCD:
		return decodeSingleOperand(op, pc, code)
	case (op&0xFFF8) == 0x4880 || (op&0xFFF8) == 0x48C0:
		return decodeSingleOperand(op, pc, code)
	case (op & 0xFFF8) == cpu.OPSWAP:
		return decodeSwap(op)
	case (op & 0xFB80) == 0x4880:
		return decodeMovem(op, pc, code)
	case (op&0xF100) == cpu.OPADDX || (op&0xF100) == cpu.OPSUBX:
		return decodeAddxSubx(op, pc, code)
	case hi == cpu.OPShiftRotateBase:
		return decodeShiftRotateGeneric(op)
	case (op & 0xFFC0) == cpu.OPPEA:
		ea := op & 0x3F
		ops, used := DecodeEA(ea, pc, code, 1)
		return "pea", ops, used
	case (op & 0xF1C0) == cpu.OPLEA:
		reg := (op >> 9) & 7
		ea := op & 0x3F
		ops, used := DecodeEA(ea, pc, code, 0)
		return "lea", fmt.Sprintf("%s,a%d", ops, reg), used
	}

	return "dc.w", fmt.Sprintf("0x%04x", op), 0
}

// NOTE: The old 'disassembleNodes' is no longer needed with this new architecture.
// The helper functions below can be moved to utility.go.

// addrQueue is a simple worklist queue for addresses to decode.
type addrQueue struct {
	items []uint32
	seen  map[uint32]bool
}

func newQueue() *addrQueue {
	return &addrQueue{seen: make(map[uint32]bool)}
}

func (q *addrQueue) push(addr uint32) {
	if addr%2 == 1 {
		addr-- // Align to word boundary
	}
	if !q.seen[addr] {
		q.items = append(q.items, addr)
		q.seen[addr] = true
	}
}

func (q *addrQueue) pop() (uint32, bool) {
	if len(q.items) == 0 {
		return 0, false
	}
	a := q.items[0]
	q.items = q.items[1:]
	return a, true
}
