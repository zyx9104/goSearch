package model

type SliceItem struct {
	Id    uint32
	Score float64
}

type Docs []*SliceItem

func (s Docs) Len() int {
	return len(s)
}
func (s Docs) Less(i, j int) bool {
	return s[i].Score > s[j].Score
}
func (s Docs) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
