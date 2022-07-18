package tree

import (
	"bytes"
	"encoding/gob"
	"log"
)

func Serialize(t *Trie) []byte {
	qs := t.RelatedSearch("")
	buf := &bytes.Buffer{}
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(qs)
	if err != nil {
		log.Fatalln(err)
	}
	return buf.Bytes()
}

func UnSerialize(data []byte) *Trie {
	qs := []string{}
	decoder := gob.NewDecoder(bytes.NewBuffer(data))
	decoder.Decode(&qs)
	t := NewTrie()
	for _, q := range qs {
		t.Insert(q)
	}

	return t
}
