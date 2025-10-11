package assembler

import (
	"fmt"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

// getDirectiveSize calculates the byte size of a directive for the sizing pass.
//
// Note: pc is passed so .even can be sized correctly.
func (asm *Assembler) getDirectiveSize(n *Node, pc uint32) (uint32, error) {
	// normalize directive: lowercase, drop an optional leading dot
	raw := strings.ToLower(n.Parts[0])
	dir := strings.TrimPrefix(raw, ".")

	switch dir {
	case "org", "equ":
		return 0, nil

	case "even":
		// if current pc is odd, .even emits one padding byte
		if pc%2 != 0 {
			return 1, nil
		}
		return 0, nil

	case "dc.b", "dc.w", "dc.l":
		if len(n.Parts) < 2 {
			return 0, fmt.Errorf("%s requires at least one value", n.Parts[0])
		}
		values := strings.Join(n.Parts[1:], " ")
		return calculateDcSize(dir, values, asm)

	case "ds.b", "ds.w", "ds.l":
		if len(n.Parts) != 2 {
			return 0, fmt.Errorf("%s requires a single count argument", n.Parts[0])
		}
		count, err := parseConstant(n.Parts[1], asm)
		if err != nil {
			return 0, fmt.Errorf("invalid count for %s: %v", n.Parts[0], err)
		}
		elementSize := getElementSize(dir)
		return uint32(count) * elementSize, nil

	default:
		return 0, fmt.Errorf("unknown directive: %s", n.Parts[0])
	}
}

// generateDirectiveCode generates the binary data for assembler directives.
// Returns 16-bit words (big-endian). .even and .org/.equ return nil (handled by assemble loop).
func (asm *Assembler) generateDirectiveCode(n *Node) ([]uint16, error) {
	// Normalize directive name once: lowercase, no leading dot.
	raw := strings.ToLower(n.Parts[0])
	dir := strings.TrimPrefix(raw, ".")

	switch dir {
	case "org", "equ":
		return nil, nil

	case "even":
		// .even is handled in the assembly loop so we return nil here.
		return nil, nil

	case "dc.b", "dc.w", "dc.l":
		if len(n.Parts) < 2 {
			return nil, fmt.Errorf("%s requires at least one value", n.Parts[0])
		}
		values := strings.Join(n.Parts[1:], " ")
		// pass the normalized directive (e.g. "dc.b") and the assembler for symbols.
		return assembleDc(dir, values, asm)

	case "ds.b", "ds.w", "ds.l":
		if len(n.Parts) != 2 {
			return nil, fmt.Errorf("%s requires a single count argument", n.Parts[0])
		}
		count, err := parseConstant(n.Parts[1], asm)
		if err != nil {
			return nil, fmt.Errorf("invalid count for %s: %v", n.Parts[0], err)
		}
		elementSize := getElementSize(dir)
		byteSize := uint32(count) * elementSize
		wordSize := (byteSize + 1) / 2
		return make([]uint16, wordSize), nil

	default:
		return nil, fmt.Errorf("unknown directive: %s", n.Parts[0])
	}
}

// calculateDcSize determines the byte size of a .dc directive's data.
func calculateDcSize(directive, values string, asm *Assembler) (uint32, error) {
	elementSize := getElementSize(directive)
	var size uint32

	// string handling for .dc.b
	if elementSize == 1 && (strings.Contains(values, "\"") || strings.Contains(values, "'")) {
		inQuote := false
		var quoteChar rune
		for _, c := range values {
			switch c {
			case '\'', '"':
				if inQuote && c == quoteChar {
					inQuote = false
				} else if !inQuote {
					inQuote = true
					quoteChar = c
				}
			default:
				if inQuote {
					size++
				}
			}
		}
	} else {
		for _, p := range strings.Split(values, ",") {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				size += elementSize
			}
		}
	}

	// align to word boundary
	if size%2 != 0 {
		size++
	}
	return size, nil
}

// assembleDc generates machine data for DC.B/DC.W/DC.L.
// It always returns words in Motorola big-endian order, regardless of host endianness.
func assembleDc(directive, values string, asm *Assembler) ([]uint16, error) {
	elementSize := int(getElementSize(directive))
	var bytesBuf []byte

	// --- 1. Parse all the values into bytes in Motorola order ---
	// Allow mixing strings and numeric constants for DC.B
	if elementSize == 1 && (strings.Contains(values, "'") || strings.Contains(values, "\"")) {
		inQuote := false
		var quoteChar rune
		token := ""
		for _, c := range values {
			switch c {
			case '\'', '"':
				if inQuote && c == quoteChar {
					for i := 0; i < len(token); i++ {
						bytesBuf = append(bytesBuf, token[i])
					}
					token = ""
					inQuote = false
				} else if !inQuote {
					inQuote = true
					quoteChar = c
				} else {
					token += string(c)
				}
			case ',':
				if !inQuote {
					token = strings.TrimSpace(token)
					if token != "" {
						val, err := parseConstant(token, asm)
						if err != nil {
							return nil, fmt.Errorf("invalid constant '%s': %v", token, err)
						}
						bytesBuf = append(bytesBuf, byte(val))
					}
					token = ""
				} else {
					token += string(c)
				}
			default:
				token += string(c)
			}
		}
		if token != "" && !inQuote {
			token = strings.TrimSpace(token)
			if token != "" {
				val, err := parseConstant(token, asm)
				if err != nil {
					return nil, fmt.Errorf("invalid constant '%s': %v", token, err)
				}
				bytesBuf = append(bytesBuf, byte(val))
			}
		}
	} else {
		// Numeric only
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
				bytesBuf = append(bytesBuf, byte(val))
			case 2:
				// Motorola order: high byte first
				bytesBuf = append(bytesBuf, byte(val>>8), byte(val))
			case 4:
				// Motorola order: highest byte first
				bytesBuf = append(bytesBuf, byte(val>>24), byte(val>>16), byte(val>>8), byte(val))
			}
		}
	}

	// --- 2. Always align to even length ---
	if len(bytesBuf)%2 != 0 {
		bytesBuf = append(bytesBuf, 0)
	}

	// --- 3. If host is little-endian, swap pairs before converting ---
	if cpu.IsLittleEndianHost() {
		for i := 0; i < len(bytesBuf); i += 2 {
			if i+1 < len(bytesBuf) {
				bytesBuf[i], bytesBuf[i+1] = bytesBuf[i+1], bytesBuf[i]
			}
		}
	}

	// --- 4. Convert bytes to words in big-endian order ---
	return cpu.BytesToWords(bytesBuf), nil
}

// getElementSize returns element size in bytes for data-storage directives.
func getElementSize(directive string) uint32 {
	// directive is normalized without leading dot (e.g. "dc.b")
	switch strings.ToLower(strings.TrimPrefix(directive, ".")) {
	case "dc.b", "ds.b", "dcb", "dsb":
		return 1
	case "dc.w", "ds.w", "dcw", "dsw":
		return 2
	case "dc.l", "ds.l", "dcl", "dsl":
		return 4
	default:
		return 1
	}
}
