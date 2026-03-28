package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

type Shell struct {
	rl *readline.Instance
}

func main() {
	initCompletion()
	rl, err := readline.NewEx(&readline.Config{Prompt: "$ "})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	rl.Config.AutoComplete = &trieCompleter{rl: rl}
	defer rl.Close()

	// Load history from HISTFILE on startup.
	if histfile := os.Getenv("HISTFILE"); histfile != "" {
		HistoryCmd([]string{"-r", histfile}, io.Discard, os.Stderr)
	}

	// Save history to HISTFILE on exit.
	defer func() {
		if histfile := os.Getenv("HISTFILE"); histfile != "" {
			HistoryCmd([]string{"-w", histfile}, io.Discard, os.Stderr)
		}
	}()

	(&Shell{rl: rl}).run()
}

func (s *Shell) run() {
	for {
		line, err := s.rl.Readline()
		if err != nil {
			// EOF / Ctrl-D: return so deferred funcs in main() run.
			return
		}
		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		// Record every non-empty command in history.
		historyEntries = append(historyEntries, input)

		segments := splitPipeline(input)
		if len(segments) == 1 {
			parsed := parseArgs(input)
			if len(parsed.Args) > 0 {
				s.execute(parsed)
			}
		} else {
			var cmds []ParsedCmd
			for _, seg := range segments {
				parsed := parseArgs(strings.TrimSpace(seg))
				if len(parsed.Args) > 0 {
					cmds = append(cmds, parsed)
				}
			}
			if len(cmds) > 0 {
				ExecPipeline(cmds, os.Stdout, os.Stderr)
			}
		}

		// Check if the exit builtin was called.
		select {
		case <-exitShell:
			return
		default:
		}
	}
}
