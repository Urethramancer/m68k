package assembler

import (
	"fmt"
	"testing"

	"github.com/Urethramancer/m68k/disassembler"
)

func TestRoundTripEncodeDecode(t *testing.T) {
	tests := []string{
		"move.l #$12345678,d0",
		"movem.l d0-d7,-(a7)",
		"lea ($10,pc),a0",
	}

	for i, src := range tests {
		name := fmt.Sprintf("case%d", i)
		code, err := Assemble(src, 0x1000)
		if err != nil {
			t.Fatalf("%s: assemble failed: %v", name, err)
		}

		text, err := disassembler.Disassemble(code)
		if err != nil {
			t.Fatalf("%s: disassemble failed: %v", name, err)
		}

		recode, err := Assemble(text, 0x1000)
		if err != nil {
			t.Fatalf("%s: re-assemble failed: %v\ntext:\n%s", name, err, text)
		}

		// Compare decoded first instruction (mnemonic + operands) to avoid
		// label-number differences in disassembly text.
		if len(code) < 2 || len(recode) < 2 {
			t.Fatalf("%s: output too small", name)
		}
		op1 := uint16(code[0])<<8 | uint16(code[1])
		var ext1 []byte
		if len(code) > 2 {
			ext1 = code[2:]
		}
		mn1, ops1, _ := disassembler.Decode(op1, 0x1000, ext1)

		op2 := uint16(recode[0])<<8 | uint16(recode[1])
		var ext2 []byte
		if len(recode) > 2 {
			ext2 = recode[2:]
		}
		mn2, ops2, _ := disassembler.Decode(op2, 0x1000, ext2)

		if mn1 != mn2 || ops1 != ops2 {
			t.Fatalf("%s: decoded mismatch after round-trip: %s %s != %s %s", name, mn1, ops1, mn2, ops2)
		}
	}
}
