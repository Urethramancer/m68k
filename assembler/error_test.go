package assembler

import (
	"testing"
)

func TestAssemblerErrors(t *testing.T) {
	tests := []string{
		"move.x d0,d1",        // invalid size
		"move #$100000000,d0", // immediate too large for move.l
		"dc.w $zzzz",          // invalid hex
		"movem d0-d8,d0",      // invalid register range
		"bra label",           // undefined label (no label present)
	}

	for _, src := range tests {
		if _, err := Assemble(src, 0); err == nil {
			t.Errorf("expected error for input: %q", src)
		}
	}
}
