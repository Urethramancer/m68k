package main

import (
	"fmt"
	"os"

	"github.com/Urethramancer/m68k/assembler"
)

func main() {
	// Load the .s or .asm file specified as the first argument.
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	asm := assembler.New()
	code, err := asm.Assemble(string(data), 0x1000)
	if err != nil {
		panic(err)
	}

	// Correctly print the byte slice as hex words.
	for i := 0; i < len(code); i += 2 {
		if i > 0 {
			fmt.Print(" ")
		}
		// Ensure we don't read past the end of the slice if there's an odd number of bytes.
		if i+1 < len(code) {
			fmt.Printf("%02x%02x", code[i], code[i+1])
		} else {
			fmt.Printf("%02x", code[i])
		}
	}
	fmt.Println()
}
