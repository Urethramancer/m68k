package assembler

import (
	"strings"
	"testing"
)

// Check that assembler chooses long branch when distance exceeds short range.
func TestBranchSizeSelection(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("start:\n")
	sb.WriteString("bra.s target\n")
	for i := 0; i < 300; i++ {
		sb.WriteString("dc.b $FF\n")
	}
	sb.WriteString("target:\n")
	sb.WriteString("nop\n")

	src := sb.String()
	_, err := Assemble(src, 0x1000)
	if err == nil {
		t.Fatalf("expected assemble to fail for forced short branch out of range")
	}
}
