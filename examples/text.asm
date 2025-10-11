; A simple routine to demonstrate that the disassembler
; correctly identifies executable code.
start:
	lea     message,a0  ; Load the address of our string into A0
	moveq   #13,d0        ; A dummy instruction
	jsr     somewhere   ; A fake subroutine call
	rts                 ; Return from subroutine

somewhere:
	nop
	rts

; This section is deliberately placed after the code to create a "gap"
; for the data analyzer to process. It contains a mix of strings and
; non-string bytes.
	even                ; Ensure data starts on an even address

message:
	dc.b    'This is a test string.', 0 ; A standard null-terminated string
	dc.b    $DE,$AD,$BE,$EF             ; Some non-string hex data

tag:
	dc.b    'VER1'                      ; A 4-byte tag, not null-terminated
	dc.b	0
	dc.b    $41,$42,$43                 ; "ABC" - too short to be a string

footer:
	dc.b    'Copyright (C) 2025'        ; Another valid string
	dc.b	0
