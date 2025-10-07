package main

import (
	"fmt"
	"io/ioutil"
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

	code, err := ioutil.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	disasm, err := disassembler.Disassemble(code)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Disassembly error: %v\n", err)
		os.Exit(1)
	}

	if outputFile == "" {
		fmt.Println(disasm)
	} else {
		err = os.WriteFile(outputFile, []byte(disasm), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Disassembly written to %s\n", outputFile)
	}
}
