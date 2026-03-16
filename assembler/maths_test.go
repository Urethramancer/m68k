package assembler

import (
	"testing"

	"github.com/Urethramancer/m68k/cpu"
)

func TestAdd_QuickImmediate(t *testing.T) {
	asm := newAssembler()
	mn := Mnemonic{Value: "add", Size: cpu.SizeWord}
	op1 := Operand{Raw: "#4", Mode: cpu.ModeOther, Register: cpu.RegImmediate}
	op2 := Operand{Mode: cpu.ModeData, Register: 2}
	words, err := asm.assembleAdd(mn, []Operand{op1, op2})
	if err != nil {
		t.Fatalf("assembleAdd quick immediate failed: %v", err)
	}
	if len(words) != 1 {
		t.Fatalf("expected single word for addq form, got %d", len(words))
	}
}

func TestAddI_ImmediateSizeWord(t *testing.T) {
	asm := newAssembler()
	mn := Mnemonic{Value: "addi", Size: cpu.SizeWord}
	op1 := Operand{Raw: "#$1234", Mode: cpu.ModeOther, Register: cpu.RegImmediate}
	op2 := Operand{Mode: cpu.ModeAddr, Register: 1}
	words, err := asm.assembleAdd(mn, []Operand{op1, op2})
	if err != nil {
		t.Fatalf("assembleAdd immediate failed: %v", err)
	}
	if len(words) < 2 {
		t.Fatalf("expected at least 2 words for addi with immediate ext, got %d", len(words))
	}
}

func TestAddA_AddressRegister(t *testing.T) {
	asm := newAssembler()
	mn := Mnemonic{Value: "adda", Size: cpu.SizeLong}
	op1 := Operand{Mode: cpu.ModeData, Register: 3}
	op2 := Operand{Mode: cpu.ModeAddr, Register: 2}
	words, err := asm.assembleAdd(mn, []Operand{op1, op2})
	if err != nil {
		t.Fatalf("assembleAdd adda failed: %v", err)
	}
	if len(words) < 1 {
		t.Fatalf("expected at least 1 word for adda, got %d", len(words))
	}
}

func TestMulDiv_SizeChecks(t *testing.T) {
	asm := newAssembler()
	// MUL word allowed
	mn := Mnemonic{Value: "muls", Size: cpu.SizeWord}
	op1 := Operand{Mode: cpu.ModeAddrInd, Register: 0}
	op2 := Operand{Mode: cpu.ModeData, Register: 0}
	_, err := asm.assembleMul(mn, []Operand{op1, op2})
	if err != nil {
		t.Fatalf("expected muls to accept word size, got err=%v", err)
	}
	// MUL with long should fail
	mn = Mnemonic{Value: "mulu", Size: cpu.SizeLong}
	_, err = asm.assembleMul(mn, []Operand{op1, op2})
	if err == nil {
		t.Fatalf("expected mulu with long to fail")
	}
}
