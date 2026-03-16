package disassembler

import (
	"os"
)

// DisassembleAndDump is a temporary helper for capture runs: it calls Disassemble,
// writes the produced disassembly to /tmp/capture_disasm_inside.txt synchronously,
// and returns the disassembly. Remove this helper after debugging.
func DisassembleAndDump(code []byte) (string, error) {
	out, err := Disassemble(code)
	if err != nil {
		return out, err
	}
	// Best-effort write; ignore errors to avoid changing Disassemble semantics.
	_ = os.WriteFile("/tmp/capture_disasm_inside.txt", []byte(out), 0644)
	// Also write a tmp atomic file for redundancy
	_ = os.WriteFile(os.TempDir()+"/capture_disasm_inside.tmp", []byte(out), 0644)
	return out, nil
}
