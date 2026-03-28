package trie

type Trie struct {
	root *trieNode
}
type trieNode struct {
	children map[rune]*trieNode
	isEnd    bool
}

func New() Trie {
	return Trie{&trieNode{children: make(map[rune]*trieNode)}}
}

func (this *Trie) Insert(word string) {
	curr := this.root
	for _, char := range word {
		if curr.children[char] == nil {
			curr.children[char] = &trieNode{children: make(map[rune]*trieNode)}
		}
		curr = curr.children[char]
	}
	curr.isEnd = true
}

func (this *Trie) Search(word string) bool {
	curr := this.root
	for _, char := range word {
		if curr.children[char] == nil {
			return false
		}
		curr = curr.children[char]
	}
	return curr.isEnd
}

func (this *Trie) StartsWith(prefix string) bool {
	curr := this.root
	for _, char := range prefix {
		if curr.children[char] == nil {
			return false
		}
		curr = curr.children[char]
	}
	return true
}

func (this *Trie) PrefixWords(prefix string) []string {
	curr := this.root
	for _, char := range prefix {
		if curr.children[char] == nil {
			return nil
		}
		curr = curr.children[char]
	}
	var words []string
	collectWords(curr, prefix, &words)
	return words
}

func collectWords(node *trieNode, current string, words *[]string) {
	if node.isEnd {
		*words = append(*words, current)
	}
	for char, child := range node.children {
		collectWords(child, current+string(char), words)
	}
}
