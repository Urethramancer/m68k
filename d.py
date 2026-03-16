#!/usr/bin/env python3
import re, sys, os

disasm_path = "/tmp/disasm.txt"
orig_path = "/tmp/orig.bin"
reasm_path = "/tmp/reasm.bin"

if not os.path.isfile(disasm_path):
    print("ERROR: missing", disasm_path, file=sys.stderr); sys.exit(2)
if not os.path.isfile(orig_path):
    print("ERROR: missing", orig_path, file=sys.stderr); sys.exit(2)
if not os.path.isfile(reasm_path):
    print("ERROR: missing", reasm_path, file=sys.stderr); sys.exit(2)

with open(disasm_path, "r", encoding="utf-8", errors="ignore") as f:
    lines = [l.rstrip("\n") for l in f]

hex_re = re.compile(r'$([0-9a-fA-F]{2})')
byte_line_map = []  # list of (byte_value:int, line_index:int)
for i, l in enumerate(lines):
    if "dc.b" in l:
        for m in hex_re.finditer(l):
            byte_line_map.append((int(m.group(1), 16), i, l))

if not byte_line_map:
    hex2_re = re.compile(r'0x([0-9a-fA-F]{2})')
for i, l in enumerate(lines):
        if "dc.b" in l or "dc.w" in l or "dc.l" in l:
            for m in hex2_re.finditer(l):
                byte_line_map.append((int(m.group(1),16), i, l))

orig = open(orig_path, "rb").read()
reasm = open(reasm_path, "rb").read()

mapping = [-1] * len(orig)
idx = 0
for bval, lineidx, l in byte_line_map:
    if idx >= len(orig):
        break
    mapping[idx] = lineidx
    idx += 1

last_line = byte_line_map[-1][1] if byte_line_map else len(lines)-1
while idx < len(orig):
    mapping[idx] = last_line
    idx += 1

print("Offset  Orig  Reasm  Line#  Disasm-line")
for i in range( min(len(orig), len(reasm)) ):
    o = orig[i]
    r = reasm[i]
    if o != r:
        ln = mapping[i]
        lstr = lines[ln] if 0 <= ln < len(lines) else "(no disasm line)"
        print(f"{i:04x}   0x{o:02x}  0x{r:02x}   {ln:5d}   {lstr}")
