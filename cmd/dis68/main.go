package main

import (
	"fmt"
	"os"

	"github.com/Urethramancer/m68k/disassembler"
	"github.com/grimdork/climate/arg"
)

func main() {
	opt := arg.New("dis68")
	opt.SetDefaultHelp(true)
	err := opt.SetPositional("INPUT", "Binary file to disassemble.", "", true, arg.VarString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting positional argument: %v\n", err)
		os.Exit(1)
	}

	err = opt.SetOption(arg.GroupDefault, "o", "out", "Write disassembly to file (default: stdout).", "", false, arg.VarString, nil)
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

	input := opt.GetPosString("INPUT")
	code, err := os.ReadFile(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	text, err := disassembler.Disassemble(code)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Disassembly error: %v\n", err)
		os.Exit(1)
	}

	fn := opt.GetString("out")
	if fn != "" {
		if err := os.WriteFile(fn, []byte(text), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Disassembly written to %s\n", fn)
		return
	}

	fmt.Print(text)
}
