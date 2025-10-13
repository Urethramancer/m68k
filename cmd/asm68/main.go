package main

import (
	"fmt"
	"os"

	"github.com/Urethramancer/m68k/assembler"
	"github.com/Urethramancer/m68k/disassembler"
	"github.com/grimdork/climate/arg"
	"github.com/grimdork/climate/str"
)

func main() {
	opt := arg.New("asm68")
	opt.SetDefaultHelp(true)
	err := opt.SetPositional("SOURCE", "Source file(s) to assemble. They will be assembled in the order listed.", "", true, arg.VarStringSlice)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting positional argument: %v\n", err)
		os.Exit(1)
	}

	err = opt.SetOption(arg.GroupDefault, "o", "out", "Binary output file (default: stdout)", "", false, arg.VarString, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting option: %v\n", err)
		os.Exit(1)
	}

	err = opt.Parse(os.Args[1:])
	if err != nil {
		if err == arg.ErrNoArgs {
			opt.PrintHelp()
			return

		}

		fmt.Fprintf(os.Stderr, "Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	files := opt.GetPosStringSlice("SOURCE")
	// List the files to be assembled.
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "No source files specified.\n")
		os.Exit(1)
	}

	src := str.NewStringer()
	var count int
	for _, fn := range files {
		data, err := os.ReadFile(fn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading source file: %v\n", err)
			os.Exit(1)
		}

		n, err := src.Write(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing source file: %v\n", err)
			os.Exit(1)
		}

		count += n
		// Add a newline between files to avoid accidental token merging.
		_, err = src.WriteString("\n")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing source file: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Read %d bytes of source code.\n", count)
	asm := assembler.New()
	code, err := asm.Assemble(string(src.String()), 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Assembly error: %v\n", err)
		os.Exit(1)
	}

	fn := opt.GetString("out")
	if fn != "" {
		if err := os.WriteFile(fn, code, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Assembled binary written in M68K big-endian format to %s\n", fn)
		return
	}

	disassembler.Hexdump(code)
}
