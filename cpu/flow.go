package cpu

// opRTS handles the RTS (Return from Subroutine) instruction.
// Format: 0100 1110 0111 0101 (4E75)
func (c *CPU) opRTS(inst *DecodedInstruction) error {
	// Get the current stack pointer (A7).
	sp := c.A[7]
	// Read the return address (a long word) from the stack.
	returnAddr := c.ReadU32(sp)
	// Pop the address off the stack by incrementing the stack pointer.
	c.A[7] += 4
	// Set the Program Counter to the return address.
	c.PC = returnAddr
	return nil
}
