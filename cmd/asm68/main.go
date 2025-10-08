package main

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/Urethramancer/m68k/assembler"
)

// isLittleEndian reports whether the current system uses little-endian byte order.
func isLittleEndian() bool {
	var x uint16 = 1
	b := [2]byte{}
	binary.LittleEndian.PutUint16(b[:], x)
	return b[0] == 1
}

// swapToBigEndian ensures the output is in M68K big-endian word order.
func swapToBigEndian(code []byte) {
	for i := 0; i+1 < len(code); i += 2 {
		code[i], code[i+1] = code[i+1], code[i]
	}
}

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
	code, err := asm.Assemble(string(src), 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Assembly error: %v\n", err)
		os.Exit(1)
	}

	// Always output M68K big-endian
	if isLittleEndian() {
		swapToBigEndian(code)
	}

	if outputFile == "" {
		// Print as hex dump for inspection
		for i, b := range code {
			fmt.Printf("%02X ", b)
			if (i+1)%16 == 0 {
				fmt.Println()
			}
		}
		fmt.Println()
	} else {
		if err := os.WriteFile(outputFile, code, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Assembled binary written in M68K big-endian format to %s\n", outputFile)
	}
}
