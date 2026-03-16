# Architecture

This document describes the internal design of the assembler, disassembler
and virtual machine.

## Assembler

### Pipeline

1. **Parsing** — Source lines are split into `Node` values, each holding a
   label, mnemonic, size suffix and parsed operands. Comment and blank lines
   are discarded. Directives (`.org`, `dc.*`, `ds.*`, `even`) are recorded as
   directive nodes.

2. **Sizing passes** — The assembler iterates over all nodes, assigning a
   program counter to each. Label addresses are recorded. Because a forward-
   referenced label might change the size of a PC-relative operand (short vs
   long), the pass repeats until all label addresses stabilise (up to 10
   iterations).

3. **Code generation** — A final pass walks the nodes again, emitting binary
   words for each instruction and raw bytes for directives. Labels are fully
   resolved; any remaining undefined label is an error.

### Operand encoding

`parseOperand` is the dispatcher. It normalises the `SP` alias to `A7`, then
tries a sequence of pattern groups in order of specificity:

- Status register names (`sr`, `ccr`, `usp`)
- Indexed modes: `d8(An,Xn.s)` and `d8(PC,Xn.s)`
- PC-relative: `label(pc)` and `(disp,pc)`
- Register modes: `Dn`, `An`, `(An)`, `(An)+`, `-(An)`, `d16(An)`
- Absolute: `$addr.w`, `$addr.l`, `($addr).w`, `($addr).l`, `$addr`
- Immediate: `#value`
- MOVEM register lists: `d0-d3/a0-a2`
- Bare labels

Bare labels are resolved during code generation. If the instruction supports
PC-relative addressing and the offset fits in 16 bits, PC-relative encoding
is chosen automatically. Otherwise absolute long addressing is used.

### Directives

| Directive | Effect |
|-----------|--------|
| `.org` / `org` | Sets the program counter |
| `dc.b` | Emits bytes (hex, decimal, character literals) |
| `dc.w` | Emits 16-bit words |
| `dc.l` | Emits 32-bit long words |
| `ds.b` / `ds.w` / `ds.l` | Reserves zero-filled space |
| `even` | Pads to the next even address |

## Disassembler

### Pipeline

1. **Linear sweep** — Every two-byte-aligned offset is decoded as a potential
   instruction. Results are stored in a map keyed by address. This is
   optimistic: data bytes may decode as nonsense instructions.

2. **Control-flow analysis** — Starting from address 0, a work queue follows
   branch and subroutine targets, marking reachable instructions as code
   (`IsCode = true`). Unreachable instructions remain unmarked.

3. **Dead code recovery** — A second pass finds decoded instructions that
   immediately follow a code-marked instruction but were never reached by the
   work queue (e.g. dead code after an unconditional branch). These are
   marked as code and their branch targets are added to the label map.

4. **Rendering** — The instruction map is walked in address order. Code
   instructions are printed with mnemonic and operands. Gaps between code
   regions are emitted as `dc.b` hex byte directives. Labels are inserted
   at their target addresses. PC-relative operands are resolved to label
   names.

### Label types

| Prefix | Meaning |
|--------|---------|
| `sub_XXXX` | Target of `jsr` or `bsr` (subroutine entry) |
| `loc_XXXX` | Target of a conditional or unconditional branch |

### Data emission

All data regions are emitted as `dc.b` lines with explicit hex values
(`$NN`). Each line contains an even number of bytes to prevent the
assembler's automatic padding from inserting unwanted zeros during
round-trip reassembly.

## Virtual machine

The VM wraps a `cpu.CPU` instance with 16 MB of flat memory. Code is loaded
at a specified address and the program counter is set accordingly.

### Execution model

The CPU fetches, decodes and executes one instruction per call to
`cpu.Execute()`. The `vm.Start()` method runs this in a goroutine;
`vm.Stop()` halts it. The `run68` CLI drives execution synchronously with
a configurable cycle limit.

Execution stops when:
- `rts` is executed and the stack is at its initial position
- `trap #15` is executed (conventional program exit)
- The cycle limit is reached
- An unimplemented or illegal opcode is encountered

### CPU state

```
D0-D7       Data registers (32-bit)
A0-A6       Address registers (32-bit)
A7 / SP     Active stack pointer (USP in user mode, SSP in supervisor mode)
PC          Program counter
SR          Status register: T.S..III..XNZVC
```

The `DumpRegisters()` method prints all registers with the status register
decoded into flag mnemonics.

## Round-trip fidelity

The assembler and disassembler are designed to round-trip: assembling a
source file, disassembling the binary, and reassembling should produce
identical bytes. This is verified by the `TestRoundTrip*` tests.

Key design choices that support this:
- Data is always emitted as explicit hex bytes, never quoted strings
- `dc.b` lines always have an even byte count (no assembler padding)
- PC-relative operands are emitted as bare labels (the assembler re-selects
  PC-relative encoding automatically)
- The dead code recovery pass prevents valid instructions from being emitted
  as data bytes
