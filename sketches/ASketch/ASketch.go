package asketch

import (
	"github.com/bruhng/distributed-sketching/shared"
	"github.com/bruhng/distributed-sketching/sketches/count-min"
)

type aCount[T shared.Number] struct {
	it  T
	old int
	new int
}

type ASketch[T shared.Number] struct {
	filter []aCount[T]
	cms    *countmin.CountMin[T]
}

func NewASketch[T shared.Number](seed int64, width uint64, depth int, m int) *ASketch[T] {
	if m <= 0 {
		panic("ASketch requires m > 0")
	}
	f := make([]aCount[T], m)
	for i := range f {
		f[i].new = -1
	}
	return &ASketch[T]{
		filter: f,
		cms:    countmin.NewCountMin[T](seed, width, depth),
	}
}

// check if x exist in filter, if in, return index, or return -1 means not find
func (a *ASketch[T]) getIndex(x T) int {
	for i := range a.filter {
		if a.filter[i].new >= 0 && a.filter[i].it == x {
			return i
		}
	}
	return -1
}

// find the first empty slot in filter(used for add function)
func (a *ASketch[T]) firstEmpty() int {
	for i := range a.filter {
		if a.filter[i].new < 0 {
			return i
		}
	}
	return -1
}

// When filter is full, to find the minimal value slot
func (a *ASketch[T]) argMinNew() int {
	minIdx := -1
	minVal := 0
	for i := range a.filter {
		if a.filter[i].new < 0 {
			continue
		}
		if minIdx == -1 || minVal > a.filter[i].new {
			minIdx = i
			minVal = a.filter[i].new
		}
	}
	return minIdx
}

func (a *ASketch[T]) AddBy(x T, u int) {
	if u <= 0 {
		return
	}

	if index := a.getIndex(x); index >= 0 {
		a.filter[index].new += u
		return
	}

	if index := a.firstEmpty(); index >= 0 {
		a.filter[index] = aCount[T]{it: x, old: 0, new: u}
		return
	}

	a.cms.AddBy(x, u)
	est := a.cms.Query(x)

	minIdx := a.argMinNew()
	minSlot := a.filter[minIdx]

	if est > minSlot.new {
		delta := minSlot.new - minSlot.old
		if delta > 0 {
			a.cms.AddBy(minSlot.it, delta)
		}
		a.filter[minIdx] = aCount[T]{it: x, old: est, new: est}
	}
}

func (a *ASketch[T]) Add(k T) {
	a.AddBy(k, 1)
}

func (a *ASketch[T]) Query(x T) int {
	if index := a.getIndex(x); index >= 0 {
		return a.filter[index].new
	}
	return a.cms.Query(x)
}
