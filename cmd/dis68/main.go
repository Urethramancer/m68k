package main

import (
	"fmt"
	"os"

	"github.com/Urethramancer/m68k/cpu"
	"github.com/Urethramancer/m68k/disassembler"
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <inputfile> [outputfile]\n", os.Args[0])
		os.Exit(1)
	}

	inputFile := os.Args[1]
	var outputFile string
	if len(os.Args) == 3 {
		outputFile = os.Args[2]
	}

	code, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Ensure M68K big-endian order
	if cpu.IsLittleEndianHost() {
		for i := 0; i+1 < len(code); i += 2 {
			code[i], code[i+1] = code[i+1], code[i]
		}
	}

	// Perform unified full disassembly
	text, err := disassembler.Disassemble(code)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Disassembly error: %v\n", err)
		os.Exit(1)
	}

	if outputFile == "" {
		fmt.Print(text)
		return
	}

	// Write to file
	if err := os.WriteFile(outputFile, []byte(text), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Disassembly written to %s\n", outputFile)
}
