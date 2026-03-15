TODO: Disassembler

High-priority
- [ ] Ensure linear-sweep advances PC by full instruction length (2 + used). (DONE)
- [ ] Ensure control-flow analysis uses Instruction.Size for advancing and label target computation. (DONE)
- [ ] Guarantee branch targets discovered during control-flow analysis are emitted as label declarations in final disassembly so re-assembly succeeds for forward/backward refs.
- [ ] Fix remaining round-trip failures (notably TestRoundTripMovemMovepDirectives): trace label emission and assembly symbol table to find why some labels (e.g., sub_0020) are undefined.
- [ ] Add targeted unit tests for branch encodings: short/word/long, bsr/jsr variants, DBcc edge cases.

Medium-priority
- [ ] Canonicalize EA formatting across disassembler and assembler: movep, movem, PC-relative, indexed modes, absolute .w/.l forms.
- [ ] Add tests for all EA permutations and for truncated extension handling.
- [ ] Ensure decodeJmpJsr and other decoders correctly return 'used' extension lengths and that Instruction.Size is set accordingly.

Nice-to-have / Future (pin for later)
- [ ] Add a "roundtrip" output mode flag to the Disassemble() API or CLI that makes the disassembler emit assembler-friendly text (labels, hex constants) optimized for re-assembly. Default mode should preserve canonical test-friendly formatting.
- [ ] Improve logging/tracing helpers to dump labelTargets/printedLabels when round-trip assembly fails for easier debugging.
- [ ] Create a comprehensive testdata file (~100-200 lines) exercising mixed code/data, odd alignments, and every addressing-mode edge case: disassembler/testdata/edgecases_100_lines.s
- [ ] Consider adding a separate renderer that takes decoded instructions + labelTargets and produces multiple output styles (canonical, roundtrip, verbose)

Developer notes
- When modifying formatting, always run 'go test ./...' and check both disassembler and assembler packages for round-trip regressions.
- Keep Instruction.Size == 2 + used; many downstream analyses rely on it.

Reference
- Files of interest:
  - disassemble.go
  - branch.go
  - utility.go
  - move.go
  - tests/*

If you address a TODO, remove the line or mark it DONE and include a test demonstrating the fix.
