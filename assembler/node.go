package assembler

// NodeType defines the type of an assembly node.
type NodeType int

const (
	// NodeInstruction type.
	NodeInstruction NodeType = iota
	// NodeLabel type.
	NodeLabel
	// NodeDirective type.
	NodeDirective
)

// Node represents one parsed element from the assembly source.
type Node struct {
	Type     NodeType
	Label    string
	Mnemonic Mnemonic
	Operands []Operand
	Parts    []string
	Size     uint32 // Still used to track size between passes
}
