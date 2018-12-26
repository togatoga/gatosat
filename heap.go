package main

import (
	"fmt"
)

//Heap is a struct for deciding the priority of the decision
type Heap struct {
	data     []Var     // The content of data
	indices  []int     // The heap index of Var
	activity []float64 // The priority of each variable.
}

//NewHeap returns a pointer of Heap
func NewHeap() *Heap {
	return &Heap{}
}

//Less returns a boolean indicating whether two variables are small
func (h *Heap) Less(i, j int) bool {
	return h.activity[i] > h.activity[j]
}

//Size returns the size of data
func (h *Heap) Size() int {
	return len(h.data)
}

//Empty returns a boolean indicating whether the size of data is zero or not
func (h *Heap) Empty() bool {
	return len(h.data) == 0
}

func (h *Heap) InHeap(x Var) bool {
	return int(x) < len(h.indices) && h.indices[x] >= 0
}

func (h *Heap) Decrease(x Var) {
	if !h.InHeap(x) {
		panic(fmt.Errorf("The var is not in heap: %d", x))
	}
	h.percolateUp(h.indices[x])
}
func (h *Heap) Increase(x Var) {
	if !h.InHeap(x) {
		panic(fmt.Errorf("The var is not in heap: %d", x))
	}
	h.percolateDown(h.indices[x])
}

func (h *Heap) Activity(x Var) float64 {
	return h.activity[x]
}

func (h *Heap) Update(x Var) {
	if !h.InHeap(x) {
		h.PushBack(x)
	} else {
		h.percolateUp(h.indices[x])
		h.percolateDown(h.indices[x])
	}
}

func (h *Heap) RemoveMin() Var {
	x := h.data[0]
	h.data[0] = h.data[h.Size()-1]
	h.indices[h.data[0]] = 0
	h.indices[x] = -1
	h.data = h.data[:h.Size()-1]
	if h.Size() > 1 {
		h.percolateDown(0)
	}
	return x
}

func (h *Heap) PushBack(x Var) {
	if h.InHeap(x) {
		panic(fmt.Errorf("This var is already inserted: %v", x))
	}
	for int(x) >= len(h.indices) {
		h.indices = append(h.indices, -1)
		h.activity = append(h.activity, 0.0)
	}
	h.data = append(h.data, x)
	h.indices[x] = len(h.data) - 1
}

func (h *Heap) percolateUp(i int) {
	x := h.data[i]
	p := parentIndex(i)

	for i != 0 && h.Less(int(x), p) {
		h.indices[h.data[p]] = i
		h.data[i] = h.data[p]

		i = p
		p = parentIndex(i)
	}
	h.data[i] = x
	h.indices[x] = i
}

func (h *Heap) percolateDown(i int) {
	x := h.data[i]
	for leftIndex(i) < len(h.data) {
		var childIndex int
		if rightIndex(i) < len(h.data) && h.Less(int(h.data[rightIndex(i)]), int(h.data[leftIndex(i)])) {
			childIndex = rightIndex(i)
		} else {
			childIndex = leftIndex(i)
		}
		//no more down
		if !h.Less(int(h.data[childIndex]), int(x)) {
			break
		}
		h.data[i] = h.data[childIndex]
		h.indices[h.data[childIndex]] = i
		i = childIndex
	}
	h.data[i] = x
	h.indices[x] = i
}

func leftIndex(i int) int {
	return 2*i + 1
}

func rightIndex(i int) int {
	return 2*i + 2
}

func parentIndex(i int) int {
	return (i - 1) >> 1
}

func (s *Solver) InsertVarOrder(x Var) {
	if !s.VarOrder.InHeap(x) && s.Decision[x] {
		s.VarOrder.PushBack(x)
	}
}
