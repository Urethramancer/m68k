package cpu

// Size defines the data size for an instruction's operation.
type Size int

const (
	// SizeInvalid is the zero value, indicating no size suffix was provided.
	SizeInvalid Size = iota
	// SizeByte represents 8-bit data size.
	SizeByte
	// SizeWord represents 16-bit data size.
	SizeWord
	// SizeLong represents 32-bit data size.
	SizeLong
	// SizeShort is used for branch instructions (displacement).
	SizeShort
)

// Opcodes for various instructions.
const (
	// Logical and Bit Manipulation Instructions
	OPAND                 = 0xC000 // AND
	OPOR                  = 0x8000 // OR
	OPEOR                 = 0xB100 // EOR
	OPANDI                = 0x0200 // ANDI
	OPORI                 = 0x0000 // ORI
	OPEORI                = 0x0A00 // EORI
	OPANDItoCCR           = 0x023C // ANDI to CCR
	OPORItoCCR            = 0x003C // ORI to CCR
	OPEORItoCCR           = 0x0A3C // EORI to CCR
	OPANDItoSR            = 0x027C // ANDI to SR (privileged)
	OPORItoSR             = 0x007C // ORI to SR (privileged)
	OPEORItoSR            = 0x0A7C // EORI to SR (privileged)
	OPNOT                 = 0x4600 // NOT
	OPCLR                 = 0x4200 // CLR
	OPTST                 = 0x4A00 // TST
	OPNEG                 = 0x4400 // NEG
	OPNEGX                = 0x4000 // NEGX
	OPNBCD                = 0x4800 // NBCD
	OPEXT                 = 0x4800 // EXT
	OPSWAP                = 0x4840 // SWAP
	OPBCHG                = 0x0840 // BCHG
	OPBCLR                = 0x0880 // BCLR
	OPBSET                = 0x08C0 // BSET
	OPBTST                = 0x0800 // BTST
	OPBitManipulationBase = 0x0100 // Base for dynamic BTST, BSET, etc.

	// Arithmetic Instructions
	OPADD  = 0xD000 // ADD
	OPADDQ = 0x5000 // ADDQ
	OPADDA = 0xD000 // ADDA (Base, size bits added separately)
	OPADDI = 0x0600 // ADDI
	OPADDX = 0xD100 // ADDX
	OPSUB  = 0x9000 // SUB
	OPSUBQ = 0x5100 // SUBQ
	OPSUBA = 0x9000 // SUBA (Base, size bits added separately)
	OPSUBI = 0x0400 // SUBI
	OPSUBX = 0x9100 // SUBX
	OPMULS = 0xC1C0 // MULS
	OPMULU = 0xC0C0 // MULU
	OPDIVS = 0x81C0 // DIVS
	OPDIVU = 0x80C0 // DIVU

	// Comparison Instructions
	OPCMP  = 0xB000 // CMP
	OPCMPI = 0x0C00 // CMPI
	OPCMPA = 0xB000 // CMPA (Base, size bits added separately)
	OPCHK  = 0x4180 // CHK

	// Shift and Rotate Instructions
	OPShiftRotateBase = 0xE000 // Base for all shifts and rotates
	OPASR             = 0xE000 // ASR
	OPASL             = 0x100  // ASL
	OPLSR             = 0xE008 // LSR
	OPLSL             = 0xE108 // LSL
	OPROR             = 0xE018 // ROR
	OPROL             = 0xE118 // ROL
	OPROXR            = 0xE020 // ROXR
	OPROXL            = 0xE120 // ROXL

	// Move Instructions
	OPMOVE        = 0x0000 // MOVE (placeholder, size bits are added)
	OPMOVEA       = 0x0040 // MOVEA (placeholder, size bits are added)
	OPMOVEQ       = 0x7000 // MOVEQ
	OPMOVEM       = 0x4880 // MOVEM (base)
	OPMOVEP       = 0x0008 // MOVEP (base)
	OPMOVEFromSR  = 0x40C0 // MOVE from SR
	OPMOVEToSR    = 0x46C0 // MOVE to SR (privileged)
	OPMOVEToCCR   = 0x44C0 // MOVE to CCR
	OPMOVEFromUSP = 0x4E68 // MOVE from USP
	OPMOVEToUSP   = 0x4E60 // MOVE to USP

	// Address Calculation and Stack Instructions
	OPPEA  = 0x4840 // PEA
	OPLEA  = 0x41C0 // LEA (Base, register is OR'd)
	OPLINK = 0x4E50 // LINK
	OPUNLK = 0x4E58 // UNLK

	// Control Instructions
	OPTRAP    = 0x4E40 // TRAP
	OPTRAPV   = 0x4E76 // TRAPV
	OPRTE     = 0x4E73 // RTE
	OPSTOP    = 0x4E72 // STOP
	OPRESET   = 0x4E70 // RESET
	OPNOP     = 0x4E71 // NOP
	OPILLEGAL = 0x4AFC // ILLEGAL
	OPRTS     = 0x4E75 // RTS
	OPRTR     = 0x4E77 // RTR
	OPTAS     = 0x4AC0 // TAS
	OPEXG     = 0xC100 // EXG (base)

	// Conditional Instructions
	OPScc  = 0x50C0 // Scc (base, condition code OR'd)
	OPDBcc = 0x50C8 // DBcc (base, condition code OR'd)

	// Branch Instructions (base values, condition codes are OR'd)
	OPBRA = 0x6000 // Branch Always
	OPBSR = 0x6100 // Branch to Subroutine
	OPBHI = 0x6200 // Branch if Higher
	OPBLS = 0x6300 // Branch if Lower or Same
	OPBCC = 0x6400 // Branch if Carry Clear (Higher or Same)
	OPBCS = 0x6500 // Branch if Carry Set (Lower)
	OPBNE = 0x6600 // Branch if Not Equal
	OPBEQ = 0x6700 // Branch if Equal
	OPBVC = 0x6800 // Branch if Overflow Clear
	OPBVS = 0x6900 // Branch if Overflow Set
	OPBPL = 0x6A00 // Branch if Plus (Positive)
	OPBMI = 0x6B00 // Branch if Minus (Negative)
	OPBGE = 0x6C00 // Branch if Greater or Equal
	OPBLT = 0x6D00 // Branch if Less Than
	OPBGT = 0x6E00 // Branch if Greater Than
	OPBLE = 0x6F00 // Branch if Less or Equal

	// Jump and Subroutine Instructions
	OPJMP = 0x4EC0 // JMP
	OPJSR = 0x4E80 // JSR
)

// BranchOpcodes maps branch mnemonics to their base opcodes.
var BranchOpcodes = map[string]uint16{
	"bra": OPBRA,
	"bsr": OPBSR,
	"bhi": OPBHI,
	"bls": OPBLS,
	"bcc": OPBCC,
	"bcs": OPBCS,
	"bne": OPBNE,
	"beq": OPBEQ,
	"bvc": OPBVC,
	"bvs": OPBVS,
	"bpl": OPBPL,
	"bmi": OPBMI,
	"bge": OPBGE,
	"blt": OPBLT,
	"bgt": OPBGT,
	"ble": OPBLE,
}

// ConditionCodes maps condition mnemonics to their 4-bit codes.
var ConditionCodes = map[string]uint16{
	"t":  0x0, // true
	"f":  0x1, // false
	"hi": 0x2, // high
	"ls": 0x3, // low or same
	"cc": 0x4, // carry clear
	"cs": 0x5, // carry set
	"ne": 0x6, // not equal
	"eq": 0x7, // equal
	"vc": 0x8, // overflow clear
	"vs": 0x9, // overflow set
	"pl": 0xA, // plus
	"mi": 0xB, // minus
	"ge": 0xC, // greater or equal
	"lt": 0xD, // less than
	"gt": 0xE, // greater than
	"le": 0xF, // less or equal
}
