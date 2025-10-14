package cpu

import "fmt"

// opADD handles the ADD instruction.
// This function calculates the result and then calls a helper to set the flags.
func (c *CPU) opADD(inst *DecodedInstruction) error {
	// Determine the direction of the operation from the opcode.
	// Bit 8 (opmode bit) determines direction:
	// 0: Dn = Dn + <ea>
	// 1: <ea> = <ea> + Dn
	var src, dst uint32
	var err error

	// Fetch operands based on direction
	if inst.OpMode&0x100 == 0 { // Direction is to Dn
		dst, err = c.GetOperand(ModeData, inst.DstReg, inst.Size)
		if err != nil {
			return fmt.Errorf("ADD failed to get destination operand: %w", err)
		}
		src, err = c.GetOperand(inst.SrcMode, inst.SrcReg, inst.Size)
		if err != nil {
			return fmt.Errorf("ADD failed to get source operand: %w", err)
		}
	} else { // Direction is to <ea>
		dst, err = c.GetOperand(inst.SrcMode, inst.SrcReg, inst.Size)
		if err != nil {
			return fmt.Errorf("ADD failed to get destination operand: %w", err)
		}
		src, err = c.GetOperand(ModeData, inst.DstReg, inst.Size)
		if err != nil {
			return fmt.Errorf("ADD failed to get source operand: %w", err)
		}
	}

	// Perform the addition and set flags
	result := dst + src
	c.setFlagsArith(src, dst, result, inst.Size)

	// Write the result back to the correct destination
	if inst.OpMode&0x100 == 0 { // Direction is to Dn
		err = c.PutOperand(ModeData, inst.DstReg, inst.Size, result)
	} else { // Direction is to <ea>
		err = c.PutOperand(inst.SrcMode, inst.SrcReg, inst.Size, result)
	}
	if err != nil {
		return fmt.Errorf("ADD failed to put result: %w", err)
	}

	return nil
}

// opADDQ handles the ADDQ (Add Quick) instruction.
// Format: 0101 <data> 0 <size> <ea>
func (c *CPU) opADDQ(inst *DecodedInstruction) error {
	// The immediate value (1-8) was stored in SrcReg by the decoder.
	src := uint32(inst.SrcReg)

	dst, err := c.GetOperand(inst.DstMode, inst.DstReg, inst.Size)
	if err != nil {
		return fmt.Errorf("ADDQ failed to get destination operand: %w", err)
	}

	result := dst + src
	c.setFlagsArith(src, dst, result, inst.Size)

	err = c.PutOperand(inst.DstMode, inst.DstReg, inst.Size, result)
	if err != nil {
		return fmt.Errorf("ADDQ failed to put result: %w", err)
	}
	return nil
}
