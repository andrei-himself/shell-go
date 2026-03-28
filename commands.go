package main

import (
	"fmt"
	"io"
	"os/exec"
)

func ExecCmd(cmd string, args []string, stdout, stderr io.Writer) {
	if _, err := exec.LookPath(cmd); err != nil {
		fmt.Fprintf(stderr, "%s: command not found\n", cmd)
		return
	}
	c := exec.Command(cmd, args...)
	c.Stdout = stdout
	c.Stderr = stderr
	c.Run()
}
