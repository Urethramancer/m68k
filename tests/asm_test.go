package assembler_test

import (
	"testing"

	"github.com/Urethramancer/m68k/assembler"
)

// assembleAndCheck assembles a source snippet and ensures output exists.
// If expectedLen > 0, it also checks that the code length matches exactly.
func assembleAndCheck(t *testing.T, name, src string, expectedLen int) {
	t.Helper()
	asm := assembler.New()
	code, err := asm.Assemble(src, 0x1000)
	if err != nil {
		t.Errorf("[%s] failed to assemble:\n%s\nerror: %v", name, src, err)
		return
	}
	if len(code) == 0 {
		t.Errorf("[%s] assembler produced no code for:\n%s", name, src)
		return
	}
	if expectedLen > 0 && len(code) != expectedLen {
		t.Errorf("[%s] expected %d bytes, got %d\nsource:\n%s", name, expectedLen, len(code), src)
	}
}

// --- General instruction coverage ---
func TestMC68000Instructions(t *testing.T) {
	coreTests := []struct {
		src         string
		expectedLen int
	}{
		// ----- Data Movement -----
		{"move.b d0,d1", 2},
		{"move.w (a0),d2", 2},
		{"move.l #$12345678,d3", 6},
		{"movea.l d4,a0", 2},
		{"moveq #1,d7", 2},
		{"movem.l d0-d3/a0-a2,-(sp)", 4},
		{"movem.w (sp)+,d0-d3/a0-a2", 4},
		{"movep.w d1,4(a0)", 4},
		{"movep.l 8(a1),d2", 4},
		// includes a label + nop (2 bytes extra)
		{"lea label,a0\nlabel: nop", 6},
		{"pea (a0)", 2},

		// Arithmetic
		{"add.b d1,d0", 2},
		{"add.w (a0),d0", 2},
		{"add.l #4,d2", 6},
		{"adda.w #2,a1", 4},
		{"addi.w #5,d2", 4},
		{"addq.w #1,d3", 2},
		{"addx.b d1,d0", 2},
		{"sub.b d1,d0", 2},
		{"subq.l #1,d2", 2},
		{"subi.w #3,d3", 4},
		{"subx.w d4,d5", 2},
		{"mulu.w d2,d3", 2},
		{"muls.w d4,d5", 2},
		{"divu.w d6,d7", 2},
		{"divs.w d7,d6", 2},

		// Logical
		{"and.b d1,d0", 2},
		{"andi.b #$FF,d0", 4},
		{"or.w d2,d1", 2},
		{"ori.w #$AAAA,d0", 4},
		{"eor.l d3,d2", 2},
		{"eori.b #1,d0", 4},
		{"not.b d0", 2},
		{"btst #3,d0", 4},
		{"bset #2,d1", 4},
		{"bclr #1,d2", 4},
		{"bchg #0,d3", 4},

		// Compare / Test
		{"cmp.w d1,d2", 2},
		{"cmpa.l a1,a0", 2},
		{"chk.w d1,d2", 2},
		{"tst.b d3", 2},

		// Shifts / Rotates
		{"lsl.w #1,d0", 2},
		{"lsr.l #8,d1", 2},
		{"asl.w #4,d2", 2},
		{"asr.w #2,d3", 2},
		{"rol.b #1,d4", 2},
		{"ror.b #1,d5", 2},

		// BCD Arithmetic
		{"abcd d1,d0", 2},
		{"sbcd d2,d3", 2},
		{"nbcd d4", 2},

		// Miscellaneous
		{"clr.w d0", 2},
		{"neg.l d1", 2},
		{"negx.b d2", 2},
		{"swap d3", 2},
		{"ext.w d4", 2},
		{"ext.l d5", 2},
		{"tas d6", 2},
		{"exg d1,d2", 2},
		{"reset", 2},
		{"stop #$2700", 4},
		{"nop", 2},
		{"illegal", 2},

		// Stack / Link
		{"link a6,#-4", 4},
		{"unlk a6", 2},

		// Status Register
		{"move #$2700,sr", 4},
		{"move sr,d0", 2},

		// Control Flow
		{"jmp (a0)", 2},
		{"jsr (a0)", 2},
		{"rts", 2},
		{"rtr", 2},
		{"rte", 2},
		{"trap #1", 2},
		{"trapv", 2},

		// Condition Codes
		{"seq d0", 2},
		{"sne d1", 2},
		{"slt d2", 2},
		{"sgt d3", 2},
	}

	// Branches (with label appended)
	branches := []struct {
		src         string
		expectedLen int
	}{
		{"bra.s label", 2},
		{"bne.s label", 2},
		{"beq.s label", 2},
		{"bcc.s label", 2},
		{"bcs.s label", 2},
		{"bpl.s label", 2},
		{"bmi.s label", 2},
		{"bvc.s label", 2},
		{"bvs.s label", 2},
		{"bge.s label", 2},
		{"blt.s label", 2},
		{"bgt.s label", 2},
		{"ble.s label", 2},
		{"bsr.s label", 2},
		{"dbra d0,label", 4},
		{"dbcc d1,label", 4},
	}

	t.Run("Branches", func(t *testing.T) {
		for _, b := range branches {
			src := b.src + "\nlabel:"
			assembleAndCheck(t, b.src, src, b.expectedLen)
		}
	})

	t.Run("Instructions", func(t *testing.T) {
		for _, tc := range coreTests {
			assembleAndCheck(t, tc.src, tc.src, tc.expectedLen)
		}
	})
}

// --- Addressing Modes ---
func TestAddressingModes_All68000Modes(t *testing.T) {
	tests := []struct {
		name        string
		src         string
		expectedLen int
	}{
		{"DataRegDirect", "move.b d1,d0", 2},
		{"AddrRegDirect", "movea.l a1,a0", 2},
		{"AddrInd", "move.w (a0),d0", 2},
		{"PostInc", "move.w (a0)+,d1", 2},
		{"PreDec", "move.w -(a0),d2", 2},
		{"AddrDisp", "move.b 4(a0),d3", 4},
		{"AddrIndex", "move.w 2(a0,d1.w),d4", 4},
		{"AbsShort", "move.w $4000,d5", 4},
		{"AbsLong", "move.l $123456,d6", 6},
		// includes label + .dc.w (2 bytes extra)
		{"PCDisp", "move.w pcrel(pc),d7\npcrel: .dc.w $abcd", 6},
		{"PCIndex", "move.b 4(pc,d1.w),d0", 4},
		{"ImmediateLong", "move.l #$deadbeef,d1", 6},
		{"MoveQ", "moveq #5,d2", 2},
		// includes label + .dc.l (4 bytes extra)
		{"LEA", "lea mydata,a2\nmydata: .dc.l $11223344", 8},
		{"PEA", "pea mydata(pc)\nmydata: .dc.l $11223344", 8},
		{"MOVEM_Store", "movem.l d0-d3/a0-a2,-(sp)", 4},
		{"MOVEM_Load", "movem.l (sp)+,d0-d3/a0-a2", 4},
		{"MOVEP_RegToMem", "movep.w d1,4(a0)", 4},
		{"MOVEP_MemToReg", "movep.l 8(a1),d2", 4},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assembleAndCheck(t, tc.name, tc.src, tc.expectedLen)
		})
	}
}
