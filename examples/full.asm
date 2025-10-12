*****************************************************************
* MC68000 Demonstration Program
*
* This program demonstrates all 14 addressing modes and a
* wide variety of instructions for the Motorola 68000.
* It is intended for educational purposes.
*****************************************************************

	.org    $1000          ; Start program at address $1000

start:
*****************************************************************
* SECTION 1: Addressing Mode Demonstration
* Using move as a simple instruction to show how data is accessed.
*****************************************************************

* 1. Data Register Direct: D0-D7
	move.l  d1,d0                 ; Move content of D1 into D0

* 2. Address Register Direct: A0-A7
	movea.l a1,a0                 ; Move content of A1 into A0

* 3. Address Register Indirect: (An)
	lea     mydata,a0             ; A0 points to MyData
	move.w  (a0),d0               ; D0 = the word at address in A0

* 4. Address Register Indirect with Post-increment: (An)+
	move.w  (a0)+,d1              ; D1 = word at A0, then A0 = A0 + 2

* 5. Address Register Indirect with Pre-decrement: -(An)
	move.w  -(a0),d2              ; A0 = A0 - 2, then D2 = word at new A0

* 6. Address Register Indirect with Displacement: d16(An)
	move.b  4(a0),d3              ; D3 = the byte at (A0 + 4)

* 7. Address Register Indirect with Index and Displacement: d8(An, Xn)
	move.w  2(a0,d1.w),d4         ; D4 = word at (A0 + D1 + 2)

* 8. Absolute Short Addressing: (xxx).W
	move.w  $2000,d5              ; D5 = word at address $2000

* 9. Absolute Long Addressing: (xxx).L
	move.l  mydata,d6             ; D6 = long word at the address of MyData

* 10. Program Counter with Displacement: d16(PC)
	move.w  pcrelative(pc),d7     ; D7 = word at (PC + offset to PCRelative)

* 11. Program Counter with Index and Displacement: d8(PC, Xn)
	lea     pcrelative(pc),a1     ; To do this, must get PC value first
	move.b  0(a1,d1.w),d0         ; D0 = byte at (PC_relative_addr + D1)

* 12. Immediate Data: #xxx
	move.l  #$deadbeef,d1         ; Load D1 with the value DEADBEEF

* 13. Quick Immediate: Implied in certain instructions
	moveq   #5,d2                 ; Move immediate value 5 (fastest way for small values)
	addq.b  #1,d2                 ; Add 1 to D2 (faster than ADD #1,D2)

* 14. Status Register (CCR and SR): CCR, SR
	move    #%00000101,ccr        ; Set Carry and Zero flags
	move    sr,d0                 ; Copy Status Register to D0 (privileged in user mode)

*****************************************************************
* SECTION 2: Instruction Variety Demonstration
*****************************************************************

* Arithmetic Instructions
	add.l   d1,d0                 ; D0 = D0 + D1
	subi.w  #100,d2               ; D2 = D2 - 100
	mulu.w  d2,d3                 ; D3 = D3 * D2 (unsigned)
	muls.w  d2,d4                 ; D4 = D4 * D2 (signed)
	divu.w  d2,d5                 ; D5 = D5 / D2 (unsigned)
	neg.w   d2                    ; D2 = -D2 (2's complement)
	ext.w   d0                    ; Extend sign bit of D0.b to D0.w
	ext.l   d0                    ; Extend sign bit of D0.w to D0.l

* Logical and Shift Instructions
	andi.b  #%11110000,d0         ; Bitwise AND immediate
	or.l    d1,d0                 ; Bitwise OR
	eor.w   d2,d2                 ; Bitwise XOR (common way to clear a register to 0)
	not.b   d3                    ; Bitwise NOT (1's complement)
	asl.w   #4,d1                 ; Arithmetic Shift Left (4 bits)
	lsr.l   #8,d0                 ; Logical Shift Right (8 bits)
	rol.b   #1,d3                 ; Rotate Left (1 bit)
	swap    d4                    ; Swap high and low words of D4

* Data Movement Instructions
	lea     mydata,a2             ; Load Effective Address of MyData into A2
	pea     mydata(pc)            ; Push Effective Address onto stack
	exg     d1,a1                 ; Exchange contents of D1 and A1
	clr.l   d7                    ; Clear D7 to zero
	movem.l d0-d3/a0-a2,-(sp)     ; Move multiple registers to the stack
	movem.l (sp)+,d0-d3/a0-a2     ; Restore multiple registers from stack

* Branch and Program Control
	cmp.w   d1,d2                 ; Compare D1 and D2, setting flags
	beq.s   equallabel            ; Branch if Equal (short branch)
	bne     notequallabel         ; Branch if Not Equal (long branch)
	bra     alwaysbranch          ; Branch Always

equallabel:
	nop                           ; No
notequallabel:
	tst.w   d0                    ; Test D0 and set flags
alwaysbranch:
	jsr     MySubroutine          ; Jump to Subroutine
	rts                           ; Return from Subroutine

*****************************************************************
* SECTION 3: Subroutine and Program Exit
*****************************************************************

MySubroutine:
	tst.w   d0                    ; Test D0 and set flags
	rts                           ; Return from Subroutine

*****************************************************************
* SECTION 4: Data Section
*****************************************************************
.org	$2000

pcrelative:
	.dc.w	$abcd                 ; Data for PC-Relative addressing mode

mydata:
	.dc.l	$11223344             ; A long word
	.dc.w	$5566, $7788          ; Some words
	.dc.b	'H','e','l','l','o',0 ; A null-terminated string
