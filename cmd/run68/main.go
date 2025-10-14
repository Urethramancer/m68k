package main

import (
	"fmt"
	"log"

	"github.com/Urethramancer/m68k/vm"
)

// This program provides a simple command-line interface to load and run
// a small M68k program and inspect the results.
func main() {
	fmt.Println("--- M68k Virtual Machine Runner ---")
	// Create a new VM with 1MB of RAM.
	v := vm.New(1024*1024, 4096)

	// Define a simple program to test the MOVEQ instruction.
	// 0x7042 -> MOVEQ #$42, D0
	// This should load the hexadecimal value 42 into data register D0.
	code := []byte{0x70, 0x42}

	// Load the code at address 0x1000 and set the PC to that address.
	startAddress := uint32(0x1000)
	v.LoadCode(startAddress, code)

	fmt.Printf("Loaded %d bytes at address %08X\n\n", len(code), startAddress)

	// --- Show state BEFORE execution ---
	fmt.Println("--- CPU State Before Execution ---")
	v.DumpRegisters()

	// Execute one instruction.
	err := v.CPU.Execute()
	if err != nil {
		log.Fatalf("CPU execution failed: %v", err)
	}

	// --- Show state AFTER execution ---
	fmt.Println("\n--- CPU State After Execution ---")
	v.DumpRegisters()

	fmt.Println("\nExecution finished successfully.")
}
