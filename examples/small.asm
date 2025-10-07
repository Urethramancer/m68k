; Small example program used during debugging of the assembler.
move #21,d0
link a5,#-100

loop:
bsr inc
bra loop

inc:
lsl #1,d0
rts

data:
.dc.b #10
.ds.b 10
