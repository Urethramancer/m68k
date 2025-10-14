package cpu

import "fmt"

// opMOVEQ handles the MOVEQ (Move Quick) instruction.
// Format: 0111 <reg> 0 <8-bit data>
func (c *CPU) opMOVEQ(inst *DecodedInstruction) error {
	// The immediate data was stored in SrcReg by the decoder.
	data := int8(inst.SrcReg & 0xFF)
	value := uint32(int32(data))

	c.D[inst.DstReg] = value

	// Update Status Register
	c.SR &^= (SRV | SRC)
	c.setNZ(value, SizeLong)
	return nil
}

// opMOVEA handles the MOVEA (Move Address) instruction.
func (c *CPU) opMOVEA(inst *DecodedInstruction) error {
	// MOVEA only supports word and long sizes.
	if inst.Size == SizeByte {
		return fmt.Errorf("invalid size for MOVEA: .B")
	}

	value, err := c.GetOperand(inst.SrcMode, inst.SrcReg, inst.Size)
	if err != nil {
		return fmt.Errorf("MOVEA failed to get source operand: %w", err)
	}

	// If the size is word, the source is sign-extended to 32 bits.
	if inst.Size == SizeWord {
		value = uint32(int32(int16(value)))
	}

	c.A[inst.DstReg] = value

	// MOVEA does not affect condition codes.
	return nil
}

// opMOVE handles the general MOVE instruction.
func (c *CPU) opMOVE(inst *DecodedInstruction) error {
	value, err := c.GetOperand(inst.SrcMode, inst.SrcReg, inst.Size)
	if err != nil {
		return fmt.Errorf("MOVE failed to get source operand: %w", err)
	}

	// Corrected the order of arguments in the call to PutOperand.
	err = c.PutOperand(inst.DstMode, inst.DstReg, inst.Size, value)
	if err != nil {
		return fmt.Errorf("MOVE failed to put destination operand: %w", err)
	}

	// Update Status Register
	c.SR &^= (SRV | SRC)
	c.setNZ(value, inst.Size)
	return nil
}
