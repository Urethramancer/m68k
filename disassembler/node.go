package disassembler

import (
	"strconv"
	"strings"
)

// parseAbsoluteAddress tries to find an absolute hex/dec address inside an EA text.
// e.g. "(0x1234).w", "(0x001000).l" or "0x2000" -> returns integer address or -1.
func parseAbsoluteAddress(op string) int {
	if op == "" {
		return -1
	}
	// look for "$", then read contiguous hex digits
	i := strings.Index(op, "$")
	if i >= 0 {
		j := i + 1
		for j < len(op) && isHexDigit(op[j]) {
			j++
		}
		if j > i+1 {
			sub := op[i+1 : j]
			if v, err := strconv.ParseInt(sub, 16, 32); err == nil {
				return int(v)
			}
		}
	}
	// try decimal parse as fallback
	if v, err := strconv.Atoi(strings.Trim(op, " ()")); err == nil {
		return v
	}
	return -1
}

func isHexDigit(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}
