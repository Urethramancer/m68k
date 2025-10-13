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

	var fn string
	if len(os.Args) == 3 {
		fn = os.Args[2]
	}

	src, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading source file: %v\n", err)
		os.Exit(1)
	}

	asm := assembler.New()
	code, err := asm.Assemble(string(src), 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Assembly error: %v\n", err)
		os.Exit(1)
	}

	if fn == "" {
		// Print as hex dump for inspection
		for i, b := range code {
			fmt.Printf("%02X ", b)
			if (i+1)%16 == 0 {
				fmt.Println()
			}
		}
		fmt.Println()
	} else {
		if err := os.WriteFile(fn, code, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Assembled binary written in M68K big-endian format to %s\n", fn)
	}
}
