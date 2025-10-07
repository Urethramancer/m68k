package assembler

import (
	"fmt"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

// getDirectiveSize calculates the byte size of a directive for the sizing pass.
func (asm *Assembler) getDirectiveSize(n *Node) (uint32, error) {
	directive := strings.ToLower(n.Parts[0])

	switch directive {
	case ".org", ".equ":
		return 0, nil // These directives occupy no space in output.

	case ".dc.b", ".dc.w", ".dc.l":
		if len(n.Parts) < 2 {
			return 0, fmt.Errorf("%s requires at least one value", directive)
		}
		values := strings.Join(n.Parts[1:], " ")
		return calculateDcSize(directive, values, asm)

	case ".ds.b", ".ds.w", ".ds.l":
		if len(n.Parts) != 2 {
			return 0, fmt.Errorf("%s requires a single count argument", directive)
		}
		count, err := parseConstant(n.Parts[1], asm)
		if err != nil {
			return 0, fmt.Errorf("invalid count for %s: %v", directive, err)
		}
		elementSize := getElementSize(directive)
		return uint32(count) * elementSize, nil

	default:
		return 0, fmt.Errorf("unknown directive: %s", directive)
	}
}

// generateDirectiveCode generates the binary data for assembler directives.
func (asm *Assembler) generateDirectiveCode(n *Node) ([]uint16, error) {
	directive := strings.ToLower(n.Parts[0])

	switch directive {
	case ".org", ".equ":
		return nil, nil // No data emitted.

	case ".dc.b", ".dc.w", ".dc.l":
		values := strings.Join(n.Parts[1:], " ")
		return assembleDc(directive, values, asm)

	case ".ds.b", ".ds.w", ".ds.l":
		count, err := parseConstant(n.Parts[1], asm)
		if err != nil {
			return nil, fmt.Errorf("invalid count for %s: %v", directive, err)
		}
		elementSize := getElementSize(directive)
		byteSize := uint32(count) * elementSize
		// Pad to even byte boundary (assembler emits words)
		wordSize := (byteSize + 1) / 2
		return make([]uint16, wordSize), nil

	default:
		return nil, fmt.Errorf("unknown directive: %s", directive)
	}
}

// calculateDcSize determines the byte size of a .dc directive's data.
func calculateDcSize(directive, values string, asm *Assembler) (uint32, error) {
	elementSize := getElementSize(directive)
	var size uint32

	// Handle string literal data for .dc.b
	if elementSize == 1 && strings.Contains(values, "\"") {
		inQuote := false
		for _, c := range values {
			if c == '"' {
				inQuote = !inQuote
			} else if inQuote {
				size++
			}
		}
	} else {
		// Handle comma-separated numeric constants
		for _, p := range strings.Split(values, ",") {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				size += elementSize
			}
		}
	}

	// Align to the next word boundary so size matches generation pass.
	return (size + 1) &^ 1, nil
}

// assembleDc generates machine data for .dc directives (.dc.b/.dc.w/.dc.l).
func assembleDc(directive, values string, asm *Assembler) ([]uint16, error) {
	elementSize := int(getElementSize(directive))
	var bytes []byte

	// Handle strings for .dc.b
	if elementSize == 1 && strings.Contains(values, "\"") {
		inQuote := false
		for _, c := range values {
			if c == '"' {
				inQuote = !inQuote
			} else if inQuote {
				bytes = append(bytes, byte(c))
			}
		}
	} else {
		// Handle numeric data
		for _, p := range strings.Split(values, ",") {
			trimmed := strings.TrimSpace(p)
			if trimmed == "" {
				continue
			}
			val, err := parseConstant(trimmed, asm)
			if err != nil {
				return nil, fmt.Errorf("invalid constant '%s': %v", trimmed, err)
			}
			switch elementSize {
			case 1:
				bytes = append(bytes, byte(val))
			case 2:
				bytes = append(bytes, byte(val>>8), byte(val))
			case 4:
				bytes = append(bytes, byte(val>>24), byte(val>>16), byte(val>>8), byte(val))
			}
		}
	}

	// Ensure even byte count for word alignment
	if len(bytes)%2 != 0 {
		bytes = append(bytes, 0)
	}

	return cpu.BytesToWords(bytes), nil
}

// getElementSize returns element size in bytes for data-storage directives.
func getElementSize(directive string) uint32 {
	switch directive {
	case ".dc.b", ".ds.b":
		return 1
	case ".dc.w", ".ds.w":
		return 2
	case ".dc.l", ".ds.l":
		return 4
	default:
		return 1 // fallback (should never happen)
	}
}
