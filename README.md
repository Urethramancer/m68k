# m68k

A Motorola 68000 assembler, disassembler and virtual machine, written in Go.

## Packages

| Package | Description |
|---------|-------------|
| `assembler` | Assembles MC68000 source into big-endian machine code |
| `disassembler` | Disassembles flat binaries back into readable assembly |
| `cpu` | Shared definitions: opcodes, addressing modes, status flags |
| `vm` | Virtual machine that loads and executes assembled programs |
| `system` | System call interface (stub, for future expansion) |

## Command-line tools

```
cmd/
├── asm68   Assembler
├── dis68   Disassembler
└── run68   Code runner / VM frontend
```

### asm68

Assembles one or more source files into a flat binary.

```sh
asm68 -o output.bin input.asm            # write binary to file
asm68 input.asm                          # assemble and hexdump to stdout
asm68 -o output.bin part1.asm part2.asm  # concatenate sources
```

### dis68

Disassembles a flat binary into MC68000 assembly.

```sh
dis68 input.bin                  # print disassembly to stdout
dis68 input.bin output.asm       # write disassembly to file
```

The disassembler performs control-flow analysis to separate code from data,
emits labels for branch targets and subroutine entry points, and resolves
PC-relative references to label names. Data regions are emitted as `dc.b`
hex byte directives.

### run68

Loads and executes a program in the virtual machine.

```sh
run68 program.asm                          # assemble and run
run68 -load 0x1000 program.bin             # load binary at address
run68 -d0 FF -a0 2000 program.asm          # set initial registers (hex)
run68 -cycles 5000 program.asm             # limit execution
```

The VM provides 16 MB of RAM. Execution halts on `rts` or `trap #15`.

## Assembler features

- Full MC68000 instruction set including all 12 standard addressing modes
- Labels, forward references and automatic PC-relative encoding
- `SP` alias for `A7` in all contexts (operands, register lists)
- Directives: `.org`, `dc.b`, `dc.w`, `dc.l`, `ds.b`, `ds.w`, `ds.l`, `even`
- Comment syntax: `;` and `*`
- Multi-pass sizing with label stabilisation

## Disassembler features

- Linear sweep followed by control-flow analysis
- Dead code recovery for instructions after unconditional branches
- Automatic labelling: `sub_XXXX` for subroutine targets, `loc_XXXX` for branches
- PC-relative operands resolved to bare label references
- Data regions emitted as even-aligned `dc.b` lines for round-trip fidelity
- Big-endian input (native M68k byte order)

## Go API

Both packages expose a minimal public interface:

```go
// Assembler
code, err := assembler.Assemble(source, baseAddress)
code, err := assembler.AssembleFile("input.asm", 0)

// Disassembler
text, err := disassembler.Disassemble(code)
text, err := disassembler.DisassembleFile("input.bin")

// Single-instruction decoding
mnemonic, operands, extBytes := disassembler.Decode(opcode, pc, extensionBytes)
```

See [docs/architecture.md](docs/architecture.md) for internals.

## Building

```sh
go build ./cmd/asm68
go build ./cmd/dis68
go build ./cmd/run68
```

Or build all at once:

```sh
go build ./...
```

## Testing

```sh
go test ./...
```

## Examples

The `examples/` directory contains demonstration programs:

| File | Description |
|------|-------------|
| `hello.asm` | Writes "Hello, World!" to memory using a byte-copy loop |
| `fibonacci.asm` | Computes Fibonacci numbers and stores them in a table |
| `bubblesort.asm` | Sorts an array of words using bubble sort |
| `main.asm` | Simple addition — runs in the VM (`run68`) |
| `full.asm` | Catalogue of all addressing modes and instruction groups |

The VM currently implements a subset of the MC68000 instruction set (see
[docs/architecture.md](docs/architecture.md)). Most examples are designed to
demonstrate the assembler and disassembler. Use `main.asm` for a working VM
demo.

Assemble and disassemble an example:

```sh
asm68 -o /tmp/hello.bin examples/hello.asm
dis68 /tmp/hello.bin
```

Run the VM-compatible example:

```sh
run68 examples/main.asm
```

## Project layout

```
.
├── assembler/       Assembler package
├── cpu/             CPU definitions (opcodes, modes, flags)
├── disassembler/    Disassembler package
├── vm/              Virtual machine
├── system/          System call stubs
├── cmd/
│   ├── asm68/       Assembler CLI
│   ├── dis68/       Disassembler CLI
│   └── run68/       VM runner CLI
├── tests/           Integration tests
├── examples/        Example programs
└── docs/            Technical documentation
```

## Licence

MIT — see [LICENSE](LICENSE).
