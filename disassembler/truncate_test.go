package disassembler

import (
	"testing"
)

// Ensure disassembler handles truncated extension words gracefully
func TestDisassemblerTruncatedExtensions(t *testing.T) {
	op := uint16(0x48E7) // movem.l <mask>,-(a7)
	ext := []byte{0x00}
	mn, _, _ := Decode(op, 0, ext)
	if mn != "movem.l" {
		t.Fatalf("expected movem.l mnemonic, got %q", mn)
	}
}
