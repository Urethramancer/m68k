package assembler

import (
	"testing"

	"github.com/Urethramancer/m68k/cpu"
)

func TestTrap_ValidAndInvalid(t *testing.T) {
	asm := New()
	// Valid vector
	_, err := asm.assembleTrapImmediate([]Operand{{Raw: "#5", Mode: cpu.ModeOther, Register: cpu.RegImmediate, ExtensionWords: []uint16{5}}})
	if err != nil {
		t.Fatalf("expected valid trap vector 5, got err: %v", err)
	}
	// Invalid vector
	_, err = asm.assembleTrapImmediate([]Operand{{Raw: "#20", Mode: cpu.ModeOther, Register: cpu.RegImmediate, ExtensionWords: []uint16{20}}})
	if err == nil {
		t.Fatalf("expected error for trap vector out of range")
	}
	// TRAPV
	w, err := assembleTrapv([]Operand{})
	if err != nil {
		t.Fatalf("TRAPV failed: %v", err)
	}
	if len(w) != 1 {
		t.Fatalf("expected single word for TRAPV, got %d", len(w))
	}
}
