; --- move from sr (supervisor only)
move_from_sr:
        move    sr,d0          ; expect: 40c0
        move    sr,(a0)        ; expect: 40d0

; --- move to sr (supervisor only)
move_to_sr:
        move    d0,sr          ; expect: 46c0
        move    (a0),sr        ; expect: 46d0

; --- move from ccr (user allowed, mc68010+)
move_from_ccr:
        move    ccr,d1         ; expect: 42c1
        move    ccr,(a1)       ; expect: 42d1

; --- move to ccr (user allowed, mc68010+)
move_to_ccr:
        move    d1,ccr         ; expect: 44c1
        move    (a1),ccr       ; expect: 44d1
