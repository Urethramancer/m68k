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

	inputFile := os.Args[1]
	var outputFile string
	if len(os.Args) == 3 {
		outputFile = os.Args[2]
	}

	// Read the binary file directly. Do NOT modify it.
	code, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Pass the clean, unmodified, big-endian byte slice to the disassembler.
	text, err := disassembler.Disassemble(code)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Disassembly error: %v\n", err)
		os.Exit(1)
	}

	if outputFile == "" {
		fmt.Print(text)
		return
	}

	if err := os.WriteFile(outputFile, []byte(text), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Disassembly written to %s\n", outputFile)
}
