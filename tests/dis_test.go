package assembler_test

import (
	"encoding/binary"
	"testing"

	"github.com/Urethramancer/m68k/assembler"
	"github.com/Urethramancer/m68k/cpu"
	"github.com/Urethramancer/m68k/disassembler"
)

// Basic sanity tests
func TestSimpleInstructions(t *testing.T) {
	tests := []struct {
		op   uint16
		want string
	}{
		{cpu.OPNOP, "nop"},
		{cpu.OPRTS, "rts"},
		{cpu.OPRESET, "reset"},
		{cpu.OPSTOP, "stop"},
		{cpu.OPRTE, "rte"},
		{cpu.OPRTR, "rtr"},
		{cpu.OPILLEGAL, "illegal"},
		{cpu.OPTRAPV, "trapv"},
	}
	for _, tt := range tests {
		mn, _, _ := disassembler.TestableDecode(tt.op, 0, nil)
		if mn != tt.want {
			t.Errorf("got %s, want %s", mn, tt.want)
		}
	}
}

// MOVEQ
func TestMoveQ(t *testing.T) {
	op := uint16(0x7010) // moveq #16,d0
	mn, ops, _ := disassembler.TestableDecode(op, 0, nil)
	if mn != "moveq" || ops != "#16,d0" {
		t.Errorf("moveq failed: got %s %s", mn, ops)
	}
}

// MOVE general
func TestMoveGeneral(t *testing.T) {
	op := uint16(0x303C) // move.w #$1234,d0
	code := make([]byte, 2)
	binary.BigEndian.PutUint16(code, 0x1234)
	mn, ops, _ := disassembler.TestableDecode(op, 0, code)
	if mn != "move.w" {
		t.Errorf("expected move.w, got %s", mn)
	}
	if ops == "" {
		t.Errorf("missing operands for move.w")
	}
}

// ADD/SUB/CMP group
func TestAddSubCmp(t *testing.T) {
	code := make([]byte, 4)

	opAdd := uint16(0xD040) // add.w d0,d0
	mn, ops, _ := disassembler.TestableDecode(opAdd, 0, code)
	if mn != "add.w" || ops != "d0,d0" {
		t.Errorf("add.w failed: got %s %s", mn, ops)
	}

	opSub := uint16(0x9441) // sub.w d1,d2
	mn, ops, _ = disassembler.TestableDecode(opSub, 0, code)
	if mn != "sub.w" || ops != "d1,d2" {
		t.Errorf("sub.w failed: got %s %s", mn, ops)
	}

	opCmp := uint16(0xB042) // cmp.w d2,d2
	mn, ops, _ = disassembler.TestableDecode(opCmp, 0, code)
	if mn != "cmp.w" {
		t.Errorf("cmp.w failed: got %s %s", mn, ops)
	}
}

// ADDQ / SUBQ
func TestAddqSubq(t *testing.T) {
	opAddq := uint16(0x5080) // addq.l #8,d0
	mn, ops, _ := disassembler.TestableDecode(opAddq, 0, nil)
	if mn != "addq.l" || ops != "#8,d0" {
		t.Errorf("addq failed: got '%s %s', want 'addq.l #8,d0'", mn, ops)
	}

	opSubq := uint16(0x5183) // subq.l #8,d3
	mn, ops, _ = disassembler.TestableDecode(opSubq, 0, nil)
	if mn != "subq.l" || ops != "#8,d3" {
		t.Errorf("subq failed: got '%s %s', want 'subq.l #8,d3'", mn, ops)
	}
}

func TestMovem(t *testing.T) {
	// Opcode for: movem.l <reglist>,-(a7)
	op := uint16(0x48E7)
	// Register mask for d0-d5, which is 0x003F
	code := []byte{0x00, 0x3F}

	mn, ops, used := disassembler.TestableDecode(op, 0, code)

	wantMnemonic := "movem.l"
	wantOps := "d0-d5,-(a7)"
	wantUsed := 2

	if mn != wantMnemonic {
		t.Errorf("movem mnemonic failed: got %s, want %s", mn, wantMnemonic)
	}
	if ops != wantOps {
		t.Errorf("movem operands incorrect: got '%s', want '%s'", ops, wantOps)
	}
	if used != wantUsed {
		t.Errorf("movem consumed bytes incorrect: got %d, want %d", used, wantUsed)
	}
}

// LEA / PEA / LINK / UNLK
func TestLeaPeaLinkUnlk(t *testing.T) {
	opLea := uint16(0x41FA) // lea (d16,pc),a0
	code := []byte{0x00, 0x10}
	mn, ops, _ := disassembler.TestableDecode(opLea, 0, code)
	if mn != "lea" {
		t.Errorf("lea failed: %s", mn)
	}

	opPea := uint16(0x4850) // pea (a0)
	mn, ops, _ = disassembler.TestableDecode(opPea, 0, nil)
	if mn != "pea" {
		t.Errorf("pea failed: got %s %s", mn, ops)
	}

	opLink := uint16(0x4E50) // link a0,#-4
	code = []byte{0xFF, 0xFC}
	mn, ops, _ = disassembler.TestableDecode(opLink, 0, code)
	if mn != "link" {
		t.Errorf("link failed: got %s", mn)
	}

	opUnlk := uint16(0x4E58) // unlk a0
	mn, ops, _ = disassembler.TestableDecode(opUnlk, 0, nil)
	if mn != "unlk" {
		t.Errorf("unlk failed: got %s", mn)
	}
}

// Immediate logicals
func TestImmediateLogicals(t *testing.T) {
	code := []byte{0x00, 0xFF}
	tests := []struct {
		op   uint16
		want string
	}{
		{0x0240, "andi.w"},
		{0x0040, "ori.w"},
		{0x0A40, "eori.w"},
		{0x0640, "addi.w"},
		{0x0440, "subi.w"},
		{0x0C40, "cmpi.w"},
	}
	for _, tt := range tests {
		mn, _, _ := disassembler.TestableDecode(tt.op, 0, code)
		if mn != tt.want {
			t.Errorf("expected %s, got %s", tt.want, mn)
		}
	}
}

// Logical (AND, OR, EOR)
func TestLogicalRegister(t *testing.T) {
	tests := []struct {
		op   uint16
		want string
		ops  string
	}{
		// AND
		{0xC141, "and.w", "d0,d1"},   // Dn -> Dn
		{0xC041, "and.w", "d1,d0"},   // Dn <- Dn
		{0xC150, "and.w", "d0,(a0)"}, // Dn -> (An)
		{0xC050, "and.w", "(a0),d0"}, // (An) -> Dn
		// OR
		{0x8543, "or.w", "d2,d3"},   // Dn -> Dn
		{0x8450, "or.w", "(a0),d2"}, // (An) -> Dn
		// EOR
		{0xB945, "eor.w", "d4,d5"},    // Dn -> Dn
		{0xB959, "eor.w", "d4,(a1)+"}, // Dn -> (An)+
	}

	for _, tt := range tests {
		mn, ops, _ := disassembler.TestableDecode(tt.op, 0, nil)
		if mn != tt.want || ops != tt.ops {
			t.Errorf("op 0x%04X: got '%s %s', want '%s %s'", tt.op, mn, ops, tt.want, tt.ops)
		}
	}
}

// Single Operand (CLR, NEG, NOT, TST, etc.)
func TestSingleOperandInstructions(t *testing.T) {
	tests := []struct {
		op   uint16
		want string
		ops  string
	}{
		// CLR
		{0x4200, "clr.b", "d0"},
		{0x4241, "clr.w", "d1"},
		{0x4282, "clr.l", "d2"},
		{0x4250, "clr.b", "(a0)"},
		// NEG
		{0x4410, "neg.b", "(a0)"},
		{0x4441, "neg.w", "d1"},
		// NEGX
		{0x4010, "negx.b", "(a0)"},
		{0x4042, "negx.w", "d2"},
		// NOT
		{0x4603, "not.b", "d3"},
		{0x4651, "not.w", "(a1)+"},
		// NBCD
		{0x4804, "nbcd", "d4"},
		// TST
		{0x4A05, "tst.b", "d5"},
		{0x4A46, "tst.w", "d6"},
		{0x4A9F, "tst.l", "(a7)+"},
		// SWAP
		{0x4840, "swap", "d0"},
		{0x4847, "swap", "d7"},
	}

	for _, tt := range tests {
		mn, ops, _ := disassembler.TestableDecode(tt.op, 0, nil)
		if mn != tt.want || ops != tt.ops {
			t.Errorf("op 0x%04X: got '%s %s', want '%s %s'", tt.op, mn, ops, tt.want, tt.ops)
		}
	}
}

// Branch family
func TestBranches(t *testing.T) {
	branches := []struct {
		op   uint16
		want string
	}{
		{0x6002, "bra"},
		{0x6104, "bsr"},
		{0x6606, "bne"},
		{0x6708, "beq"},
		{0x6C0A, "bge"},
		{0x6D0C, "blt"},
		{0x6E0E, "bgt"},
		{0x6F10, "ble"},
	}
	for _, b := range branches {
		mn, _, _ := disassembler.TestableDecode(b.op, 0, nil)
		if mn != b.want {
			t.Errorf("expected %s, got %s", b.want, mn)
		}
	}
}

// Shift / Rotate
func TestShiftRotate(t *testing.T) {
	tests := []struct {
		op   uint16
		want string
	}{
		{0xE048, "lsr.w"},
		{0xE058, "ror.w"},
		{0xE148, "lsl.w"},
		{0xE158, "rol.w"},
	}
	for _, tt := range tests {
		mn, _, _ := disassembler.TestableDecode(tt.op, 0, nil)
		if mn != tt.want {
			t.Errorf("expected %s, got %s", tt.want, mn)
		}
	}
}

// Bit manipulation instructions
func TestBitManipulation(t *testing.T) {
	tests := []struct {
		op   uint16
		code []byte
		want string
		ops  string
	}{
		// Dynamic (register) form
		{0x0101, nil, "btst.l", "d0,d1"},
		{0x0542, nil, "bchg.l", "d2,d2"},
		{0x0783, nil, "bclr.l", "d3,d3"},
		{0x09C4, nil, "bset.l", "d4,d4"},
		{0x0F5F, nil, "bchg.b", "d7,(a7)+"},
		// Static (immediate) form - no size suffix
		{0x0801, []byte{0x00, 0x0F}, "btst", "#15,d1"},
		{0x0842, []byte{0x00, 0x10}, "bchg", "#16,d2"},
		{0x089F, []byte{0x00, 0x01}, "bclr", "#1,(a7)+"},
		{0x08C4, []byte{0x00, 0x02}, "bset", "#2,d4"},
	}

	for _, tt := range tests {
		mn, ops, _ := disassembler.TestableDecode(tt.op, 0, tt.code)
		if mn != tt.want || ops != tt.ops {
			t.Errorf("op 0x%04X: got '%s %s', want '%s %s'", tt.op, mn, ops, tt.want, tt.ops)
		}
	}
}

// TestJmpJsrSccDbcc tests JMP and JSR instructions.
func TestJmpJsr(t *testing.T) {
	tests := []struct {
		op   uint16
		code []byte
		want string
	}{
		// JMP
		{0x4ED0, nil, "jmp (a0)"},
		{0x4EF9, []byte{0x00, 0x00, 0x12, 0x34}, "jmp $1234.l"},

		// JSR
		{0x4E91, nil, "jsr (a1)"},
	}

	for _, tt := range tests {
		mn, ops, _ := disassembler.TestableDecode(tt.op, 0, tt.code)
		full := mn
		if ops != "" {
			full += " " + ops
		}
		if full != tt.want {
			t.Errorf("op 0x%04X: got '%s', want '%s'", tt.op, full, tt.want)
		}
	}
}

// TestSccDbcc tests Scc and DBcc instructions.
func TestSccDbcc(t *testing.T) {
	code := []byte{0x00, 0x0A} // displacement for DBcc
	tests := []struct {
		op   uint16
		want string
	}{
		// Scc
		{0x50C0, "st d0"}, // Scc T
		{0x51C1, "sf d1"}, // Scc F
		{0x54E0, "scc -(a0)"},

		// DBcc
		{0x51C8, "dbf d0,+10"},
		{0x54C9, "dbcc d1,+10"},
	}

	for _, tt := range tests {
		mn, ops, _ := disassembler.TestableDecode(tt.op, 0, code)
		full := mn
		if ops != "" {
			full += " " + ops
		}
		if full != tt.want {
			t.Errorf("op 0x%04X: got '%s', want '%s'", tt.op, full, tt.want)
		}
	}
}

// TestCmpm tests the CMPM instruction.
func TestCmpm(t *testing.T) {
	tests := []struct {
		op   uint16
		want string
		ops  string
	}{
		{0xB308, "cmpm.b", "(a0)+,(a1)+"}, // was 0xB108
		{0xB748, "cmpm.w", "(a0)+,(a3)+"}, // was 0xB348
		{0xBB88, "cmpm.l", "(a0)+,(a5)+"}, // was 0xB588
	}

	for _, tt := range tests {
		mn, ops, _ := disassembler.TestableDecode(tt.op, 0, nil)
		if mn != tt.want || ops != tt.ops {
			t.Errorf("op 0x%04x: got '%s %s', want '%s %s'", tt.op, mn, ops, tt.want, tt.ops)
		}
	}
}

// TestExtExg tests EXT and EXG instructions.
func TestExtExg(t *testing.T) {
	tests := []struct {
		op   uint16
		want string
		ops  string
	}{
		// EXT
		{0x4880, "ext.w", "d0"},
		{0x4881, "ext.w", "d1"},
		{0x48C2, "ext.l", "d2"},
		{0x48C3, "ext.l", "d3"},
		// EXG
		{0xC140, "exg", "d0,d0"},
		{0xC542, "exg", "d2,d2"},
		{0xC148, "exg", "a0,a0"},
		{0xC789, "exg", "d3,a1"},
		{0xCB8A, "exg", "d5,a2"},
	}

	for _, tt := range tests {
		mn, ops, _ := disassembler.TestableDecode(tt.op, 0, nil)
		if mn != tt.want || ops != tt.ops {
			t.Errorf("op 0x%04X: got '%s %s', want '%s %s'", tt.op, mn, ops, tt.want, tt.ops)
		}
	}
}

// TestMovep tests the MOVEP instruction.
func TestMovep(t *testing.T) {
	tests := []struct {
		op   uint16
		code []byte
		want string
		ops  string
	}{
		// Register to memory, word: 0000 000 110 001 001 -> 0x0189
		{0x0189, []byte{0x12, 0x34}, "movep.w", "d0,(4660,a1)"},
		// Register to memory, long: 0000 000 111 001 010 -> 0x01CA
		{0x01CA, []byte{0xFF, 0xFE}, "movep.l", "d0,(-2,a2)"},
		// Memory to register, word: 0000 000 100 001 011 -> 0x010B
		{0x010B, []byte{0x00, 0x00}, "movep.w", "(0,a3),d0"},
		// Memory to register, long: 0000 010 101 001 101 -> 0x054D
		{0x054D, []byte{0x80, 0x00}, "movep.l", "(-32768,a5),d2"},
	}

	for _, tt := range tests {
		mn, ops, _ := disassembler.TestableDecode(tt.op, 0, tt.code)
		if mn != tt.want || ops != tt.ops {
			t.Errorf("op 0x%04X: got '%s %s', want '%s %s'", tt.op, mn, ops, tt.want, tt.ops)
		}
	}
}

// TestSystemInstructions tests system-level instructions like TAS.
func TestSystemInstructions(t *testing.T) {
	asm := assembler.New()
	tests := []string{
		// TAS
		"tas d0",
		"tas (a1)",
		"tas (a2)+",
		"tas -(a3)",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			code, err := asm.Assemble(tt, 0)
			if err != nil {
				t.Fatalf("Failed to assemble '%s': %v", tt, err)
			}

			op := binary.BigEndian.Uint16(code)
			var ext []byte
			if len(code) > 2 {
				ext = code[2:]
			}

			mn, ops, _ := disassembler.TestableDecode(op, 0, ext)
			result := mn
			if ops != "" {
				result += " " + ops
			}

			if result != tt {
				t.Errorf("got '%s', want '%s'", result, tt)
			}
		})
	}
}

// TestPrivilegedImmediate tests immediate operations on CCR and SR.
func TestPrivilegedImmediate(t *testing.T) {
	asm := assembler.New()
	tests := []string{
		"andi #16,ccr",
		"ori #4,ccr",
		"eori #1,ccr",
		"andi #$700,sr",
		"ori #$2000,sr",
		"eori #$8000,sr",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			code, err := asm.Assemble(tt, 0)
			if err != nil {
				t.Fatalf("Failed to assemble '%s': %v", tt, err)
			}

			op := binary.BigEndian.Uint16(code)
			var ext []byte
			if len(code) > 2 {
				ext = code[2:]
			}

			mn, ops, _ := disassembler.TestableDecode(op, 0, ext)
			result := mn
			if ops != "" {
				result += " " + ops
			}

			if result != tt {
				t.Errorf("got '%s', want '%s'", result, tt)
			}
		})
	}
}

// TestMoveSystemRegisters tests MOVE to/from SR, CCR, and USP.
func TestMoveSystemRegisters(t *testing.T) {
	asm := assembler.New()
	tests := []string{
		"move sr,d0",
		"move d1,ccr",
		"move d2,sr",
		"move.l usp,a3",
		"move.l a4,usp",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			code, err := asm.Assemble(tt, 0)
			if err != nil {
				t.Fatalf("Failed to assemble '%s': %v", tt, err)
			}

			op := binary.BigEndian.Uint16(code)
			var ext []byte
			if len(code) > 2 {
				ext = code[2:]
			}

			mn, ops, _ := disassembler.TestableDecode(op, 0, ext)
			result := mn
			if ops != "" {
				result += " " + ops
			}

			if result != tt {
				t.Errorf("got '%s', want '%s'", result, tt)
				t.Logf("%s = %04x", tt, binary.BigEndian.Uint16(code))
			}
		})
	}
}
