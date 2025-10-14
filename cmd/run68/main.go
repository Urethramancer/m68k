package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Urethramancer/m68k/assembler"
	"github.com/Urethramancer/m68k/vm"
)

var (
	// Configuration flags
	loadAddress = flag.Uint64("load", 0x0000, "Load address for binary files (hex).")
	pcAddress   = flag.Uint64("pc", 0, "Initial program counter (hex), defaults to load address.")
	maxCycles   = flag.Int("cycles", 1000000, "Maximum number of instructions to execute.")

	// Register value flags
	regD [8]string
	regA [8]string
)

func init() {
	// Dynamically create flags for all 16 general-purpose registers
	for i := 0; i < 8; i++ {
		flag.StringVar(&regD[i], fmt.Sprintf("d%d", i), "", "Set initial value for data register D (hex).")
		flag.StringVar(&regA[i], fmt.Sprintf("a%d", i), "", "Set initial value for address register A (hex).")
	}
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	// We need exactly one non-flag argument: the filename
	if flag.NArg() != 1 {
		log.Println("Usage: run68 [options] <filename>")
		flag.PrintDefaults()
		os.Exit(1)
	}
	filename := flag.Arg(0)

	v := vm.New(16*1024*1024, 1024) // 16MB RAM

	// Set registers from command-line flags
	err := setRegisters(v)
	if err != nil {
		log.Fatalf("Error setting registers: %v", err)
	}

	// Load code based on file extension
	var code []byte
	var startAddress uint32
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".asm", ".s":
		log.Printf("Assembling %s...", filename)
		sourceBytes, err := os.ReadFile(filename)
		if err != nil {
			log.Fatalf("Couldn't read source file: %v", err)
		}
		asm := assembler.New()
		// For assembly, the ORG directive determines the load address.
		// We pass 0 and let the assembler figure it out.
		code, err = asm.Assemble(string(sourceBytes), 0)
		if err != nil {
			log.Fatalf("Assembly failed: %v", err)
		}
		// The assembler sets the PC to the ORG address.
		startAddress = asm.BaseAddress()
		v.LoadCode(startAddress, code)

	case ".bin", ".m68":
		log.Printf("Loading binary %s...", filename)
		code, err = os.ReadFile(filename)
		if err != nil {
			log.Fatalf("Couldn't read binary file: %v", err)
		}
		startAddress = uint32(*loadAddress)
		v.LoadCode(startAddress, code)

	default:
		log.Fatalf("Unknown file extension: %s. Use .asm, .s, .bin, or .m68", ext)
	}

	// Set program counter, overriding assembler ORG if specified
	if *pcAddress != 0 {
		v.CPU.PC = uint32(*pcAddress)
	} else {
		v.CPU.PC = startAddress
	}

	log.Printf("Loaded %d bytes. Execution starts at 0x%08X", len(code), v.CPU.PC)
	log.Println("\n--- CPU State Before Execution ---")
	v.DumpRegisters()

	// --- Execution Loop ---
	v.CPU.Running = true
	var executedCycles int
	for executedCycles = 0; executedCycles < *maxCycles; executedCycles++ {
		if !v.CPU.Running {
			break
		}
		err := v.CPU.Execute()
		if err != nil {
			log.Printf("\n--- CPU State at Failure ---")
			v.DumpRegisters()
			log.Fatalf("\nCPU execution failed after %d instructions: %s at 0x%08X",
				executedCycles+1, err, v.CPU.PC-2)
		}
	}

	log.Println("\n--- CPU State After Execution ---")
	v.DumpRegisters()

	if executedCycles >= *maxCycles {
		log.Printf("\nExecution finished: Maximum cycle count (%d) reached.", *maxCycles)
	} else {
		log.Printf("\nExecution finished successfully after %d instructions.", executedCycles)
	}
}

// setRegisters parses the string flags and sets CPU registers.
func setRegisters(v *vm.VM) error {
	for i := 0; i < 8; i++ {
		if regD[i] != "" {
			val, err := strconv.ParseUint(strings.TrimPrefix(regD[i], "0x"), 16, 32)
			if err != nil {
				return fmt.Errorf("invalid value for d%d: %w", i, err)
			}
			v.CPU.D[i] = uint32(val)
		}
		if regA[i] != "" {
			val, err := strconv.ParseUint(strings.TrimPrefix(regA[i], "0x"), 16, 32)
			if err != nil {
				return fmt.Errorf("invalid value for a%d: %w", i, err)
			}
			v.CPU.A[i] = uint32(val)
		}
	}
	return nil
}
