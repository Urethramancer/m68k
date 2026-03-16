package main

import (
	"fmt"
	"os"
	"github.com/Urethramancer/m68k/assembler"
	"github.com/Urethramancer/m68k/disassembler"
)

func main(){
	src := ".org 0\n.dc.b \"This is a test string.\", 0xDE, 0xAD, 0xBE, 0xEF\n.even\n.dc.w 0x1234, 42\n"
	asm := assembler.New()
	orig, err := asm.Assemble(src, 0)
	if err!=nil{fmt.Fprintf(os.Stderr,"assemble failed: %v\n",err);os.Exit(1)}
	_ = os.WriteFile("/tmp/orig.bin", orig, 0644)
	out, err := disassembler.DisassembleAndDump(orig)
	if err!=nil{fmt.Fprintf(os.Stderr,"disasm failed: %v\n",err);os.Exit(1)}
	_ = os.WriteFile("/tmp/disasm.txt", []byte(out), 0644)
	// clean then reassemble
	cleaned := ""
	for _, line := range splitLines(out) {
		if len(line)==0{continue}
		if line[len(line)-1]==':'{cleaned+=line+"\n";continue}
		i:=0
		for i<len(line) && (line[i]==' '||line[i]=='\t'){i++}
		cleaned+=line[i:]+"\n"
	}
	re := assembler.New()
	b, err := re.Assemble(cleaned,0)
	if err!=nil{_=os.WriteFile("./reasm_error.txt", []byte(err.Error()),0644);fmt.Fprintf(os.Stderr,"reasm failed: %v\n",err);os.Exit(1)}
	_ = os.WriteFile("/tmp/reasm.bin", b,0644)
	fmt.Println("wrote artifacts")
}
