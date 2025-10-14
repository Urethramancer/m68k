package cpu

import "fmt"

// Execute fetches, decodes, and executes a single instruction.
func (c *CPU) Execute() error {
	if !c.Running {
		return nil
	}

	// Fetch
	opcode := c.ReadU16(c.PC)
	c.PC += 2

	// Decode
	inst, err := c.Decode(opcode)
	if err != nil {
		return fmt.Errorf("decode failed: %w", err)
	}

	if inst.Handler == nil {
		return fmt.Errorf("no handler for opcode %04X", opcode)
	}

	// Execute
	err = inst.Handler(c, inst)
	if err != nil {
		return fmt.Errorf("execution failed for opcode %04X: %w", opcode, err)
	}

	return nil
}
