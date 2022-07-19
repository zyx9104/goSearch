package tree

import (
	"fmt"
	"testing"

	"github.com/z-y-x233/goSearch/pkg/tools"
)

func TestIrie(t *testing.T) {
	tools.Seg.LoadDictionary("../data/dict.txt")
	tr := NewTrie()
	tr.Insert("12")
	s := tr.RelatedSearch("", tr.Size)
	fmt.Println(s)
}
