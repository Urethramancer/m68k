# m68k Assembler and Disassembler

This repository contains a Motorola 68000 assembler (asm68) and disassembler (dis68), both written in Go. They're implemented as the packages **assembler** and **disassembler**, with a common **cpu** package defining instruction opcodes, modes and other bits.

## Overview

The assembler converts Motorola 68000 assembly source code into machine code, while the disassembler performs the reverse — translating binary machine code back into human-readable assembly.

Both tools are designed to be lightweight, portable, and easily extendable. They follow Go conventions, with separate command-line tools under the cmd/ directory.

```
cmd/
├── asm68/	# Assembler CLI tool
└── dis68/	# Disassembler CLI tool
└── run68/	# Code runner CLI tool
```

## Assembler (asm68)

The assembler provides comprehensive support for the standard MC68000 instruction set and addressing modes. It can assemble straight binary code and accepts standard directives.

### Features

* Full support for all **12 standard MC68000 addressing modes**.
* Encodes **branches**, **jumps**, **bit manipulation**, **logical**, **arithmetic**, and **shift/rotate** operations.
* Handles **labels** and basic **directives**, including ORG for setting the internal program counter during assembly.
* Supports **comment syntax** (; and \#) consistent with standard Motorola assemblers.

## Disassembler (dis68)

The disassembler takes raw Motorola 68000 machine code (big-endian) and reconstructs readable assembly listings. It is designed for static analysis of flat binary data (e.g., ROM images, raw assembled binaries).

### Key Features

* **Instruction decoding:** Handles most core opcodes, including arithmetic, logic, move, branch, system, and bit manipulation instructions.
* **Automatic data detection:** Uses multiple heuristics to distinguish between code and data:
  * Recognises **ASCII strings** and **NULL-terminated text**.
  * Detects **aligned 4-byte tags** (common in version or identifier fields).
  * Identifies long runs of unreferenced words or bytes as probable **data blocks**, formatted as dc.b lines.
* **Flow-aware decoding:** Tracks **branch and subroutine targets** (bra, bsr, jmp, jsr) to identify reachable code. All other regions are treated as data.
* **Readable output:** Uses contextual labels such as sub\_XXXX: for subroutines and loc\_XXXX: for local branch targets for reference clarity. Outputs both string literals and raw data bytes using standard Motorola syntax.
* **Consistent endianness:** All decoding assumes **big-endian input** (the native M68k byte order), regardless of host platform.

### **Example**

Given a binary file displayed with hexdump \-C:

$ hexdump -C input.bin

```
00000000  41 fa 00 10 70 0d 4e b9  00 00 00 0e 4e 75 4e 71  |A...p.N.....NuNq|
00000010  4e 75 54 68 69 73 20 69  73 20 61 20 74 65 73 74  |NuThis is a test|
00000020  20 73 74 72 69 6e 67 2e  00 00 de ad be ef 56 45  | string.......VE|
00000030  52 31 00 00 41 42 43 00  43 6f 70 79 72 69 67 68  |R1..ABC.Copyrigh|
00000040  74 20 28 43 29 20 32 30  32 35 00 00              |t (C) 2025..|
0000004c
```

...

The disassembler outputs:

```
    lea      ($10,pc),a0
    moveq    #13,d0
    jsr      sub_000E
    rts
sub_000E:
    nop
    rts
string1: dc.b    'This is a test string.',$00
    dc.b    $00,$de,$ad,$be,$ef
string2: dc.b    'VER1',$00
    dc.b    $00
    dc.b    $41,$42,$43
    dc.b    $00
string3: dc.b    'Copyright (C) 2025',$00
    dc.b    $00
```

### **Heuristics Summary**

| Type | Detection Rule | Output Example |
| :---- | :---- | :---- |
| **Code** | Follows branch and subroutine flow | bsr, bra, jsr, etc. |
| **String** | ≥4 printable bytes ending in 0x00 | dc.b 'Text here',$00 |
| **Tag** | 4 aligned printable bytes, not null-terminated | dc.b 'VER1' |
| **Data** | Non-printable or invalid opcode runs | dc.b $00,$de,$ad,$be,$ef |

## **Building**

To build both tools:

go mod tidy
go build \-o bin/asm68 ./cmd/asm68
go build \-o bin/dis68 ./cmd/dis68

The compiled tools are placed in the bin/ directory.

## **Usage**

### **Assembler**

./bin/asm68 input.asm output.bin

### **Disassembler**

./bin/dis68 input.bin output.asm

To print disassembly to the terminal:

./bin/dis68 input.bin

## **Project Layout**

```
.
├── assembler/       \# Core assembler logic (mnemonic parsing, operand encoding)
├── cpu/             \# CPU constants, opcodes, addressing modes, endianness helpers
├── disassembler/    \# Disassembler logic (decoding, EA resolution, data heuristics)
├── cmd/
│   ├── asm68/       \# Assembler CLI
│   └── dis68/       \# Disassembler CLI
└── README.md
````

## Possible improvements

* Introduce a relocatable object format.
* Implement .org-based binary layout control (needs the above format).
* A virtual machine package that runs m68 code and shows register status.

## **Licence**

MIT License © 2025 Ronny Bangsund