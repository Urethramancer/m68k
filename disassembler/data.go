package disassembler

import (
	"fmt"
	"strings"
)

// isPrintableASCII checks if a byte is a standard printable ASCII character.
func isPrintableASCII(b byte) bool {
	return b >= 0x20 && b <= 0x7E
}
func analyzeAndFormatData(data []byte, baseAddr uint32, stringCounter *int) string {
	var sb strings.Builder
	n := len(data)
	if n == 0 {
		return ""
	}

	i := 0
	minStrLen := 4

	for i < n {
		// Skip non-printables first
		start := i
		for start < n && !isPrintableASCII(data[start]) {
			start++
		}
		if start > i {
			sb.WriteString(formatHexBytes(data[i:start]))
		}

		// Find printable run
		end := start
		for end < n && isPrintableASCII(data[end]) {
			end++
		}
		if end <= start {
			i = start
			continue
		}

		run := data[start:end]
		runAddr := baseAddr + uint32(start)
		isNullTerminated := end < n && data[end] == 0x00

		// Rule 1: printable + NUL ≥ 4 chars → string
		if isNullTerminated && len(run) >= minStrLen {
			label := fmt.Sprintf("string%d:", *stringCounter)
			(*stringCounter)++
			escaped := strings.ReplaceAll(string(run), "'", "''")
			sb.WriteString(fmt.Sprintf("%-8s dc.b    '%s',$00\n", label, escaped))
			i = end + 1
			continue
		}

		// Rule 2: 4-byte aligned, 4 printable chars → tag
		if len(run) == 4 && allPrintable(run) && runAddr%4 == 0 {
			label := fmt.Sprintf("string%d:", *stringCounter)
			(*stringCounter)++
			escaped := strings.ReplaceAll(string(run), "'", "''")
			sb.WriteString(fmt.Sprintf("%-8s dc.b    '%s'\n", label, escaped))
			i = end
			continue
		}

		// Rule 3: anything else, emit as hex
		sb.WriteString(formatHexBytes(run))
		i = end
	}

	return sb.String()
}

// allPrintable reports whether all bytes are standard printable ASCII.
func allPrintable(b []byte) bool {
	for _, c := range b {
		if !isPrintableASCII(c) {
			return false
		}
	}
	return true
}

// formatHexBytes formats a slice of bytes into `dc.b` directives, 16 bytes per line.
func formatHexBytes(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	var sb strings.Builder
	const bytesPerLine = 16

	for i := 0; i < len(data); i += bytesPerLine {
		end := i + bytesPerLine
		if end > len(data) {
			end = len(data)
		}
		chunk := data[i:end]

		sb.WriteString("    dc.b    ")
		for j, b := range chunk {
			if j > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(fmt.Sprintf("$%02x", b))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
