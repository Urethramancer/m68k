package disassembler

import (
	"fmt"
	"testing"

	"github.com/Urethramancer/m68k/assembler"
)

func TestInspectBranch(t *testing.T) {
	src := "start: bra.s start\n"
	asm := assembler.New()
	bytes, err := asm.Assemble(src, 0)
	if err != nil {
		t.Fatalf("assemble failed: %v", err)
	}
	fmt.Printf("assembled bytes: % x\n", bytes)
	out, err := Disassemble(bytes)
	if err != nil {
		t.Fatalf("disassemble failed: %v", err)
	}
	fmt.Printf("disassembly:\n%s\n", out)
}
