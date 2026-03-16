package disassembler

import (
	"testing"

	"github.com/Urethramancer/m68k/assembler"
)

func TestRoundTripMovemMovepDirectives(t *testing.T) {
	cases := []string{
		// MOVEM variants
		"movem.l (a0)+,d0-d3/a0/a6\n",
		"movem.w d0-d7,-(a7)\n",
		// MOVEP variants
		"movep.w d2,$10(a3)\n",
		"movep.w $10(a3),d2\n",
		// LINK/UNLK
		"link a6,#-32\nunlk a6\n",
		// Branches with labels
		"start: bra.s start\nlabel1: beq.w label1\n",
		// TRAP/STOP
		"trap #5\nstop #$2700\n",
		// Directives: strings, padding and mixed data
		".org 0\n.dc.b \"This is a test string.\", 0xDE, 0xAD, 0xBE, 0xEF\n.even\n.dc.w 0x1234, 42\n",
	}

	for _, src := range cases {
		bytes, err := assembler.Assemble(src, 0)
		if err != nil {
			t.Fatalf("assemble failed for src:\n%s\nerr: %v", src, err)
		}
		out, err := Disassemble(bytes)
		if err != nil {
			t.Fatalf("disassemble failed: %v", err)
		}

		// Clean and re-assemble
		cleaned := ""
		for _, line := range splitLines(out) {
			if len(line) == 0 {
				continue
			}
			// skip labels unchanged
			if line[len(line)-1] == ':' {
				cleaned += line + "\n"
				continue
			}
			// strip leading spaces
			i := 0
			for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
				i++
			}
			cleaned += line[i:] + "\n"
		}

		bytes2, err := assembler.Assemble(cleaned, 0)
		if err != nil {
			t.Fatalf("re-assemble failed for disassembly:\n%s\nerr: %v", out, err)
		}
		if len(bytes) != len(bytes2) {
			t.Fatalf("round-trip length mismatch for src:\n%s\norig=%d reasm=%d\nDisassembly:\n%s", src, len(bytes), len(bytes2), out)
		}
		for i := range bytes {
			if bytes[i] != bytes2[i] {
				t.Fatalf("byte mismatch at %d for src:\n%s\norig 0x%02x reasm 0x%02x\nDisassembly:\n%s", i, src, bytes[i], bytes2[i], out)
			}
		}
	}
}
