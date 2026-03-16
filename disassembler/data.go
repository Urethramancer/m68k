package disassembler

import (
	"fmt"
	"strings"
)

// isPrintableASCII checks if a byte is a standard printable ASCII character.
func isPrintableASCII(b byte) bool {
	return b >= 0x20 && b <= 0x7E
}

// analyzeAndFormatData emits a data region as dc.b hex byte directives.
func analyzeAndFormatData(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	return formatHexBytes(data)
}

// formatHexBytes formats a slice of bytes into dc.b directives, 16 bytes per
// line. Each line is guaranteed to contain an even number of bytes so that the
// assembler's automatic dc.b even-padding doesn't insert unwanted zeros.
func formatHexBytes(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	const bytesPerLine = 16
	var sb strings.Builder

	// Fast path: total byte count is even — every 16-byte chunk is even and
	// the last chunk (≤16 bytes) is also even.
	if len(data)%2 == 0 {
		for i := 0; i < len(data); i += bytesPerLine {
			end := i + bytesPerLine
			if end > len(data) {
				end = len(data)
			}
			writeHexLine(&sb, data[i:end])
		}
		return sb.String()
	}

	// Odd total: emit full 16-byte lines, then split the tail into two
	// even-length lines so no single dc.b has an odd byte count.
	i := 0
	remaining := len(data)
	for remaining > bytesPerLine {
		writeHexLine(&sb, data[i:i+bytesPerLine])
		i += bytesPerLine
		remaining -= bytesPerLine
	}
	// remaining ≤ 16 and is odd. Split into even + odd-1 (both even).
	if remaining > 1 {
		split := remaining/2*2 // largest even ≤ remaining
		writeHexLine(&sb, data[i:i+split])
		if split < remaining {
			writeHexLine(&sb, data[i+split:i+remaining])
		}
	} else {
		// Single byte — unavoidable; assembler will pad.
		writeHexLine(&sb, data[i:i+remaining])
	}

	return sb.String()
}

// writeHexLine writes a single "    dc.b    $NN,$NN,...\n" line.
func writeHexLine(sb *strings.Builder, chunk []byte) {
	sb.WriteString("    dc.b    ")
	for j, b := range chunk {
		if j > 0 {
			sb.WriteByte(',')
		}
		fmt.Fprintf(sb, "$%02x", b)
	}
	sb.WriteByte('\n')
}
