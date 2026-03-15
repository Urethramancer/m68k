package vm

import (
	"fmt"
	"strings"

	"github.com/Urethramancer/m68k/cpu"
)

// VM is a virtual machine to run 68k code in.
type VM struct {
	// CPU instance.
	CPU *cpu.CPU
	// Quit flag.
	Quit chan bool
}

// New creates a VM with a 68k CPU instance and specified amount of memory.
func New(memsize, cachesize int) *VM {
	return &VM{
		CPU:  cpu.New(memsize, cachesize),
		Quit: make(chan bool),
	}
}

// Start VM in a goroutine.
func (v *VM) Start() {
	v.CPU.Running = true
	go func() {
		for v.CPU.Running {
			err := v.CPU.Execute()
			if err != nil {
				v.Quit <- true
				return
			}
		}
		v.Quit <- true
	}()
}

// Pause the VM.
func (v *VM) Pause() {
	v.CPU.Running = false
}

// Resume the VM.
func (v *VM) Resume() {
	if !v.CPU.Running {
		v.Start()
	}
}

// Stop the VM.
func (v *VM) Stop() {
	v.CPU.Running = false
	<-v.Quit
}

// LoadCode to specified address.
func (v *VM) LoadCode(addr uint32, code []byte) {
	copy(v.CPU.Mem[addr:], code)
	v.CPU.PC = addr
}

// formatSR formats the Status Register bits into a human-readable string.
// T.S..I2I1I0..XNZVC (a '.' means the flag is not set)
func (v *VM) formatSR() string {
	sr := v.CPU.SR
	var b strings.Builder
	b.WriteString("[")
	if (sr & cpu.SRT) != 0 {
		b.WriteRune('T')
	} else {
		b.WriteRune('.')
	}
	if (sr & cpu.SRS) != 0 {
		b.WriteRune('S')
	} else {
		b.WriteRune('.')
	}
	b.WriteString("..")
	// Interrupt Mask
	b.WriteString(fmt.Sprintf("%d", (sr&cpu.SRI)>>8))
	b.WriteString("..")
	if (sr & cpu.SRX) != 0 {
		b.WriteRune('X')
	} else {
		b.WriteRune('.')
	}
	if (sr & cpu.SRN) != 0 {
		b.WriteRune('N')
	} else {
		b.WriteRune('.')
	}
	if (sr & cpu.SRZ) != 0 {
		b.WriteRune('Z')
	} else {
		b.WriteRune('.')
	}
	if (sr & cpu.SRV) != 0 {
		b.WriteRune('V')
	} else {
		b.WriteRune('.')
	}
	if (sr & cpu.SRC) != 0 {
		b.WriteRune('C')
	} else {
		b.WriteRune('.')
	}
	b.WriteString("]")
	return b.String()
}

// DumpRegisters prints the current state of the CPU registers to standard output.
// It provides a formatted view of data, address, and status registers.
func (v *VM) DumpRegisters() {
	c := v.CPU
	fmt.Println("---------------- CPU STATE ----------------")
	// Print Data Registers
	for i := 0; i < 8; i += 4 {
		fmt.Printf("D%d: %08X   D%d: %08X   D%d: %08X   D%d: %08X\n",
			i, c.D[i], i+1, c.D[i+1], i+2, c.D[i+2], i+3, c.D[i+3])
	}

	// Print Address Registers
	for i := 0; i < 8; i += 4 {
		fmt.Printf("A%d: %08X   A%d: %08X   A%d: %08X   A%d: %08X\n",
			i, c.A[i], i+1, c.A[i+1], i+2, c.A[i+2], i+3, c.A[i+3])
	}

	// Print Program Counter and Status Register
	fmt.Printf("PC: %08X   SR: %04X %s\n", c.PC, c.SR, v.formatSR())

	// Differentiate between the stack pointers for clarity
	if c.SR&cpu.SRS != 0 {
		// Supervisor mode
		fmt.Printf("SP: A7(SSP)=%08X USP=%08X\n", c.A[7], c.USP)
	} else {
		// User mode
		fmt.Printf("SP: A7(USP)=%08X SSP=%08X\n", c.A[7], c.SSP)
	}
	fmt.Println("-------------------------------------------")
}

// DumpMemory prints a hex dump of a memory region.
func (v *VM) DumpMemory(addr, count uint32) {
	fmt.Println("---------------- MEMORY DUMP ----------------")
	if addr+count > uint32(len(v.CPU.Mem)) {
		fmt.Printf("  Address range [0x%08X - 0x%08X] is out of bounds.\n", addr, addr+count)
		return
	}
	if count == 0 {
		return
	}

	for i := uint32(0); i < count; i += 16 {
		lineAddr := addr + i
		end := i + 16
		if lineAddr+16 > addr+count {
			end = count - i
		}
		data := v.CPU.Mem[lineAddr : lineAddr+end]

		hexPart := []string{}
		asciiPart := []rune{}
		for _, b := range data {
			hexPart = append(hexPart, fmt.Sprintf("%02X", b))
			if b >= 32 && b <= 126 {
				asciiPart = append(asciiPart, rune(b))
			} else {
				asciiPart = append(asciiPart, '.')
			}
		}

		hexStr := strings.Join(hexPart, " ")
		fmt.Printf("%08X: %-47s |%s|\n", lineAddr, hexStr, string(asciiPart))
	}
	fmt.Println("-------------------------------------------")
}
