package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Urethramancer/m68k/assembler"
	"github.com/Urethramancer/m68k/disassembler"
)

func main() {
	src := ".org 0\n.dc.b \"This is a test string.\", 0xDE, 0xAD, 0xBE, 0xEF\n.even\n.dc.w 0x1234, 42\n"
	asm := assembler.New()
	bytes, err := asm.Assemble(src, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "assemble failed: %v\n", err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile("/tmp/capture_orig.bin", bytes, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write orig failed: %v\n", err)
		os.Exit(1)
	}
	out, err := disassembler.Disassemble(bytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "disassemble failed: %v\n", err)
		os.Exit(1)
	}
	// clean and reassemble same as test
	cleaned := ""
	for _, line := range splitLines(out) {
		if len(line) == 0 {
			continue
		}
		if line[len(line)-1] == ':' {
			cleaned += line + "\n"
			continue
		}
		i := 0
		for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
			i++
		}
		cleaned += line[i:] + "\n"
	}

	// write disassembly for inspection
	_ = ioutil.WriteFile("/tmp/capture_disasm.txt", []byte(out), 0644)

	reAsm := assembler.New()
	bytes2, err := reAsm.Assemble(cleaned, 0)
	if err != nil {
		// write assembler error and the cleaned disassembly for debugging
		_ = ioutil.WriteFile("/tmp/capture_reasm_error.txt", []byte(fmt.Sprintf("err: %v\n\ncleaned:\n%s\n", err, cleaned)), 0644)
		os.Exit(1)
	}
	if err := ioutil.WriteFile("/tmp/capture_reasm.bin", bytes2, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write reasm failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wrote /tmp/capture_orig.bin (%d bytes), /tmp/capture_disasm.txt, and /tmp/capture_reasm.bin (%d bytes)\n", len(bytes), len(bytes2))
}

func splitLines(s string) []string {
	var lines []string
	cur := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}
