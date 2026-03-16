package assembler

import (
	"testing"

	"github.com/Urethramancer/m68k/cpu"
)

func TestMovep_RegisterToMemory(t *testing.T) {
	asm := newAssembler()
	src := Operand{Mode: cpu.ModeData, Register: 2}
	dst := Operand{Mode: cpu.ModeAddrDisp, Register: 3, ExtensionWords: []uint16{0x0010}}
	mn := Mnemonic{Value: "movep", Size: cpu.SizeWord}
	words, err := asm.assembleMovep(mn, []Operand{src, dst})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(words) != 2 {
		t.Fatalf("expected 2 words, got %d", len(words))
	}
}

func TestMovep_MemoryToRegister_LongSize(t *testing.T) {
	asm := newAssembler()
	src := Operand{Mode: cpu.ModeAddrDisp, Register: 1, ExtensionWords: []uint16{0xFFFE}}
	dst := Operand{Mode: cpu.ModeData, Register: 0}
	mn := Mnemonic{Value: "movep", Size: cpu.SizeLong}
	words, err := asm.assembleMovep(mn, []Operand{src, dst})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(words) != 2 {
		t.Fatalf("expected 2 words, got %d", len(words))
	}
}
