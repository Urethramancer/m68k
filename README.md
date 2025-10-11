# m68k Assembler and Disassembler

This repository contains a Motorola 68000 assembler (`asm68`) and disassembler (`dis68`), both written in Go.

## Overview

The assembler converts Motorola 68000 assembly source code into machine code, while the disassembler performs the reverse — translating binary machine code back into human-readable assembly.

Both tools are designed to be lightweight, portable, and easily extendable.
They follow Go conventions, with separate command-line tools under the `cmd/` directory.

```
cmd/
├── asm68/   # Assembler CLI tool
└── dis68/   # Disassembler CLI tool
```

---

## Assembler (`asm68`)

The assembler supports most of the core Motorola 68000 instruction set and addressing modes.
It can assemble straight binary code and accepts standard directives such as `.dc`, `.ds`, `.even`, and `.org`.

### Features

- Full support for **data and address registers**, **immediate values**, and **effective addressing modes**.
- Encodes **branches**, **jumps**, **bit manipulation**, **logical**, **arithmetic**, and **shift/rotate** operations.
- Handles **labels** and basic **directives** (`.org`, `.dc.b`, `.dc.w`, `.dc.l`, `.ds.b`, `.even`).
- Supports **comment syntax** (`;` and `#`) consistent with standard Motorola assemblers.

### Note on `.org`

The `.org` directive is parsed for completeness but does **not** currently affect output layout.
There is not yet a binary or object format using absolute addressing, so `.org` has no runtime effect beyond internal tracking.

---

## Disassembler (`dis68`)

The disassembler takes raw Motorola 68000 machine code (big-endian) and reconstructs readable assembly listings.
It is designed for static analysis of flat binary data (e.g., ROM images, raw assembled binaries).

### Key Features

- **Instruction decoding:**
  Handles most core opcodes, including arithmetic, logic, move, branch, system, and bit manipulation instructions.

- **Automatic data detection:**
  Uses multiple heuristics to distinguish between code and data:
  - Recognizes **ASCII strings** and **NULL-terminated text**.
  - Detects **aligned 4-byte tags** (common in version or identifier fields).
  - Identifies long runs of unreferenced words or bytes as probable **data blocks**, formatted as `dc.b` lines.

- **Flow-aware decoding:**
  Tracks **branch and subroutine targets** (`bra`, `bsr`, `jmp`, `jsr`) to identify reachable code.
  All other regions are treated as data.

- **Readable output:**
  Uses labels such as `string1:`, `data_XXXX:`, or `loc_XXXX:` for reference clarity.
  Outputs both string literals and raw data bytes using standard Motorola syntax.

- **Consistent endianness:**
  All decoding assumes **big-endian input** (the native M68k byte order), regardless of host platform.
  The disassembler uses the shared `cpu/endian.go` utilities to ensure consistent interpretation on both little- and big-endian hosts.

### Example

Given a binary file:

```
00000000: 41fa0010 700d 4eb9 0000000e 4e75 4e71 4e75 ...
```

The disassembler outputs:

```asm
    lea      ($10,pc),a0
    moveq    #13,d0
    jsr      $e.l
    rts
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

### Heuristics Summary

| Type | Detection Rule | Output Example |
|------|----------------|----------------|
| **Code** | Follows branch and subroutine flow | `bsr`, `bra`, `jsr`, etc. |
| **String** | ≥4 printable bytes ending in `0x00` | `dc.b 'Text here',$00` |
| **Tag** | 4 aligned printable bytes, not null-terminated | `dc.b 'VER1'` |
| **Data** | Non-printable or invalid opcode runs | `dc.b $00,$de,$ad,$be,$ef` |

---

## Building

To build both tools:

```bash
go mod tidy
go build -o bin/asm68 ./cmd/asm68
go build -o bin/dis68 ./cmd/dis68
```

Or use the preconfigured VSCode build tasks (`.vscode/tasks.json`).

### Build Output

The compiled tools are placed in the `bin/` directory.

---

## Usage

### Assembler

```bash
asm68 input.asm output.bin
```

### Disassembler

```bash
dis68 input.bin output.asm
```

To print disassembly to the terminal:

```bash
dis68 input.bin
```

---

## Project Layout

```
.
├── assembler/       # Core assembler logic (mnemonic parsing, operand encoding)
├── cpu/             # CPU constants, opcodes, addressing modes, endianness helpers
├── disassembler/    # Disassembler logic (decoding, EA resolution, data heuristics)
├── cmd/
│   ├── asm68/       # Assembler CLI
│   └── dis68/       # Disassembler CLI
└── README.md
```

---

## Future Plans

- Implement `.org`-based binary layout control.
- Expand disassembler to cover all M68000/010/020 instruction sets.
- Improve data/code flow tracking for interleaved data tables.
- Introduce a relocatable object format and linker.
- Add a comprehensive test suite for instruction encoding and decoding.

---

## License

MIT License © 2025 Ronny Bangsund
