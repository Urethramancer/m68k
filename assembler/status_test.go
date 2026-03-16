package assembler

import (
	"testing"

	"github.com/Urethramancer/m68k/cpu"
)

func TestMoveToFromSrCcrUsp(t *testing.T) {
	asm := newAssembler()
	// MOVE D0,SR
	op := Operand{Mode: cpu.ModeData, Register: 0}
	_, err := asm.assembleMoveToSr(op)
	if err != nil {
		t.Fatalf("MOVE to SR failed: %v", err)
	}

	// MOVE SR,D1
	op2 := Operand{Mode: cpu.ModeData, Register: 1}
	_, err = asm.assembleMoveFromSr(op2)
	if err != nil {
		t.Fatalf("MOVE from SR failed: %v", err)
	}

	// ANDI #$2700, SR
	src := Operand{Raw: "#$2700", Mode: cpu.ModeOther, Register: cpu.RegImmediate, ExtensionWords: []uint16{0x2700}}
	_, err = assembleLogicImmediateToSr(cpu.OPANDItoSR, src, "ANDI")
	if err != nil {
		t.Fatalf("ANDI to SR failed: %v", err)
	}

	// MOVE A3,USP
	reg := Operand{Mode: cpu.ModeAddr, Register: 3}
	_, err = assembleMoveToUsp(reg)
	if err != nil {
		t.Fatalf("MOVE to USP failed: %v", err)
	}

	// MOVE USP,A4
	reg2 := Operand{Mode: cpu.ModeAddr, Register: 4}
	_, err = assembleMoveFromUsp(reg2)
	if err != nil {
		t.Fatalf("MOVE from USP failed: %v", err)
	}
}
