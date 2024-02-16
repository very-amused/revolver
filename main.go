package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

var shell = "/bin/sh"

// CLI args
var (
	listMethods bool
	methodIndex uint
)

func main() {
	// Process flags
	flag.BoolVar(&listMethods, "l", false, "List defined methods for opening the files with format i:flags:cmd")
	flag.UintVar(&methodIndex, "p", 0, "Which method to use for opening the file")
	flag.Parse()

	// Run commands with $SHELL if set
	if userShell := os.Getenv("SHELL"); len(userShell) > 0 {
		shell = userShell
	}

	files := flag.Args()
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "no files specified\n")
		os.Exit(1)
	}

	// Run the command, forking if requested
	if method := config.Match(files[0], methodIndex); method != nil {
		// Verify the match can open all files specified
		for _, file := range files[1:] {
			if !method.Match(file) {
				fmt.Fprintf(os.Stderr, "files cannot have different types")
				os.Exit(1)
			}
		}

		cmd := exec.Command(shell, "-c", method.Command)
		if method.Fork {
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Setpgid: true,
				Pgid:    0}
		}
		if !method.Fork || method.Term {
			// Connect stdio if not forking or running in a new term
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to run command: %v", err)
		}
	}
}
