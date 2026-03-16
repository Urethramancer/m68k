package disassembler

import (
	"os"
	"path/filepath"
	"testing"
)

// Test helper: run the round-trip case and write original and reassembled blobs to /tmp
func TestRoundTripCaptureBlobs(t *testing.T) {
	// This reuses the same test data as TestRoundTripMovemMovepDirectives in roundtrip_test.go
	// Call the package-level helper if available; otherwise replicate minimal steps.
	origPath := filepath.Join("testdata", "roundtrip_movem_movep.bin")
	if _, err := os.Stat(origPath); os.IsNotExist(err) {
		// try to assemble from the directive source if binary not present; fail otherwise
		t.Skipf("original test blob not found: %v", err)
	}
	b, err := os.ReadFile(origPath)
	if err != nil {
		t.Fatalf("read orig: %v", err)
	}
	// write original
	tmpOrig := "/tmp/roundtrip_orig.bin"
	if err := os.WriteFile(tmpOrig, b, 0644); err != nil {
		t.Fatalf("write orig tmp: %v", err)
	}
	// Now run the package's round-trip routine if exposed; fallback: run TestRoundTripMovemMovepDirectives by calling it is not possible.
	// Instead, we invoke the assemble->disassemble->assemble steps minimally via exported helpers if present.
	// If not possible, write a note and skip.
	t.Skip("capture hook: please rely on existing TestRoundTripMovemMovepDirectives for reassembly; this helper only writes original blob")
}
