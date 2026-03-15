package assembler

import (
	"testing"

	"github.com/Urethramancer/m68k/cpu"
)

func TestAbcd_RegisterForm(t *testing.T) {
	asm := New()
	op1 := Operand{Mode: cpu.ModeData, Register: 1}
	op2 := Operand{Mode: cpu.ModeData, Register: 2}
	words, err := asm.assembleAbcdSbcd(true, []Operand{op1, op2})
	if err != nil {
		t.Fatalf("ABCD register form failed: %v", err)
	}
	if len(words) != 1 {
		t.Fatalf("expected one word for ABCD register form, got %d", len(words))
	}
}

func TestAbcd_PreDecForm(t *testing.T) {
	asm := New()
	op1 := Operand{Mode: cpu.ModeAddrPreDec, Register: 3}
	op2 := Operand{Mode: cpu.ModeAddrPreDec, Register: 4}
	words, err := asm.assembleAbcdSbcd(true, []Operand{op1, op2})
	if err != nil {
		t.Fatalf("ABCD predec form failed: %v", err)
	}
	if len(words) != 1 {
		t.Fatalf("expected one word for ABCD predec form, got %d", len(words))
	}
}

func TestNbcd_MemoryAndDataReg(t *testing.T) {
	asm := New()
	// memory operand
	op := Operand{Mode: cpu.ModeAddrInd, Register: 0}
	_, err := asm.assembleNbcd([]Operand{op})
	if err != nil {
		t.Fatalf("NBCD memory form failed: %v", err)
	}
	// data register
	op = Operand{Mode: cpu.ModeData, Register: 2}
	_, err = asm.assembleNbcd([]Operand{op})
	if err != nil {
		t.Fatalf("NBCD data-register form failed: %v", err)
	}
}
