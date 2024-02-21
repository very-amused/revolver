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
	flag.BoolVar(&listMethods, "l", false, "List defined methods for opening the files (format i:label:flags:cmd)")
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

	// List all matching methods if -l was passed
	if listMethods {
		// Verify the matches can open all files specified
		matches := config.AllMatches(files[0])
		for _, file := range files[1:] {
			for _, method := range matches {
				if !method.Match(file) {
					fmt.Fprintln(os.Stderr, "files cannot have different types")
					os.Exit(1)
				}
			}
		}

		// Print the matches
		for i, method := range matches {
			fmt.Printf("%d:%s\n", i, method)
		}
		return
	}

	// Run the command, forking if requested
	if method := config.Match(files[0], methodIndex); method != nil {
		// Verify the match can open all files specified
		for _, file := range files[1:] {
			if !method.Match(file) {
				fmt.Fprintln(os.Stderr, "files cannot have different types")
				os.Exit(1)
			}
		}

		var cmd *exec.Cmd
		if method.Term {
			// Read $TERMCMD
			termCmd := os.Getenv("TERMCMD")
			if len(termCmd) == 0 {
				fmt.Fprintln(os.Stderr, "TERMCMD must be set to run commands in new terminals")
				os.Exit(1)
			}

			// Use $TERMCMD -e to run in a new term
			cmd = exec.Command(termCmd, "-e", shell, "-c", method.buildCommand(files))
		} else {
			cmd = exec.Command(shell, "-c", method.buildCommand(files))
		}
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

		// Run the commmand
		cmd.Start()

		// If forking, don't wait for an exit
		if method.Fork {
			return
		}
		if err := cmd.Wait(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to run command: %v\n", err)
			os.Exit(cmd.ProcessState.ExitCode())
		}
	}
}
