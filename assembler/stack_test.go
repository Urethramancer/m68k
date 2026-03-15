package assembler

import (
	"testing"

	"github.com/Urethramancer/m68k/cpu"
)

func TestLink_Unlk(t *testing.T) {
	// LINK A6,#-32
	var tmp int16 = -32
	v := uint16(tmp)
	imm := Operand{Raw: "#-32", Mode: cpu.ModeOther, Register: cpu.RegImmediate, ExtensionWords: []uint16{v}}
	reg := Operand{Mode: cpu.ModeAddr, Register: 6}
	words, err := assembleLink([]Operand{reg, imm})
	if err != nil {
		t.Fatalf("assembleLink failed: %v", err)
	}
	if len(words) != 2 {
		t.Fatalf("expected 2 words for LINK, got %d", len(words))
	}

	// UNLK A6
	w, err := assembleUnlk([]Operand{reg})
	if err != nil {
		t.Fatalf("assembleUnlk failed: %v", err)
	}
	if len(w) != 1 {
		t.Fatalf("expected 1 word for UNLK, got %d", len(w))
	}
}
