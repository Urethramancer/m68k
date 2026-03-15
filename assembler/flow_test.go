package assembler

import (
	"testing"

	"github.com/Urethramancer/m68k/cpu"
)

func TestGetSizeBra_ShortVsWord(t *testing.T) {
	asm := New()
	// Create node with no explicit size and a forward label
	n := &Node{}
	n.Mnemonic.Value = "bra"
	n.Operands = []Operand{{Raw: "target"}}
	asm.labels = map[string]uint32{"target": 300}
	// pc small so offset fits in byte
	sz := asm.getSizeBra(n, 290)
	if sz != 2 {
		t.Fatalf("expected short size (2), got %d", sz)
	}
	// pc far so offset doesn't fit
	sz = asm.getSizeBra(n, 0)
	if sz != 4 {
		t.Fatalf("expected word size (4) for far branch, got %d", sz)
	}
}

func TestAssembleBra_OutOfRangeShort(t *testing.T) {
	labels := map[string]uint32{"L": 1000}
	_, err := assembleBra(Mnemonic{Value: "bra"}, []Operand{{Raw: "L"}}, labels, 0, 2)
	if err == nil {
		t.Fatalf("expected out-of-range error for short branch")
	}
}

func TestAssembleDbcc_OutOfRange(t *testing.T) {
	asm := New()
	asm.labels = map[string]uint32{"loop": 70000}
	_, err := asm.assembleDbcc(Mnemonic{Value: "dbne"}, []Operand{{Mode: cpu.ModeData, Register: 0}, {Raw: "loop"}}, asm.labels, 0)
	if err == nil {
		t.Fatalf("expected out-of-range error for DBcc")
	}
}
