package assembler

import (
	"testing"
)

func TestParseAddressIndex_OutOfRange(t *testing.T) {
	asm := New()
	_, err := asm.parseAddressIndex([]string{"128(a0,d1.w)", "128", "0", "d", "1", "w"})
	if err == nil {
		t.Fatalf("expected error for out-of-range index displacement")
	}
}

func TestParsePCRelIndex_OutOfRange(t *testing.T) {
	asm := New()
	_, err := asm.parsePCRelIndex([]string{"128(pc,d1.w)", "128", "d", "1", "w"})
	if err == nil {
		t.Fatalf("expected error for out-of-range pc-index displacement")
	}
}

func TestParseConstant_CharEscapes(t *testing.T) {
	asm := New()
	v, err := asm.parseConstant("'\\n'")
	if err != nil || v != int64('\n') {
		t.Fatalf("expected newline char, got %v err=%v", v, err)
	}
	v, err = asm.parseConstant("'\\\\'")
	if err != nil || v != int64('\\') {
		t.Fatalf("expected backslash char, got %v err=%v", v, err)
	}
	v, err = asm.parseConstant("'A'")
	if err != nil || v != int64('A') {
		t.Fatalf("expected 'A' char, got %v err=%v", v, err)
	}
}

func TestTryParseImmediate_SizeSelection(t *testing.T) {
	asm := New()
	op, ok, err := asm.tryParseImmediateMode("#$12345")
	if err != nil || !ok {
		t.Fatalf("expected immediate parsed, got err=%v ok=%v", err, ok)
	}
	if len(op.ExtensionWords) != 2 {
		t.Fatalf("expected 2 extension words for large immediate, got %d", len(op.ExtensionWords))
	}

	op, ok, err = asm.tryParseImmediateMode("#$1234")
	if err != nil || !ok {
		t.Fatalf("expected immediate parsed, got err=%v ok=%v", err, ok)
	}
	if len(op.ExtensionWords) != 1 {
		t.Fatalf("expected 1 extension word for small immediate, got %d", len(op.ExtensionWords))
	}
}
