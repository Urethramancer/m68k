package disassembler

import (
	"encoding/binary"
	"testing"
)

// Test MOVEP decode: opword 0x0108 pattern handled earlier; build op and one extension
func TestDecodeMovep(t *testing.T) {
	// Construct MOVEP with data reg D2 (reg=2) and addr reg A3 (mode bits)
	op := uint16(0x0108) | (2 << 9) | 3
	code := make([]byte, 2+2)
	binary.BigEndian.PutUint16(code[0:], op)
	// displacement 0x0010
	binary.BigEndian.PutUint16(code[2:], 0x0010)
	mn, ops, used := TestableDecode(op, 0, code[2:])
	if mn != "movep" && mn != "movep.w" {
		t.Fatalf("expected movep mnemonic, got %s", mn)
	}
	if used != 2 {
		t.Fatalf("expected used=2 for movep ext, got %d", used)
	}
	if ops == "" {
		t.Fatalf("expected operands text, got empty")
	}
}

func TestDecodePCRelIndex(t *testing.T) {
	// Build extension word for PC-relative with index D1.w and displacement -4
	// mode=7 reg=3 for PC index -> ea = (7<<3)|3 = 0x3B
	ext := uint16(1<<12) | uint16(0xFC) // disp -4 as uint8 = 0xFC
	code := make([]byte, 2)
	binary.BigEndian.PutUint16(code[0:], ext)
	s, used := DecodeEA(0x3B, 0, code, 1)
	if used != 2 {
		t.Fatalf("expected used=2 for pc-rel index, got %d", used)
	}
	if s == "" {
		t.Fatalf("expected non-empty operand string for pc-rel index")
	}
}

func TestReadImmediateFormatting(t *testing.T) {
	// Byte immediate, value -2
	code := []byte{0xFF, 0xFE}
	s, used := readImmediateBySize(code, 0, 0)
	if used != 2 {
		t.Fatalf("expected used=2 for byte immediate, got %d", used)
	}
	if s != "#-2" {
		t.Fatalf("unexpected immediate text: %s", s)
	}

	// Word immediate, small positive should be decimal
	code = []byte{0x00, 0x7F}
	s, used = readImmediateBySize(code, 0, 1)
	if used != 2 {
		t.Fatalf("expected used=2 for word immediate, got %d", used)
	}
	if s != "#127" {
		t.Fatalf("unexpected word immediate text: %s", s)
	}

	// Word immediate large should be hex
	code = []byte{0x12, 0x34}
	s, used = readImmediateBySize(code, 0, 1)
	if used != 2 {
		t.Fatalf("expected used=2 for word immediate, got %d", used)
	}
	if s != "#$1234" {
		t.Fatalf("expected hex formatting for large word immediate, got %s", s)
	}
}
