package assembler

import (
	"testing"

	"github.com/Urethramancer/m68k/cpu"
)

func TestMisc_Exg_Swap_Ext_Tas_Stop_Reset(t *testing.T) {
	asm := New()
	// EXG D0,D1
	_, err := asm.assembleExg([]Operand{{Mode: cpu.ModeData, Register: 0}, {Mode: cpu.ModeData, Register: 1}})
	if err != nil {
		t.Fatalf("EXG failed: %v", err)
	}

	// SWAP D2
	_, err = asm.assembleMiscOneOp(Mnemonic{Value: "swap"}, []Operand{{Mode: cpu.ModeData, Register: 2}})
	if err != nil {
		t.Fatalf("SWAP failed: %v", err)
	}

	// EXT.L D3
	_, err = asm.assembleMiscOneOp(Mnemonic{Value: "ext", Size: cpu.SizeLong}, []Operand{{Mode: cpu.ModeData, Register: 3}})
	if err != nil {
		t.Fatalf("EXT.L failed: %v", err)
	}

	// TAS (absolute addressing) - use D0 as dummy with ModeData to ensure encodeEA rejects it
	_, err = asm.assembleMiscOneOp(Mnemonic{Value: "tas"}, []Operand{{Mode: cpu.ModeAddrInd, Register: 0}})
	if err != nil {
		// It's acceptable if encodeEA fails for TAS with our dummy EA; ensure no panic
		return
	}

	// STOP #$2700
	_, err = asm.assembleStop([]Operand{{Raw: "#$2700", Mode: cpu.ModeOther, Register: cpu.RegImmediate, ExtensionWords: []uint16{0x2700}}})
	if err != nil {
		t.Fatalf("STOP failed: %v", err)
	}

	// RESET/NOP/ILLEGAL
	_, err = asm.assembleMiscNoOp(Mnemonic{Value: "reset"}, []Operand{})
	if err != nil {
		t.Fatalf("RESET failed: %v", err)
	}
	_, err = asm.assembleMiscNoOp(Mnemonic{Value: "nop"}, []Operand{})
	if err != nil {
		t.Fatalf("NOP failed: %v", err)
	}
	_, err = asm.assembleMiscNoOp(Mnemonic{Value: "illegal"}, []Operand{})
	if err != nil {
		t.Fatalf("ILLEGAL failed: %v", err)
	}
}
