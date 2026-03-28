package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// CommandFunc is the signature every builtin command must satisfy.
type CommandFunc func(args []string, stdout, stderr io.Writer)

// builtinCmds is the single source of truth for all builtin commands. Adding a new builtin only requires adding an entry here. Populated in init() to avoid an initialisation cycle with TypeCmd.
var builtinCmds map[string]CommandFunc

func init() {
	builtinCmds = map[string]CommandFunc{
		"exit": func(_ []string, _, _ io.Writer) { close(exitShell) },
		"echo": func(args []string, stdout, _ io.Writer) {
			fmt.Fprintln(stdout, strings.Join(args, " "))
		},
		"type":    func(args []string, stdout, _ io.Writer) { TypeCmd(args, stdout) },
		"pwd":     func(_ []string, stdout, stderr io.Writer) { PwdCmd(stdout, stderr) },
		"cd":      func(args []string, _, stderr io.Writer) { CdCmd(args, stderr) },
		"history": func(args []string, stdout, stderr io.Writer) { HistoryCmd(args, stdout, stderr) },
	}
}

func TypeCmd(args []string, stdout io.Writer) {
	if len(args) == 0 {
		return
	}
	name := args[0]
	if _, ok := builtinCmds[name]; ok {
		fmt.Fprintln(stdout, name, "is a shell builtin")
		return
	}
	if path, err := exec.LookPath(name); err == nil {
		fmt.Fprintf(stdout, "%s is %s\n", name, path)
		return
	}
	fmt.Fprintf(stdout, "%s: not found\n", name)
}

func PwdCmd(stdout, stderr io.Writer) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stderr, "pwd:", err)
		return
	}
	fmt.Fprintln(stdout, wd)
}

func CdCmd(args []string, stderr io.Writer) {
	path := "~"
	if len(args) > 0 && args[0] != "" {
		path = args[0]
	}
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(stderr, "cd:", err)
			return
		}
		if len(path) > 1 {
			path = filepath.Join(home, path[1:])
		} else {
			path = home
		}
	}
	if err := os.Chdir(filepath.Clean(path)); err != nil {
		fmt.Fprintf(stderr, "cd: %s: No such file or directory\n", path)
	}
}

// HistoryCmd implements the history builtin.
//
//	history          – print all entries
//	history <n>      – print the last n entries
//	history -r <file>– append entries from file into in-memory history
func HistoryCmd(args []string, stdout, stderr io.Writer) {
	// -a <file>: append only new in-memory entries (since last -a) to file
	if len(args) >= 2 && args[0] == "-a" {
		newEntries := historyEntries[lastAppendIdx:]
		if len(newEntries) == 0 {
			return
		}
		f, err := os.OpenFile(args[1], os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(stderr, "history: %s: %v\n", args[1], err)
			return
		}
		defer f.Close()
		for _, entry := range newEntries {
			fmt.Fprintln(f, entry)
		}
		lastAppendIdx = len(historyEntries)
		return
	}

	// -w <file>: write in-memory history to a file (overwrite)
	if len(args) >= 2 && args[0] == "-w" {
		var sb strings.Builder
		for _, entry := range historyEntries {
			sb.WriteString(entry)
			sb.WriteByte('\n')
		}
		if err := os.WriteFile(args[1], []byte(sb.String()), 0644); err != nil {
			fmt.Fprintf(stderr, "history: %s: %v\n", args[1], err)
		}
		return
	}

	// -r <file>: read (append) history from a file
	if len(args) >= 2 && args[0] == "-r" {
		data, err := os.ReadFile(args[1])
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				fmt.Fprintf(stderr, "history: %s: %v\n", args[1], err)
			}
			return
		}
		for line := range strings.SplitSeq(string(data), "\n") {
			line = strings.TrimRight(line, "\r")
			if line != "" {
				historyEntries = append(historyEntries, line)
			}
		}
		return
	}

	// Optional numeric argument: show last n entries.
	start := 0
	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 0 {
			fmt.Fprintf(stderr, "history: %s: numeric argument required\n", args[0])
			return
		}
		if n < len(historyEntries) {
			start = len(historyEntries) - n
		}
	}
	for i, entry := range historyEntries[start:] {
		fmt.Fprintf(stdout, "    %d  %s\n", start+i+1, entry)
	}
}

// exitShell is closed by the exit builtin to signal run() to return cleanly, allowing deferred functions in main() (e.g. HISTFILE save) to execute.
var exitShell = make(chan struct{})
var historyEntries []string

// lastAppendIdx tracks how many entries have already been written by history -a so that subsequent -a calls only append the new ones.
var lastAppendIdx int
