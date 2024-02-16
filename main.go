package main

import (
	"os"
	"os/exec"
	"syscall"
)

var shell = "/bin/sh"

func main() {
	// Run commands with $SHELL if set
	if userShell := os.Getenv("SHELL"); len(userShell) > 0 {
		shell = userShell
	}

	// Run the command, forking if requested
	if command := config.Match(os.Args[1]); len(command) > 0 {
		cmd := exec.Command(shell, "-c", command)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
			Pgid:    0}
	}
}
