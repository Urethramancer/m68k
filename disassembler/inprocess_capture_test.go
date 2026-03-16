package disassembler

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Urethramancer/m68k/assembler"
)

func TestInProcessCaptureRoundTrip(t *testing.T) {
	t.Log("stage: assemble start")
	// Original sample used previously (includes a printable ASCII run).
	src := ".org 0\n.dc.b \"This is a test string.\", 0xDE, 0xAD, 0xBE, 0xEF\n.even\n.dc.w 0x1234, 42\n"
	asm := assembler.New()
	orig, err := asm.Assemble(src, 0)
	if err != nil {
		t.Fatalf("assemble failed: %v", err)
	}
	t.Logf("stage: assemble done (%d bytes)", len(orig))
	d := t.TempDir()
	origPath := filepath.Join(d, "orig.bin")
	if err := os.WriteFile(origPath, orig, 0644); err != nil {
		t.Fatalf("write orig failed: %v", err)
	}

	t.Log("stage: disassemble start")
	out, err := DisassembleAndDump(orig)
	if err != nil {
		t.Fatalf("disassemble failed: %v", err)
	}
	t.Log("stage: disassemble done")
	dasmPath := filepath.Join(d, "disasm.txt")
	if err := os.WriteFile(dasmPath, []byte(out), 0644); err != nil {
		t.Fatalf("write disasm failed: %v", err)
	}

	// attempt reassembly
	t.Log("stage: reassembly prepare")
	cleaned := ""
	for _, line := range splitLines(out) {
		if len(line) == 0 {
			continue
		}
		if line[len(line)-1] == ':' {
			cleaned += line + "\n"
			continue
		}
		i := 0
		for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
			i++
		}
		cleaned += line[i:] + "\n"
	}
	t.Log("stage: reassembly start")
	reAsm := assembler.New()
	reasmBytes, err := reAsm.Assemble(cleaned, 0)
	if err != nil {
		errPath := filepath.Join(d, "reasm_error.txt")
		_ = os.WriteFile(errPath, []byte(err.Error()), 0644)
		t.Fatalf("reassemble failed: %v (wrote %s)", err, errPath)
	}
	t.Log("stage: reassembly done")
	reasmPath := filepath.Join(d, "reasm.bin")
	if err := os.WriteFile(reasmPath, reasmBytes, 0644); err != nil {
		t.Fatalf("write reasm failed: %v", err)
	}

	if len(reasmBytes) != len(orig) {
		t.Fatalf("length mismatch: orig %d bytes vs reasm %d bytes", len(orig), len(reasmBytes))
	}

	if !equal(orig, reasmBytes) {
		// produce hex diff
		diffs := make([]string, 0)
		for i := 0; i < len(orig); i++ {
			if orig[i] != reasmBytes[i] {
				diffs = append(diffs, fmt.Sprintf("%04x: orig %02x reasm %02x", i, orig[i], reasmBytes[i]))
			}
		}
		hexOrig := hex.EncodeToString(orig)
		hexReasm := hex.EncodeToString(reasmBytes)
		_ = os.WriteFile(filepath.Join(d, "hex_orig.txt"), []byte(hexOrig), 0644)
		_ = os.WriteFile(filepath.Join(d, "hex_reasm.txt"), []byte(hexReasm), 0644)
		t.Fatalf("bytes differ: %d differing bytes:\n%s\nArtifacts in %s", len(diffs), joinLines(diffs), d)
	}
}

func equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func joinLines(ss []string) string {
	out := ""
	for _, s := range ss {
		out += s + "\n"
	}
	return out
}
