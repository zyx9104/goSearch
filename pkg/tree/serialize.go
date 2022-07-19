package tree

import (
	"bytes"
	"encoding/gob"

	"github.com/z-y-x233/goSearch/pkg/tools"
)

func (t *Trie) Serialize() ([]byte, error) {
	qs := t.RelatedSearch("")
	buf := &bytes.Buffer{}
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(qs)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (t *Trie) UnSerialize(data []byte) error {
	qs := []string{}
	decoder := gob.NewDecoder(bytes.NewBuffer(data))
	err := decoder.Decode(&qs)
	if err != nil {
		return err
	}
	for _, q := range qs {
		t.Insert(q)
	}
	return nil
}

func (t *Trie) LoadData(path string) error {
	data, err := tools.ReadBytes(path)
	if err != nil {
		return nil
	}
	err = t.UnSerialize(data)
	return err
}

func (t *Trie) Save(path string) error {
	data, err := t.Serialize()
	if err != nil {
		return nil
	}
	err = tools.WriteBytes(data, path)
	return err
}
