package main

import (
	"log"

	"github.com/Urethramancer/m68k/assembler"
	"github.com/Urethramancer/m68k/vm"
)

func main() {
	log.SetFlags(0)
	log.Println("--- M68k Virtual Machine Runner ---")

	v := vm.New(1024*1024, 1024) // 1MB RAM, 1k cache

	// New test program with ADD instruction.
	// It tests basic addition, overflow, and carry flags.
	source := `
	org $1000
	move.l #10, d0
	add.l  #5, d0        ; d0 should be 15. Flags: ----
	move.l #$7FFFFFFF, d1 ; Move max positive long
	add.l  #1, d1        ; d1 should be $80000000. Flags: V--N-
	move.l #$FFFFFFFF, d2 ; Move -1
	add.l  #1, d2        ; d2 should be 0. Flags: --CZ-
`
	asm := assembler.New()
	code, err := asm.Assemble(source, 0x1000)
	if err != nil {
		log.Fatalf("Assembly failed: %s", err)
	}

	v.LoadCode(0x1000, code)
	log.Printf("Loaded %d bytes at address %08X\n", len(code), 0x1000)
	log.Println()

	log.Println("--- CPU State Before Execution ---")
	v.DumpRegisters()

	// Set the CPU to the running state before execution.
	v.CPU.Running = true

	// Execute all instructions in the program.
	for i := 0; i < 6; i++ {
		err := v.CPU.Execute()
		if err != nil {
			log.Fatalf("CPU execution failed: %s at %08X", err, v.CPU.PC-2)
		}
	}
	log.Println()

	log.Println("--- CPU State After Execution ---")
	v.DumpRegisters()

	log.Println("\nExecution finished successfully.")
}
