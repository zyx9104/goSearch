package tree

import (
	"github.com/spf13/viper"
)

var (
	trie *Trie
)

func Init() {
	trie = NewTrie()
	trie.LoadData(viper.GetString("db.searchHistory"))
}

func AddQuery(q string) {
	trie.Insert(q)
}

func FindRelated(q string, num int) []Search {
	return trie.RelatedSearch(q, num)
}
