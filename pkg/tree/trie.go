package tree

import (
	"container/heap"

	"github.com/z-y-x233/goSearch/pkg/tools"
)

type TrieNode struct {
	Cnt int
	Son [charSet]*TrieNode // 子节点
}

type Trie struct {
	Root *TrieNode
	Size int
	Cnt  int
}

type Query struct {
	Q     string
	Cnt   int
	Score float64
}

func NewTrie() *Trie {
	return &Trie{Root: &TrieNode{}}
}

func (t *Trie) Insert(s string) {
	bytes := []byte(s)
	u := t.Root
	for _, c := range bytes {
		if u.Son[c] == nil {
			u.Son[c] = &TrieNode{}
		}
		u = u.Son[c]
	}
	if u != t.Root {
		u.Cnt++
		t.Size++
		if u.Cnt == 1 {
			t.Cnt++
		}
	}
}

func (t *Trie) FindNode(s string) *TrieNode {
	bytes := []byte(s)
	u := t.Root
	for _, c := range bytes {
		u = u.Son[c]
		if u == nil {
			break
		}
	}
	return u
}

func (t *Trie) Walk(node *TrieNode, wordMap map[string]bool, s []byte, ch chan *Query) {
	if node == nil {
		return
	}
	if node.Cnt > 0 {
		q := string(s)
		words := tools.WordCut(q)
		score := float64(0)
		for _, word := range words {
			if wordMap[word] {
				score++
			}
		}
		Q := &Query{Q: q, Cnt: node.Cnt, Score: score}
		ch <- Q
	}
	for i := 0; i < charSet; i++ {
		u := node.Son[i]
		b := byte(i)

		t.Walk(u, wordMap, append(s, b), ch)
	}
}

func (t *Trie) RelatedSearch(q string) (res []string) {
	ch := make(chan *Query, 10000)
	wordMap := make(map[string]bool, 10)
	words := tools.WordCut(q)
	u := t.FindNode(q)
	for _, word := range words {
		wordMap[word] = true
	}
	t.Walk(u, wordMap, []byte(q), ch)
	close(ch)
	h := NewMaxHeap()
	for q := range ch {
		heap.Push(h, q)
		if h.Len() > 10 {
			heap.Pop(h)
		}
	}
	for h.Len() > 0 {
		res = append(res, heap.Pop(h).(*Query).Q)
	}
	return
}
