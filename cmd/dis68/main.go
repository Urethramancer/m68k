package main

import (
	"fmt"
	"os"

	"github.com/Urethramancer/m68k/disassembler"
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <inputfile> [outputfile]\n", os.Args[0])
		os.Exit(1)
	}

	var fn string
	if len(os.Args) == 3 {
		fn = os.Args[2]
	}

	// Read the binary file directly. Do NOT modify it.
	code, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// If an output file is specified, run the disassembler and write to it.
	if fn != "" {
		text, err := disassembler.Disassemble(code)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Disassembly error: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(fn, []byte(text), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Disassembly written to %s\n", fn)
		return
	}

	// If no output file is specified, print a hexdump of the input binary to the console.
	disassembler.Hexdump(code)
}
