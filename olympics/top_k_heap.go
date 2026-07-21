package main

import (
	"container/heap"
	"maps"
	"slices"
)

type minHeap []*orderedRecord

type orderedRecord struct {
	Athlete string `json:"athlete"`
	Country string `json:"country"`
	Medals  medals `json:"medals"`
}

func (h minHeap) Len() int { return len(h) }

func (h minHeap) Less(i, j int) bool {
	if h[i].Medals.Gold != h[j].Medals.Gold {
		return h[i].Medals.Gold < h[j].Medals.Gold
	}
	if h[i].Medals.Silver != h[j].Medals.Silver {
		return h[i].Medals.Silver < h[j].Medals.Silver
	}
	if h[i].Medals.Bronze != h[j].Medals.Bronze {
		return h[i].Medals.Bronze < h[j].Medals.Bronze
	}
	if h[i].Athlete != "" && h[j].Athlete != "" {
		return h[i].Athlete > h[j].Athlete
	}
	return h[i].Country > h[j].Country
}

func (h minHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *minHeap) Push(x any) {
	*h = append(*h, x.(*orderedRecord))
}

func (h *minHeap) Pop() any {
	old := *h
	x := old[h.Len()-1]
	*h = old[:h.Len()-1]
	return x
}

func getTopK(athletes []*orderedRecord, k int) []*orderedRecord {
	buf := make(minHeap, 0, k+1)
	h := &buf
	heap.Init(h)

	for _, a := range athletes {
		heap.Push(h, a)
		if h.Len() > k {
			heap.Pop(h)
		}
	}

	res := make([]*orderedRecord, h.Len())
	for h.Len() > 0 {
		res[h.Len()-1] = heap.Pop(h).(*orderedRecord)
	}

	return res
}

func getMapValues(m map[string]*orderedRecord) []*orderedRecord {
	return slices.Collect(maps.Values(m))
}
