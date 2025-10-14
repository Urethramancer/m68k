; Main testing program for the m68k VM.
; This program calculates 12 + 7, resulting in 19,
; and stores the result at memory address $2000.

org     $1000           ; Set the origin address for the code.

moveq   #12, d0         ; Load data register D0 with the immediate value 12.
                        ; This tests the MOVEQ instruction.

moveq   #7, d1          ; Load data register D1 with the immediate value 7.

add.l   d1, d0          ; Add the contents of D1 to D0.
                        ; D0 should now contain 19 (decimal).
                        ; This tests the ADD instruction between two registers.

movea.l #$2000, a0      ; Load address register A0 with the immediate address $2000.
                        ; This tests the MOVEA instruction.

move.w  d0, (a0)        ; Move the lower word of D0 to the memory location
                        ; pointed to by A0. After this, memory at $2000
                        ; should contain the value 19 ($0013).
                        ; This tests the MOVE instruction to memory.

rts						 ; The code runner considers RTS as the end of the program. Alternatively, use trap #15.

; --- End of Program ---
