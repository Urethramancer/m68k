; bubblesort.asm — Sort an array of 16-bit words using bubble sort.
;
; The array "data" contains 10 unsorted words. After execution,
; they are sorted in ascending order in place.
;
; Algorithm:
;   repeat
;     swapped = false
;     for i = 0 to n-2
;       if array[i] > array[i+1]
;         swap them
;         swapped = true
;     until not swapped
;
; Run: run68 bubblesort.asm

    org     $1000

start:
    lea     data,a0             ; base address of the array
    moveq   #9,d6             ; number of elements minus 1

outer:
    moveq   #0,d5               ; swapped flag (0 = no swaps)
    move.l  a0,a1               ; reset scan pointer to array start
    move.w  d6,d7               ; inner loop counter = n-1
    subq.w  #1,d7               ; compare n-1 pairs

inner:
    move.w  (a1),d0             ; d0 = array[i]
    move.w  2(a1),d1            ; d1 = array[i+1]
    cmp.w   d1,d0               ; compare array[i] with array[i+1]
    ble.s   noswap              ; if array[i] <= array[i+1], skip

    ; Swap the two elements.
    move.w  d1,(a1)             ; array[i] = smaller value
    move.w  d0,2(a1)            ; array[i+1] = larger value
    moveq   #1,d5               ; set swapped flag

noswap:
    addq.l  #2,a1               ; advance to next element
    dbra    d7,inner            ; next pair

    tst.w   d5                  ; any swaps this pass?
    bne.s   outer               ; if yes, repeat

    trap    #15                 ; halt

data:
    dc.w    42, 17, 93, 5, 88
    dc.w    31, 72, 11, 56, 3
