; fibonacci.asm — Compute the first 16 Fibonacci numbers.
;
; Stores 16 long words starting at the "table" label.
; F(0) = 0, F(1) = 1, F(n) = F(n-1) + F(n-2).
;
; Run: run68 fibonacci.asm
; After execution, table contains:
;   0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233, 377, 610

    org     $1000

start:
    lea     table,a0            ; pointer to output table
    moveq   #0,d0               ; F(n-2) = 0
    moveq   #1,d1               ; F(n-1) = 1
    moveq   #15,d7            ; loop counter (dbra counts down to -1)

    move.l  d0,(a0)+            ; store F(0)
    move.l  d1,(a0)+            ; store F(1)
    subq    #2,d7               ; two values already stored

loop:
    move.l  d0,d2               ; d2 = F(n-2)
    add.l   d1,d2               ; d2 = F(n-2) + F(n-1) = F(n)
    move.l  d2,(a0)+            ; store F(n)
    move.l  d1,d0               ; shift: F(n-2) = old F(n-1)
    move.l  d2,d1               ; shift: F(n-1) = F(n)
    dbra    d7,loop             ; decrement and branch

    trap    #15                 ; halt

table:
    ds.l    16                  ; space for 16 long words
