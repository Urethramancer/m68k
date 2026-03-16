package disassembler

import (
	"fmt"
	"strings"
)

// isPrintableASCII checks if a byte is a standard printable ASCII character.
func isPrintableASCII(b byte) bool {
	return b >= 0x20 && b <= 0x7E
}

// analyzeAndFormatData emits a data region. Runs of zero bytes are collapsed
// into ds.l, ds.w or ds.b directives; non-zero bytes are emitted as dc.b hex.
func analyzeAndFormatData(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	var sb strings.Builder
	i := 0
	for i < len(data) {
		// Count consecutive zero bytes starting at i.
		zeroStart := i
		for i < len(data) && data[i] == 0 {
			i++
		}
		zeroLen := i - zeroStart
		if zeroLen > 0 {
			sb.WriteString(formatZeroRun(zeroLen))
			continue
		}

		// Count consecutive non-zero (or mixed) bytes until the next
		// zero run worth collapsing (4+ zeros).
		nonZeroStart := i
		for i < len(data) {
			// Look ahead for a collapsible zero run.
			if data[i] == 0 {
				j := i
				for j < len(data) && data[j] == 0 {
					j++
				}
				if j-i >= 4 {
					break // Let the zero-run handler pick this up.
				}
				i = j // Short zero run — include in non-zero chunk.
			} else {
				i++
			}
		}
		sb.WriteString(formatHexBytes(data[nonZeroStart:i]))
	}
	return sb.String()
}

// formatZeroRun emits the most compact ds.X directive for a run of n zero
// bytes. Prefers ds.l when longword-aligned, then ds.w, then ds.b.
func formatZeroRun(n int) string {
	if n == 0 {
		return ""
	}
	var sb strings.Builder
	if n >= 4 {
		longs := n / 4
		fmt.Fprintf(&sb, "    ds.l    %d\n", longs)
		n -= longs * 4
	}
	if n >= 2 {
		words := n / 2
		fmt.Fprintf(&sb, "    ds.w    %d\n", words)
		n -= words * 2
	}
	if n > 0 {
		fmt.Fprintf(&sb, "    ds.b    %d\n", n)
	}
	return sb.String()
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
		split := remaining / 2 * 2 // largest even ≤ remaining
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
