package capturelib

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Urethramancer/m68k/assembler"
	"github.com/Urethramancer/m68k/disassembler"
)

// childMain runs the short-lived child work: disassemble and write stage1 artifacts.
func childMain(bytes []byte) {
	// synchronous in-library disassembly dump to survive early kills
	out, err := disassembler.DisassembleAndDump(bytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "child disassemble failed: %v\n", err)
		os.Exit(2)
	}
	// write heartbeat to indicate stage1 completed
	_ = os.WriteFile("/tmp/capture_stage1_done", []byte("ok"), 0644)

	// also write the textual disassembly file (redundant)
	if err := os.WriteFile("/tmp/capture_disasm.txt", []byte(out), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "child write disasm failed: %v\n", err)
		os.Exit(2)
	}
	// child exits quickly after producing stage1 artifacts
	fmt.Printf("child: wrote /tmp/capture_disasm.txt and /tmp/capture_stage1_done\n")
	os.Exit(0)
}

// Run performs the full capture round-trip: assemble, disassemble (in a child process), and reassemble.
func Run() {
	// assemble the original blob
	src := ".org 0\n.dc.b \"This is a test string.\", 0xDE, 0xAD, 0xBE, 0xEF\n.even\n.dc.w 0x1234, 42\n"
	asm := assembler.New()
	bytes, err := asm.Assemble(src, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "assemble failed: %v\n", err)
		os.Exit(1)
	}
	// write original bytes immediately
	if err := os.WriteFile("/tmp/capture_orig.bin", bytes, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write orig failed: %v\n", err)
		os.Exit(1)
	}

	// If invoked as the child, run short-lived disassembly and exit.
	if os.Getenv("CAPTURE_CHILD") == "1" {
		childMain(bytes)
	}

	// spawn a short-lived child process to perform the disassembly work
	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(), "CAPTURE_CHILD=1")
	// inherit stdio so we can see child prints in logs
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start child failed: %v\n", err)
		// fall back to in-process disassembly
		out, err := disassembler.DisassembleAndDump(bytes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "disassemble failed: %v\n", err)
			os.Exit(1)
		}
		_ = os.WriteFile("/tmp/capture_disasm.txt", []byte(out), 0644)
		_ = os.WriteFile("/tmp/capture_stage1_done", []byte("ok"), 0644)
	} else {
		// wait for child with a modest timeout
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case err := <-done:
			if err != nil {
				fmt.Fprintf(os.Stderr, "child exited with error: %v\n", err)
			}
		case <-func() <-chan time.Time {
			// configurable timeout via CAPTURE_CHILD_TIMEOUT (seconds); default 10s
			sec := 10
			if v := os.Getenv("CAPTURE_CHILD_TIMEOUT"); v != "" {
				if t, err := strconv.Atoi(v); err == nil && t > 0 {
					sec = t
				}
			}
			return time.After(time.Duration(sec) * time.Second)
		}():
			// timeout: try to kill the child process to avoid zombies
			_ = cmd.Process.Kill()
			fmt.Fprintf(os.Stderr, "child timed out and was killed\n")
		}
	}

	// At this point, hope that /tmp/capture_disasm.txt and /tmp/capture_stage1_done exist.
	// If not present, fall back to in-process disassembly (last resort).
	var out string
	if _, err := os.Stat("/tmp/capture_stage1_done"); err == nil {
		b, _ := os.ReadFile("/tmp/capture_disasm.txt")
		out = string(b)
		fmt.Printf("parent: observed stage1 artifacts\n")
	} else {
		fmt.Printf("parent: stage1 artifacts missing, running in-process disassembly\n")
		o, err := disassembler.DisassembleAndDump(bytes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "disassemble failed: %v\n", err)
			os.Exit(1)
		}
		out = o
		_ = os.WriteFile("/tmp/capture_disasm.txt", []byte(out), 0644)
		_ = os.WriteFile("/tmp/capture_stage1_done", []byte("ok"), 0644)
	}

	fmt.Printf("wrote /tmp/capture_disasm_inside.txt (via DisassembleAndDump), /tmp/capture_disasm.txt, and /tmp/capture_orig.bin\n")

	// give the system a short grace window for the stage1 files to be observed
	if _, err := os.Stat("/tmp/capture_stage1_done"); err == nil {
		time.Sleep(2 * time.Second)
	}

	// proceed to reassembly (may be killed later)
	cleaned := ""
	for _, line := range splitLines(out) {
		if len(line) == 0 {
			continue
		}
		if line[len(line)-1] == ':' {
			cleaned += line + "\n"
			continue
		}
		i := 0
		for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
			i++
		}
		cleaned += line[i:] + "\n"
	}

	reAsm := assembler.New()
	bytes2, err := reAsm.Assemble(cleaned, 0)
	if err != nil {
		// write assembler error and the cleaned disassembly for debugging
		_ = os.WriteFile("/tmp/capture_reasm_error.txt", []byte(fmt.Sprintf("err: %v\n\ncleaned:\n%s\n", err, cleaned)), 0644)
		os.Exit(1)
	}
	if err := os.WriteFile("/tmp/capture_reasm.bin", bytes2, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write reasm failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wrote /tmp/capture_orig.bin (%d bytes), /tmp/capture_disasm.txt, and /tmp/capture_reasm.bin (%d bytes)\n", len(bytes), len(bytes2))
}

// RunAtomic is a simplified version of Run that focuses on the atomic write of the disassembly file.
func RunAtomic() {
	src := ".org 0\n.dc.b \"This is a test string.\", 0xDE, 0xAD, 0xBE, 0xEF\n.even\n.dc.w 0x1234, 42\n"
	asm := assembler.New()
	bytes, err := asm.Assemble(src, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "assemble failed: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile("/tmp/capture_orig.bin", bytes, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write orig failed: %v\n", err)
		os.Exit(1)
	}
	out, err := disassembler.DisassembleAndDump(bytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "disassemble failed: %v\n", err)
		os.Exit(1)
	}
	// atomic write: write to temp file then rename
	tmp := filepath.Join(os.TempDir(), "capture_disasm.tmp")
	if err := os.WriteFile(tmp, []byte(out), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write tmp disasm failed: %v\n", err)
		os.Exit(1)
	}
	if err := os.Rename(tmp, "/tmp/capture_disasm.txt"); err != nil {
		fmt.Fprintf(os.Stderr, "rename disasm failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wrote /tmp/capture_disasm.txt and /tmp/capture_orig.bin\n")
}

func splitLines(s string) []string {
	var lines []string
	cur := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, cur)
			cur = ""
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}
