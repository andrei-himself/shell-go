package main

import (
	"os"
	"path/filepath"

	trie "github.com/andrei-himself/gorsh/internal"
)

var builtinTrie *trie.Trie

func initCompletion() {
	t := trie.New()
	for cmd := range builtinCmds {
		t.Insert(cmd)
	}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if info, err := e.Info(); err == nil && info.Mode()&0111 != 0 {
				t.Insert(e.Name())
			}
		}
	}
	builtinTrie = &t
}
