; hello.asm — Write "Hello, World!" to memory at $2000.
;
; Demonstrates: lea, moveq, byte-copy loop, branch, trap.
; Run: run68 hello.asm
; After execution, memory at $2000 contains the string.

    org     $1000

start:
    lea     message,a0          ; source pointer
    lea     $2000,a1            ; destination pointer

copy:
    move.b  (a0)+,d0            ; load next byte
    move.b  d0,(a1)+            ; store it
    bne.s   copy                ; loop until NUL terminator

    trap    #15                 ; halt

message:
    dc.b    'Hello, World!',0
    even
