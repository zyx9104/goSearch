package tree

import (
	"container/heap"

	"github.com/z-y-x233/goSearch/pkg/logger"
	"github.com/z-y-x233/goSearch/pkg/tools"
)

type TrieNode struct {
	Cnt int

	Son map[byte]*TrieNode // 子节点
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
	return &Trie{Root: &TrieNode{Son: make(map[byte]*TrieNode, 5)}}
}

func (t *Trie) Insert(s string) {
	t.InsertQuery(Search{Q: s, Cnt: 1})
}

func (t *Trie) InsertQuery(s Search) {
	bytes := []byte(s.Q)
	u := t.Root
	for _, c := range bytes {
		if u.Son[c] == nil {
			u.Son[c] = &TrieNode{Son: make(map[byte]*TrieNode, 5)}
		}
		u = u.Son[c]
	}
	if u != t.Root {
		if u.Cnt == 0 {
			t.Cnt++
		}
		u.Cnt += s.Cnt
		t.Size += s.Cnt
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
	for k, v := range node.Son {
		t.Walk(v, wordMap, append(s, k), ch)
	}
}

func (t *Trie) RelatedSearch(q string, num int) (res []Search) {
	ch := make(chan *Query, t.Size*2)
	wordMap := make(map[string]bool, 10)
	words := tools.WordCut(q)
	u := t.FindNode(q)
	if u == nil {
		return
	}
	logger.Debug("find node!")
	for _, word := range words {
		wordMap[word] = true
	}
	logger.Debug("walk start!")

	t.Walk(u, wordMap, []byte(q), ch)
	logger.Debug("walk done!")

	close(ch)
	h := NewMaxHeap()
	for q := range ch {
		heap.Push(h, q)
		if h.Len() > num {
			heap.Pop(h)
		}
	}
	for h.Len() > 0 {
		f := heap.Pop(h).(*Query)
		se := Search{Q: f.Q, Cnt: f.Cnt}
		res = append(res, se)
	}
	return
}
