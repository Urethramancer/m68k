package assembler

import (
	"testing"

	"github.com/Urethramancer/m68k/cpu"
)

func TestEncodeEA_ImmediateSizes(t *testing.T) {
	asm := New()
	// Byte immediate
	op := Operand{Raw: "#$7F", Mode: cpu.ModeOther, Register: cpu.ModeImmediate}
	w, exts, err := asm.encodeEA(op, cpu.SizeByte)
	if err != nil {
		t.Fatalf("encodeEA byte immediate failed: %v", err)
	}
	if len(exts) != 1 {
		t.Fatalf("expected 1 extension word for byte immediate, got %d", len(exts))
	}
	_ = w

	// Word immediate
	op = Operand{Raw: "#$1234", Mode: cpu.ModeOther, Register: cpu.ModeImmediate}
	_, exts, err = asm.encodeEA(op, cpu.SizeWord)
	if err != nil {
		t.Fatalf("encodeEA word immediate failed: %v", err)
	}
	if len(exts) != 1 {
		t.Fatalf("expected 1 extension word for word immediate, got %d", len(exts))
	}

	// Long immediate
	op = Operand{Raw: "#$12345678", Mode: cpu.ModeOther, Register: cpu.ModeImmediate}
	_, exts, err = asm.encodeEA(op, cpu.SizeLong)
	if err != nil {
		t.Fatalf("encodeEA long immediate failed: %v", err)
	}
	if len(exts) != 2 {
		t.Fatalf("expected 2 extension words for long immediate, got %d", len(exts))
	}
}

func TestShiftRotate_RegisterImmediateEncoding(t *testing.T) {
	asm := New()
	mn := Mnemonic{Value: "asl", Size: cpu.SizeWord}
	op1 := Operand{Raw: "#2", Mode: cpu.ModeOther, Register: cpu.RegImmediate}
	op2 := Operand{Mode: cpu.ModeData, Register: 3}
	words, err := asm.assembleShiftRotate(mn, []Operand{op1, op2})
	if err != nil {
		t.Fatalf("assembleShiftRotate failed: %v", err)
	}
	if len(words) != 1 {
		t.Fatalf("expected 1 word for register form, got %d", len(words))
	}
}

func TestShiftRotate_MemoryFormEncoding(t *testing.T) {
	asm := New()
	mn := Mnemonic{Value: "lsl", Size: cpu.SizeWord}
	op := Operand{Mode: cpu.ModeAddrInd, Register: 0}
	words, err := asm.assembleShiftRotate(mn, []Operand{op})
	if err != nil {
		t.Fatalf("assembleShiftRotate memory form failed: %v", err)
	}
	if len(words) < 1 {
		t.Fatalf("expected at least 1 word for memory form, got %d", len(words))
	}
}

func TestBitManipulation_ImmediateAndRegister(t *testing.T) {
	asm := New()
	// Immediate form
	mn := Mnemonic{Value: "btst"}
	op1 := Operand{Raw: "#1", Mode: cpu.ModeOther, Register: cpu.RegImmediate}
	op2 := Operand{Mode: cpu.ModeAddrInd, Register: 0}
	words, err := asm.assembleBitManipulation(mn, []Operand{op1, op2})
	if err != nil {
		t.Fatalf("assembleBitManipulation immediate failed: %v", err)
	}
	if len(words) < 2 {
		t.Fatalf("expected at least 2 words (op + ext) for immediate form, got %d", len(words))
	}

	// Register form
	op1 = Operand{Mode: cpu.ModeData, Register: 2}
	words, err = asm.assembleBitManipulation(mn, []Operand{op1, op2})
	if err != nil {
		t.Fatalf("assembleBitManipulation register form failed: %v", err)
	}
	if len(words) < 1 {
		t.Fatalf("expected at least 1 word for register form, got %d", len(words))
	}
}
