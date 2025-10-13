package disassembler

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
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

	// Linear Sweep
	// Disassemble every word as a potential instruction.
	instructions := make(map[uint32]*Instruction)
	for pc := 0; pc+1 < len(code); {
		addr := uint32(pc)
		op := binary.BigEndian.Uint16(code[pc:])

		// Create a temporary slice for the decoder to read extension words from.
		// pc+2 ensures it doesn't read its own opcode.
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

		// Move to the next word boundary.
		pc += 2
	}

	// Control Flow Analysis
	// Trace execution from the entry point to find all reachable code.
	q := newQueue()
	q.push(0) // Start at address 0

	for {
		addr, ok := q.pop()
		if !ok {
			break
		}

		inst, exists := instructions[addr]
		if !exists || inst.IsCode {
			continue // Already processed or not a valid instruction boundary
		}
		inst.IsCode = true

		// Add the next instruction in the linear flow to the queue,
		// unless this instruction terminates the flow.
		if !isTerminal(inst.Mnemonic) {
			q.push(addr + inst.Size)
		}

		// If it's a branch or jump, add the target address to the queue.
		if isBranchMnemonic(inst.Mnemonic) || inst.Mnemonic == "jsr" || inst.Mnemonic == "jmp" {
			offsetPC := inst.Address + 2

			// Try to parse branch offset
			if isBranchMnemonic(inst.Mnemonic) {
				offset := parseBranchOffset(inst.Operands)
				target := int64(offsetPC) + int64(offset)
				if target >= 0 {
					q.push(uint32(target))
				}
			}

			// Try to parse absolute JMP/JSR address
			if addr := parseAbsoluteAddress(inst.Operands); addr >= 0 {
				q.push(uint32(addr))
			}
		}
	}

	// Render final output
	var out strings.Builder
	stringCounter := 1
	pc := uint32(0)
	totalLen := uint32(len(code))

	// Build a map of all code addresses for easy lookup.
	codeAddrs := make(map[uint32]bool)
	for _, inst := range instructions {
		if inst.IsCode {
			codeAddrs[inst.Address] = true
		}
	}

	for pc < totalLen {
		// If the current PC is not the start of a valid code instruction,
		// find the end of this data block and format it.
		if !codeAddrs[pc] {
			dataStart := pc
			dataEnd := dataStart
			for dataEnd < totalLen && !codeAddrs[dataEnd] {
				dataEnd++
			}
			out.WriteString(analyzeAndFormatData(code[dataStart:dataEnd], dataStart, &stringCounter))
			pc = dataEnd
			continue
		}

		// This is a valid code instruction.
		inst := instructions[pc]

		// In this simplified model, labels are generated on the fly.
		// A more advanced version would build a label map here.
		out.WriteString(fmt.Sprintf("loc_%04X:\n", inst.Address))

		// Print this instruction and all subsequent contiguous instructions.
		for pc < totalLen && codeAddrs[pc] {
			inst = instructions[pc]
			if inst.Operands != "" {
				fmt.Fprintf(&out, "    %-8s %s\n", inst.Mnemonic, inst.Operands)
			} else {
				fmt.Fprintf(&out, "    %s\n", inst.Mnemonic)
			}
			pc += inst.Size
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
