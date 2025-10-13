package main

import (
	"fmt"
	"os"

	"github.com/Urethramancer/m68k/disassembler"
)

// hexdump prints data in the style of the 'hexdump -C' command.
func hexdump(data []byte) {
	const bytesPerLine = 16
	for i := 0; i < len(data); i += bytesPerLine {
		// Print the offset for the current line.
		fmt.Printf("%08x  ", i)

		// Print the hex values for the bytes in the line.
		for j := 0; j < bytesPerLine; j++ {
			if j == 8 {
				fmt.Print(" ") // Add an extra space in the middle.
			}
			if i+j < len(data) {
				fmt.Printf("%02x ", data[i+j])
			} else {
				fmt.Print("   ") // Pad with spaces if the line is short.
			}
		}

		// Print the ASCII representation.
		fmt.Print(" |")
		end := i + bytesPerLine
		if end > len(data) {
			end = len(data)
		}
		for _, b := range data[i:end] {
			if b >= 32 && b <= 126 {
				fmt.Printf("%c", b)
			} else {
				fmt.Print(".") // Use a dot for non-printable characters.
			}
		}
		fmt.Println("|")
	}
}

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <inputfile> [outputfile]\n", os.Args[0])
		os.Exit(1)
	}

	src := os.Args[1]
	var fn string
	if len(os.Args) == 3 {
		fn = os.Args[2]
	}

	// Read the binary file directly. Do NOT modify it.
	code, err := os.ReadFile(src)
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
	hexdump(code)
}
