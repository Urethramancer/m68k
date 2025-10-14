package assembler

import (
	"fmt"
	"strings"
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
		return asm.calculateDcSize(dir, values)

	case "ds.b", "ds.w", "ds.l":
		if len(n.Parts) != 2 {
			return 0, fmt.Errorf("%s requires a single count argument", n.Parts[0])
		}
		count, err := asm.parseConstant(n.Parts[1])
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
// Returns a byte slice, as directives like DC.B are not always word-aligned.
func (asm *Assembler) generateDirectiveCode(n *Node) ([]byte, error) {
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
		return asm.assembleDc(dir, values)

	case "ds.b", "ds.w", "ds.l":
		if len(n.Parts) != 2 {
			return nil, fmt.Errorf("%s requires a single count argument", n.Parts[0])
		}
		count, err := asm.parseConstant(n.Parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid count for %s: %v", n.Parts[0], err)
		}
		elementSize := getElementSize(dir)
		byteSize := uint32(count) * elementSize
		return make([]byte, byteSize), nil

	default:
		return nil, fmt.Errorf("unknown directive: %s", n.Parts[0])
	}
}

// calculateDcSize determines the byte size of a .dc directive's data.
func (asm *Assembler) calculateDcSize(directive, values string) (uint32, error) {
	elementSize := getElementSize(directive)
	var size uint32

	tokens := splitDcValues(values)
	for _, tok := range tokens {
		if tok.Quoted {
			size += uint32(len(tok.Value))
		} else {
			// It's a numeric value. It contributes `elementSize` bytes.
			size += elementSize
		}
	}

	return size, nil
}

// directives.go

// assembleDc generates machine data for DC.B/DC.W/DC.L.
func (asm *Assembler) assembleDc(directive, values string) ([]byte, error) {
	elementSize := int(getElementSize(directive))
	var bytesBuf []byte

	tokens := splitDcValues(values)
	for _, tok := range tokens {
		if tok.Quoted {
			// Append string bytes in natural order
			bytesBuf = append(bytesBuf, []byte(tok.Value)...)
			continue
		}

		val, err := asm.parseConstant(tok.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid constant '%s': %v", tok.Value, err)
		}

		switch elementSize {
		case 1:
			bytesBuf = append(bytesBuf, byte(val))
		case 2:
			bytesBuf = append(bytesBuf, byte(val>>8), byte(val))
		case 4:
			bytesBuf = append(bytesBuf,
				byte(val>>24), byte(val>>16),
				byte(val>>8), byte(val))
		}
	}

	return bytesBuf, nil
}

// splitDcValues handles mixed quoted strings and numbers correctly.
type dcToken struct {
	Value  string
	Quoted bool
}

func splitDcValues(s string) []dcToken {
	var tokens []dcToken
	inQuote := false
	var quoteChar rune
	var cur strings.Builder
	for _, c := range s {
		switch c {
		case '\'', '"':
			if inQuote && rune(c) == quoteChar {
				tokens = append(tokens, dcToken{Value: cur.String(), Quoted: true})
				cur.Reset()
				inQuote = false
			} else if !inQuote {
				inQuote = true
				quoteChar = rune(c)
			} else {
				cur.WriteRune(c)
			}
		case ',':
			if !inQuote {
				if val := strings.TrimSpace(cur.String()); val != "" {
					tokens = append(tokens, dcToken{Value: val})
				}
				cur.Reset()
			} else {
				cur.WriteRune(c)
			}
		default:
			cur.WriteRune(c)
		}
	}
	if val := strings.TrimSpace(cur.String()); val != "" && !inQuote {
		tokens = append(tokens, dcToken{Value: val})
	}
	return tokens
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
