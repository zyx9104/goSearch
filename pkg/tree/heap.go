package tree

import "container/heap"

type MaxHeap []*Query

func (h MaxHeap) Len() int {
	return len(h)
}

func (h MaxHeap) Less(i, j int) bool {
	if h[i].Cnt == h[j].Cnt {
		return h[i].Score > h[j].Score
	}
	return h[i].Cnt > h[j].Cnt

}

func (h MaxHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *MaxHeap) Push(x any) {
	*h = append(*h, x.(*Query))
}
func (h *MaxHeap) Pop() any {
	n := len(*h) - 1
	x := (*h)[n]
	*h = (*h)[:n]
	return x
}

func NewMaxHeap() *MaxHeap {
	h := &MaxHeap{}
	heap.Init(h)
	return h
}
