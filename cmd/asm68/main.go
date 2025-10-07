package main

import (
	"fmt"
	"os"

	"github.com/Urethramancer/m68k/assembler"
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <sourcefile> [outputfile]\n", os.Args[0])
		os.Exit(1)
	}

	inputFile := os.Args[1]
	var outputFile string
	if len(os.Args) == 3 {
		outputFile = os.Args[2]
	}

	src, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading source file: %v\n", err)
		os.Exit(1)
	}

	asm := assembler.New()
	// Default base address 0 for now — may be changed later if .org is used
	code, err := asm.Assemble(string(src), 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Assembly error: %v\n", err)
		os.Exit(1)
	}

	if outputFile == "" {
		// Write machine code as hex dump to stdout
		for i, b := range code {
			fmt.Printf("%02X ", b)
			if (i+1)%16 == 0 {
				fmt.Println()
			}
		}
		fmt.Println()
	} else {
		err = os.WriteFile(outputFile, code, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Assembled binary written to %s\n", outputFile)
	}
}
