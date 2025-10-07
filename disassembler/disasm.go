package disassembler

import (
	"strings"
)

// Disassemble takes a byte slice of M68k machine code and returns it as a
// formatted assembly language string.
func Disassemble(code []byte) (string, error) {
	var result strings.Builder
	return result.String(), nil
}
