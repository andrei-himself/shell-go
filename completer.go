package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/chzyer/readline"
)

type trieCompleter struct {
	rl         *readline.Instance
	lastPrefix string
	tabCount   int
}

func (t *trieCompleter) Do(line []rune, pos int) ([][]rune, int) {
	input := string(line[:pos])

	// If there's a space before the cursor, we're completing a filename argument
	if idx := strings.LastIndex(input, " "); idx != -1 {
		return t.completeFile(input, idx)
	}

	return t.completeCommand(input)
}

func (t *trieCompleter) completeFile(input string, lastSpace int) ([][]rune, int) {
	token := input[lastSpace+1:]

	dir, filePrefix := ".", token
	if idx := strings.LastIndex(token, "/"); idx != -1 {
		dir = token[:idx]
		filePrefix = token[idx+1:]
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, 0
	}

	var matches []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), filePrefix) {
			name := e.Name()
			if dir != "." {
				name = dir + "/" + name
			}
			if e.IsDir() {
				name += "/"
			}
			matches = append(matches, name)
		}
	}
	sort.Strings(matches)

	if len(matches) == 0 {
		t.rl.Terminal.Bell()
		return nil, 0
	}

	if len(matches) == 1 {
		suffix := matches[0][len(token):]
		// trailing char is already / or will need a space for files
		if !strings.HasSuffix(matches[0], "/") {
			suffix += " "
		}
		return [][]rune{[]rune(suffix)}, len([]rune(token))
	}

	common := longestCommonPrefix(matches)
	if common != token {
		t.lastPrefix = common
		suffix := common[len(token):]
		return [][]rune{[]rune(suffix)}, len([]rune(token))
	}

	t.tabCount++
	if t.tabCount == 1 {
		t.rl.Terminal.Bell()
		return nil, 0
	}
	t.tabCount = 0
	fmt.Fprintf(t.rl.Terminal, "\n%s\n", strings.Join(matches, "  "))
	t.rl.Refresh()
	return nil, 0
}

func (t *trieCompleter) completeCommand(prefix string) ([][]rune, int) {
	options := builtinTrie.PrefixWords(prefix)
	sort.Strings(options)

	if prefix != t.lastPrefix {
		t.tabCount = 0
		t.lastPrefix = prefix
	}

	if len(options) == 0 {
		t.rl.Terminal.Bell()
		t.tabCount = 0
		return nil, 0
	}

	if len(options) == 1 {
		t.tabCount = 0
		suffix := options[0][len(prefix):] + " "
		return [][]rune{[]rune(suffix)}, len([]rune(prefix))
	}

	common := longestCommonPrefix(options)
	if common != prefix {
		t.tabCount = 0
		t.lastPrefix = common
		suffix := common[len(prefix):]
		return [][]rune{[]rune(suffix)}, len([]rune(prefix))
	}

	t.tabCount++
	if t.tabCount == 1 {
		t.rl.Terminal.Bell()
		return nil, 0
	}
	t.tabCount = 0
	fmt.Fprintf(t.rl.Terminal, "\n%s\n", strings.Join(options, "  "))
	t.rl.Refresh()
	return nil, 0
}

func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	prefix := strs[0]
	for _, s := range strs[1:] {
		for !strings.HasPrefix(s, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}
	return prefix
}
