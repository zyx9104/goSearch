package tree

import (
	"testing"

	"github.com/z-y-x233/goSearch/pkg/tools"
)

func TestSerialize(t *testing.T) {
	tools.Seg.LoadDictionary("../data/dict.txt")
	tr := NewTrie()
	strs := []string{"123", "asgfas", "tetetw", "111254", "111254", "111254", "asaf", "gjdfgh", "asagvsd", "asgsdg", "awtewrtg", "aszxf"}
	for _, s := range strs {
		tr.Insert(s)
	}
	data, _ := tr.Serialize()
	tr.UnSerialize(data)
	res := tr.RelatedSearch("")
	t.Log(res)

}
