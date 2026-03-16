package disassembler

import (
	"testing"

	"github.com/Urethramancer/m68k/assembler"
)

func TestRoundTripSimple(t *testing.T) {
	src := "move.l d0,d1\naddq.w #1,d1\n"
	bytes, err := assembler.Assemble(src, 0)
	if err != nil {
		t.Fatalf("assemble failed: %v", err)
	}
	out, err := Disassemble(bytes)
	if err != nil {
		t.Fatalf("disassemble failed: %v", err)
	}
	// Try to re-assemble the disassembly output
	// Remove leading spaces from each line to make it suitable for assembler
	var cleaned string
	for _, line := range splitLines(out) {
		if len(line) == 0 {
			continue
		}
		// skip labels
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
		t.Fatalf("re-assemble failed: %v\nDisassembly:\n%s", err, out)
	}
	if len(bytes) != len(bytes2) {
		t.Fatalf("round-trip byte length mismatch: got %d reassembled vs %d original", len(bytes2), len(bytes))
	}
	for i := range bytes {
		if bytes[i] != bytes2[i] {
			t.Fatalf("byte mismatch at %d: orig 0x%02x reasm 0x%02x\nDisassembly:\n%s", i, bytes[i], bytes2[i], out)
		}
	}
}

func splitLines(s string) []string {
	var lines []string
	cur := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}
