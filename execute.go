package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

func (s *Shell) execute(parsedCmd ParsedCmd) {
	cmd, args := parsedCmd.Args[0], parsedCmd.Args[1:]

	stdout, cleanupOut, err := openRedirect(parsedCmd.RedirectStdout, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer cleanupOut()

	stderr, cleanupErr, err := openRedirect(parsedCmd.RedirectStderr, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer cleanupErr()

	if fn, ok := builtinCmds[cmd]; ok {
		fn(args, stdout, stderr)
	} else {
		ExecCmd(cmd, args, stdout, stderr)
	}
}

// openRedirect returns an io.Writer for the given Redirect, falling back to fallback when no file is specified. The returned cleanup func must be deferred.
func openRedirect(r Redirect, fallback io.Writer) (io.Writer, func(), error) {
	if r.File == "" {
		return fallback, func() {}, nil
	}
	flags := os.O_CREATE | os.O_WRONLY
	if r.Type == resultAppend {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}
	f, err := os.OpenFile(r.File, flags, 0644)
	if err != nil {
		return nil, nil, err
	}
	return f, func() { f.Close() }, nil
}

func ExecPipeline(cmds []ParsedCmd, finalStdout io.Writer, finalStderr io.Writer) {
	n := len(cmds)

	// Build n-1 OS pipes connecting adjacent commands.
	readers := make([]*os.File, n-1)
	writers := make([]*os.File, n-1)
	for i := range readers {
		r, w, err := os.Pipe()
		if err != nil {
			fmt.Fprintln(finalStderr, err)
			return
		}
		readers[i], writers[i] = r, w
	}

	// Track which writer indices are owned by builtin goroutines so we don't double-close them in the parent cleanup loop below.
	builtinWriter := make([]bool, n-1)

	var wg sync.WaitGroup
	processes := make([]*exec.Cmd, 0, n)

	for i, parsedCmd := range cmds {
		if len(parsedCmd.Args) == 0 {
			continue
		}

		var stdin io.Reader = os.Stdin
		if i > 0 {
			stdin = readers[i-1]
		}

		var stdout io.Writer = finalStdout
		if i < n-1 {
			stdout = writers[i]
		}

		cmdName := parsedCmd.Args[0]
		cmdArgs := parsedCmd.Args[1:]

		if fn, ok := builtinCmds[cmdName]; ok {
			// Builtins run in-process in a goroutine so pipe I/O doesn't deadlock.
			capturedFn := fn
			capturedArgs := cmdArgs
			capturedStdout := stdout
			capturedStderr := finalStderr
			writerIdx := i // index into writers[], -1 if last command

			if i < n-1 {
				builtinWriter[i] = true
			}

			wg.Go(func() {
				capturedFn(capturedArgs, capturedStdout, capturedStderr)
				// Close this command's write-end so the next command gets EOF.
				if writerIdx < n-1 {
					writers[writerIdx].Close()
				}
			})
		} else {
			c := exec.Command(cmdName, cmdArgs...)
			c.Stdin = stdin
			c.Stdout = stdout
			c.Stderr = finalStderr
			if err := c.Start(); err != nil {
				fmt.Fprintln(finalStderr, cmdName+": command not found")
				continue
			}
			processes = append(processes, c)
		}
	}

	// Close parent's write ends for external-command segments. Builtin goroutines close their own write ends, so skip those.
	for i, w := range writers {
		if !builtinWriter[i] {
			w.Close()
		}
	}
	// Always close all read ends in the parent.
	for _, r := range readers {
		r.Close()
	}

	// Wait for external processes, then builtin goroutines.
	for _, p := range processes {
		p.Wait()
	}
	wg.Wait()
}
