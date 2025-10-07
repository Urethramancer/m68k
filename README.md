# m68k Assembler and Disassembler

This repository contains a Motorola 68000 assembler (`asm68`) and disassembler (`dis68`), both written in Go.

## Overview

The assembler converts Motorola 68000 assembly source code into machine code, while the disassembler performs the reverse—translating binary machine code back into human-readable assembly.

Both tools are designed to be lightweight, portable, and easily extendable. They follow the Go convention of having separate command-line tools under the `cmd/` directory.

```
cmd/
├── asm68/   # Assembler CLI tool
└── dis68/   # Disassembler CLI tool
```

The assembler currently supports all standard addressing modes, most core instructions, and recognizes labels and `.org` directives.

### Note on `.org`

The `.org` directive is currently parsed for completeness and future expansion but does **not** affect output layout yet. There is not yet a binary or object format that uses absolute addressing, so `.org` has no effect other than reserving the address internally.

---

## Building

To build both tools:

```bash
go mod tidy
go build -o bin/asm68 ./cmd/asm68
go build -o bin/dis68 ./cmd/dis68
```

Or use the preconfigured VSCode build tasks (see `.vscode/tasks.json`).

### Build Output

The compiled tools will be located in the `bin/` directory.

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

Or, to print it to the terminal, omit the output:
```bash
dis68 input.bin
```

---

## Project Layout

```
.
├── assembler/       # Core assembler logic (mnemonic parsing, operand encoding)
├── cpu/             # CPU constants, opcodes, addressing modes, endianness helpers
├── disassembler/    # Disassembler logic (in progress)
├── cmd/
│   ├── asm68/       # Assembler CLI
│   └── dis68/       # Disassembler CLI
└── README.md
```

---

## Future Plans

- Implement `.org`-based binary layout control.
- Add full disassembler support for all instruction types.
- Introduce an object format for relocatable assembly.
- Add test suite covering all 14 addressing modes and instruction types.

---

## License

MIT License © 2025 Ronny Bangsund
