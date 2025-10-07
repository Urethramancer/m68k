package cpu

// CPU memory and registers.
type CPU struct {
	// D is for data registers.
	D [8]uint32
	// A is for address registers. A7 is the current stack pointer.
	A [8]uint32
	// PC is the program counter.
	PC uint32
	// USP is the user stack pointer.
	USP uint32
	// SSP is the supervisor stack pointer.
	SSP uint32
	// SR is the status register.
	SR uint16
	// ISP is the interrupt stack pointer.
	ISP uint32

	// Memory
	Mem []byte
	// Cache for instructions.
	ICache map[uint32]uint32

	// Cycles count.
	Cycles int32
	// Running or not.
	Running bool
}

// Status register flags.
const (
	// SR_C is carry
	SRC = 1 << 0
	// SR_V is overflow
	SRV = 1 << 1
	// SR_Z is zero
	SRZ = 1 << 2
	// SR_N is negative
	SRN = 1 << 3
	// SR_X is extend
	SRX = 1 << 4
	// SR_I0 is interrupt level 0
	SRI0 = 1 << 8
	// SR_I1 is interrupt level 1
	SRI1 = 1 << 9
	// SR_I2 is interrupt level 2
	SRI2 = 1 << 10
	// SR_S is supervisor state
	SRS = 1 << 13
	// SR_T is trace mode
	SRT = 1 << 15
)

// New creates a new CPU instance with given memory size.
func New(memsize, cachesize int) *CPU {
	cpu := &CPU{
		Mem:     make([]byte, memsize),
		ICache:  make(map[uint32]uint32, cachesize),
		Running: false,
	}
	return cpu
}

// Execute a single instruction.
func (c *CPU) Execute() error {
	if !c.Running {
		return nil
	}

	// Placeholder
	return nil
}

// LoadCode to specified address.
func (c *CPU) LoadCode(addr uint32, code []byte) {
	copy(c.Mem[addr:], code)
	c.PC = addr
}
