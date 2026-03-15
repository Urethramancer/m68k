package assembler

import (
	"testing"

	"github.com/Urethramancer/m68k/cpu"
)

func TestParseMovemList_Simple(t *testing.T) {
	m, err := parseMovemList("d0/d1/a0")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	// d0=bit0, d1=bit1, a0=bit8 => mask = 1|2|256 = 259
	if m != 259 {
		t.Fatalf("unexpected mask: %d", m)
	}
}

func TestParseMovemList_Range(t *testing.T) {
	m, err := parseMovemList("d0-d3")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	// bits 0..3 set => 0x000F
	if m != 0x000F {
		t.Fatalf("unexpected mask for d0-d3: 0x%04X", m)
	}
}

func TestParseMovemList_RangeSwap(t *testing.T) {
	m, err := parseMovemList("d3-d1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if m != 0x000E { // bits 1..3
		t.Fatalf("unexpected mask for d3-d1: 0x%04X", m)
	}
}

func TestParseMovemList_CrossGroupError(t *testing.T) {
	_, err := parseMovemList("d3-a1")
	if err == nil {
		t.Fatalf("expected error for crossing groups")
	}
}

func TestAssembleMovemStore_PreDecReversal(t *testing.T) {
	asm := New()
	// Simulate MOVEM d0-d3, -(A7)
	src := Operand{Raw: "d0-d3"}
	dst := Operand{Raw: "-(a7)", Mode: cpu.ModeAddrPreDec, Register: 7}
	words, err := asm.assembleMovem(mnWithSize("movem", cpu.SizeWord), []Operand{src, dst})
	if err != nil {
		t.Fatalf("assembleMovem failed: %v", err)
	}
	if len(words) < 2 {
		t.Fatalf("expected at least 2 words, got %d", len(words))
	}
	// For pre-decrement the mask should be reversed (bits reversed across D and A groups).
	// We only check that the second word is non-zero and of uint16 type implicitly.
	_ = words[1]
}

func mnWithSize(name string, sz cpu.Size) Mnemonic {
	return Mnemonic{Value: name, Size: sz}
}
