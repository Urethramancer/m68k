package assembler_test

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/Urethramancer/m68k/assembler"
)

// Assembles source and checks against an expected byte sequence (in hex).
// Automatically validates output length and content.
func assembleAndMatchHex(t *testing.T, name, src, expectedHex string) {
	t.Helper()

	expectedHex = strings.ToLower(strings.Join(strings.Fields(expectedHex), ""))
	expected, err := hex.DecodeString(expectedHex)
	if err != nil {
		t.Fatalf("[%s] invalid expected hex string: %v", name, err)
	}

	asm := assembler.New()
	code, err := asm.Assemble(src, 0x1000)
	if err != nil {
		t.Fatalf("[%s] failed to assemble:\n%s\nerror: %v", name, src, err)
	}
	if len(code) != len(expected) {
		t.Fatalf("[%s] expected %d bytes, got %d\nexpected: % X\ngot:      % X",
			name, len(expected), len(code), expected, code)
	}
	for i := range code {
		if code[i] != expected[i] {
			t.Errorf("[%s] mismatch at byte %d\nexpected: % X\ngot:      % X",
				name, i, expected, code)
			break
		}
	}
}

// Core instruction encodings
func TestBasicEncodings(t *testing.T) {
	tests := []struct {
		name, src, hex string
	}{
		{"MOVE_B_D0_D1", "move.b d0,d1", "12 00"},
		{"MOVE_L_Immediate", "move.l #$12345678,d3", "26 3C 12 34 56 78"},
		{"MOVEQ", "moveq #1,d7", "7E 01"},
		{"LEA_PCDisp", "lea ($10,pc),a0", "41 FA 00 10"},
		{"RTS", "rts", "4E 75"},
		{"NOP", "nop", "4E 71"},
		{"STOP", "stop #$2700", "4E 72 27 00"},
		{"TRAP1", "trap #1", "4E 41"},
	}
	for _, tc := range tests {
		assembleAndMatchHex(t, tc.name, tc.src, tc.hex)
	}
}

func TestDirectives_Encodings(t *testing.T) {
	tests := []struct {
		name, src, hex string
	}{
		// DC.B — byte order preserved, padded to even length
		{"DCB", "dc.b $11,$22,$33", "11 22 33 00"},
		// DC.W — each word stored big-endian
		{"DCW", "dc.w $1122,$3344", "11 22 33 44"},
		// DC.L — each longword stored big-endian
		{"DCL", "dc.l $11223344,$55667788", "11 22 33 44 55 66 77 88"},
		// Strings are written naturally (ASCII order)
		{"DCB_String", "dc.b 'ABCD',$00", "41 42 43 44 00 00"},
		// EVEN — pads correctly, preserving big-endian byte order
		{"EVEN", "dc.b $11\n.even\ndc.b $22", "11 00 22 00"},
		// MixedDCB — ASCII + bytes in correct order
		{"MixedDCB", "dc.b 'A',$42,'B','C',$00", "41 42 42 43 00 00"},
		// DC.W mixed numeric — big-endian words
		{"MixedDCW", "dc.w $0001,$0203,$0405", "00 01 02 03 04 05"},
		// DC.L single long, padded to even — big-endian
		{"OddDCL", "dc.l $01020304", "01 02 03 04"},
		// DS directives — filled with zeros
		{"DSB", "ds.b 4", "00 00 00 00"},
		{"DSW", "ds.w 2", "00 00 00 00"},
		{"DSL", "ds.l 1", "00 00 00 00"},
	}
	for _, tc := range tests {
		assembleAndMatchHex(t, tc.name, tc.src, tc.hex)
	}
}

func TestOrgAndEqu(t *testing.T) {
	tests := []struct {
		name, src, hex string
	}{
		// ORG simply relocates, producing only the emitted instruction
		{"ORG_Skip", "org $2000\nnop", "4E 71"},
		// EQU defines constant used in data directive — stored big-endian
		{"EQU_Usage", "value equ $1234\ndc.w value", "12 34"},
	}
	for _, tc := range tests {
		assembleAndMatchHex(t, tc.name, tc.src, tc.hex)
	}
}

// Addressing Modes
func TestAddressingModes_Encodings(t *testing.T) {
	tests := []struct {
		name, src, hex string
	}{
		{"Addr_Indirect", "move.w (a0),d0", "30 10"},
		{"Addr_PostInc", "move.w (a0)+,d1", "32 18"},
		{"Addr_PreDec", "move.w -(a0),d2", "34 20"},
		{"Addr_Disp", "move.w 4(a0),d3", "36 28 00 04"},
		{"Addr_Index", "move.w 8(a0,d1.w),d4", "38 30 10 08"},
		{"PC_Relative", "move.w label(pc),d5\nlabel: dc.w $1234", "3A 3A 00 02 12 34"},
		{"Immediate", "move.w #$ABCD,d6", "3C 3C AB CD"},
		{"Absolute_Short", "move.w ($1234).w,d7", "3E 38 12 34"},
		{"Absolute_Long", "move.l ($123456).l,d0", "20 39 00 12 34 56"},
	}
	for _, tc := range tests {
		assembleAndMatchHex(t, tc.name, tc.src, tc.hex)
	}
}

// Label resolution and PC-relative
func TestLabelResolution(t *testing.T) {
	src := `
start:
    bra.s forward
back:
    dc.w $BEEF
forward:
    dc.w $CAFE
    lea back(pc),a0
`
	assembleAndMatchHex(t, "LabelResolution", src,
		"60 02 BE EF CA FE 41 FA FF FA")
}

// Control flow and branches
func TestFlowControl_Encodings(t *testing.T) {
	tests := []struct {
		name, src, hex string
	}{
		{"BRA_Short", "bra.s label\nlabel: nop", "60 00 4E 71"},
		{"BNE_Short", "bne.s label\nlabel: nop", "66 00 4E 71"},
		{"BEQ_Short", "beq.s label\nlabel: nop", "67 00 4E 71"},
		{"BSR_Short", "bsr.s label\nlabel: nop", "61 00 4E 71"},
		{"JSR_AbsLong", "jsr $E.l", "4E B9 00 00 00 0E"},
		{"JMP_Indirect", "jmp (a0)", "4E D0"},
		{"RTE", "rte", "4E 73"},
		{"RTR", "rtr", "4E 77"},
	}
	for _, tc := range tests {
		assembleAndMatchHex(t, tc.name, tc.src, tc.hex)
	}
}

// TestCombinedCodeAndData checks a realistic mixed code and data scenario.
func TestCombinedCodeAndData(t *testing.T) {
	src := `
start:
    lea string,a0
    moveq #13,d0
    jsr somewhere
    rts
somewhere:
    nop
    rts
string:
    dc.b 'This is a test string.',$00
    dc.b $00,$de,$ad,$be,$ef
string2:
    dc.b 'VER1',$00
    dc.b $00
    dc.b $41,$42,$43
    dc.b $00
string3:
    dc.b 'Copyright (C) 2025',$00
    dc.b $00
`
	expected := `
41 FA 00 10 70 0D 4E B9 00 00 10 0E 4E 75 4E 71
4E 75 54 68 69 73 20 69 73 20 61 20 74 65 73 74
20 73 74 72 69 6E 67 2E 00 00 00 DE AD BE EF 00
56 45 52 31 00 00 00 00 41 42 43 00 00 00 43
6F 70 79 72 69 67 68 74 20 28 43 29 20 32 30 32
35 00 00 00 00
`

	assembleAndMatchHex(t, "CombinedCodeAndData", src, expected)
}
