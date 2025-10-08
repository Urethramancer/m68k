package cpu

// Addressing mode constants (3-bit mode field + 3-bit register field)
const (
	// 000 — Data Register Direct: Dn
	ModeData uint16 = 0

	// 001 — Address Register Direct: An
	ModeAddr uint16 = 1

	// 010 — Address Register Indirect: (An)
	ModeAddrInd uint16 = 2

	// 011 — Address Register Indirect with Postincrement: (An)+
	ModeAddrPostInc uint16 = 3

	// 100 — Address Register Indirect with Predecrement: -(An)
	ModeAddrPreDec uint16 = 4

	// 101 — Address Register Indirect with Displacement: (d16,An)
	ModeAddrDisp uint16 = 5

	// 110 — Address Register Indirect with Index: (d8,An,Xn)
	ModeAddrIndex uint16 = 6

	// 111 — Miscellaneous / “other” addressing modes
	ModeOther uint16 = 7
)

// Submodes for ModeOther (register field = 3 bits)
const (
	// 000 — Absolute short address: (xxx).W
	RegAbsShort uint16 = 0

	// 001 — Absolute long address: (xxx).L
	RegAbsLong uint16 = 1

	// 010 — Program counter with displacement: (d16,PC)
	RegPCDisp uint16 = 2

	// 011 — Program counter with index: (d8,PC,Xn)
	RegPCIndex uint16 = 3

	// 100 — Immediate: #<data>
	RegImmediate uint16 = 4
)

// Derived addressing modes (for clarity in assembler code)
const (
	ModeAbsShort   uint16 = 0 // (xxx).W
	ModeAbsLong    uint16 = 1 // (xxx).L
	ModePCRelative uint16 = 2 // (d16,PC)
	ModeImmediate  uint16 = 4 // #<data>
)

// Register numbers
const (
	// Data registers
	D0 = 0
	D1 = 1
	D2 = 2
	D3 = 3
	D4 = 4
	D5 = 5
	D6 = 6
	D7 = 7

	// Address registers
	A0 = 0
	A1 = 1
	A2 = 2
	A3 = 3
	A4 = 4
	A5 = 5
	A6 = 6
	A7 = 7 // stack pointer
)
