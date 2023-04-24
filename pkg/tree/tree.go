package tree

import (
	"github.com/spf13/viper"

	"github.com/z-y-x233/goSearch/pkg/log"
)

var (
	trie     *Trie
	filepath string
)

func Init() {
	trie = NewTrie()
	filepath = viper.GetString("db.searchHistory")
	log.Infoln("Load History Search")
	trie.LoadData(filepath)
}

func Close() {
	log.Infoln("Close Trie")
	log.Infoln("Save Data to", filepath)
	trie.Save(filepath)
}

func AddQuery(q string) {
	trie.Insert(q)
}

func FindRelated(q string, num int) []Search {
	if q == "" {
		return []Search{}
	}
	return trie.RelatedSearch(q, num)
}
