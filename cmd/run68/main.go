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

	// Opcodes:
	// movea.l: 207c <0000 2000> -> move immediate long to a0
	// move.w:  30bc <beef> -> move immediate word to (a0)
	source := `
	org $1000
	movea.l #$2000, a0
	move.w  #$beef, (a0)
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
	v.DumpMemory(0x2000, 16)

	// Set the CPU to the running state before execution.
	v.CPU.Running = true

	// Execute instructions one by one
	for i := 0; i < 2; i++ {
		err := v.CPU.Execute()
		if err != nil {
			log.Fatalf("CPU execution failed: %s at %08X", err, v.CPU.PC-2)
		}
	}
	log.Println()

	log.Println("--- CPU State After Execution ---")
	v.DumpRegisters()
	v.DumpMemory(0x2000, 16)

	log.Println("\nExecution finished successfully.")
}
