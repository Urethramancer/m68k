package cpu

// opTRAP handles the TRAP instruction.
// Format: 0100 1110 0100 <vector>
func (c *CPU) opTRAP(inst *DecodedInstruction) error {
	// The trap vector is stored in the lower 4 bits of the opcode.
	// The decoder will place it in the DstReg field for us.
	vector := inst.DstReg
	println("TRAP instruction invoked with vector:", vector)
	// We'll use TRAP #15 as a special instruction to halt the VM.
	if vector == 15 {
		c.Running = false
	}

	// In a full OS, other TRAP vectors would trigger exceptions
	// and call system routines. For now, we just halt on #15.
	return nil
}
