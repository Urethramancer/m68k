package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Urethramancer/m68k/disassembler"
)

func main() {
	orig, err := os.ReadFile("/tmp/orig.bin")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read orig: %v\n", err)
		os.Exit(2)
	}
	out, err := disassembler.DisassembleAndDump(orig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "disasm: %v\n", err)
		os.Exit(2)
	}
	// parse dc.b $xx tokens
	lines := strings.Split(out, "\n")
	re := regexp.MustCompile(`\$([0-9a-fA-F]{2})`)
	mapped := make([]int, len(orig))
	for i := range mapped {
		mapped[i] = -1
	}
	idx := 0
	for li, l := range lines {
		if strings.Contains(l, "dc.b") {
			matches := re.FindAllStringSubmatch(l, -1)
			for _, m := range matches {
				if idx >= len(orig) {
					break
				}
				v, _ := strconv.ParseInt(m[1], 16, 0)
				_ = byte(v)
				mapped[idx] = li
				idx++
			}
		}
	}
	// fallback: remaining bytes map to last line
	for idx < len(orig) {
		mapped[idx] = len(lines) - 1
		idx++
	}
	// read reasm
	reasm, _ := os.ReadFile("/tmp/reasm.bin")
	fmt.Printf("Offset Orig Reasm Line\n")
	for i := 0; i < len(orig) && i < len(reasm); i++ {
		if orig[i] != reasm[i] {
			line := "(none)"
			if mapped[i] >= 0 && mapped[i] < len(lines) {
				line = strings.TrimSpace(lines[mapped[i]])
			}
			fmt.Printf("%04x  0x%02x 0x%02x  line %d: %s\n", i, orig[i], reasm[i], mapped[i], line)
		}
	}
}
