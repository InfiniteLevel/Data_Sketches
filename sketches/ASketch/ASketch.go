package asketch

import (
	"sort"

	"github.com/bruhng/distributed-sketching/shared"
	countmin "github.com/bruhng/distributed-sketching/sketches/count-min"
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

type FilterSlot[T shared.Number] struct {
	Item T
	Old  int
	New  int
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

func (a *ASketch[T]) Merge(other *ASketch[T]) {
	if other == nil {
		return
	}
	a.cms.Merge(*other.cms)
	for _, slot := range other.filter {
		if slot.new < 0 {
			continue
		}
		a.AddBy(slot.it, slot.new)
	}
}

func NewASketchFromState[T shared.Number](filter []FilterSlot[T], rows [][]int, seeds []uint32) *ASketch[T] {
	f := make([]aCount[T], len(filter))
	for i, slot := range filter {
		f[i] = aCount[T]{
			it:  slot.Item,
			old: slot.Old,
			new: slot.New,
		}
	}
	cm := countmin.NewCountMinFromData[T](rows, seeds)
	return &ASketch[T]{
		filter: f,
		cms:    cm,
	}
}

func (a *ASketch[T]) Snapshot() ([]FilterSlot[T], [][]int, []uint32) {
	filterCopy := make([]FilterSlot[T], len(a.filter))
	for i, slot := range a.filter {
		filterCopy[i] = FilterSlot[T]{
			Item: slot.it,
			Old:  slot.old,
			New:  slot.new,
		}
	}
	rowsCopy := make([][]int, len(a.cms.Sketch))
	for i := range a.cms.Sketch {
		rowsCopy[i] = append([]int(nil), a.cms.Sketch[i]...)
	}
	seedsCopy := append([]uint32(nil), a.cms.Seeds...)
	return filterCopy, rowsCopy, seedsCopy
}

func (a *ASketch[T]) FilterSnapshot() []FilterSlot[T] {
	out := make([]FilterSlot[T], 0, len(a.filter))
	for _, s := range a.filter {
		if s.new >= 0 {
			out = append(out, FilterSlot[T]{Item: s.it, Old: s.old, New: s.new})
		}
	}
	return out
}

func (a *ASketch[T]) TopK(k int) []FilterSlot[T] {
	snap := a.FilterSnapshot()
	sort.Slice(snap, func(i, j int) bool { return snap[i].New > snap[j].New })
	if k > len(snap) {
		k = len(snap)
	}
	return snap[:k]
}
