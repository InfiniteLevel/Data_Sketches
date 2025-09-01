package kll

import (
	"cmp"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"sort"
)

type KLLSketch[T cmp.Ordered] struct {
	Sketch [][]T
	K      int
	N      int64
}

func NewKLLSketch[T cmp.Ordered](k int) *KLLSketch[T] {
	arr := make([][]T, 1)
	return &KLLSketch[T]{Sketch: arr, K: k}
}

func NewKLLFromData[T cmp.Ordered](arr [][]T, n int64, k int) *KLLSketch[T] {
	return &KLLSketch[T]{Sketch: arr, K: k, N: n}
}

func getSize(k int, h int, H int) int {
	diff := float64(H - 1 - h)
	exp := float64(k) * math.Pow(2.0/3.0, diff)
	rounded := math.Round(exp)
	ked := int(rounded)
	size := max(2, ked)

	return size
}

func (kll *KLLSketch[T]) Add(item T) {
	kll.Sketch[0] = append(kll.Sketch[0], item)
	kll.N++
	compress(kll)
}

func everyOther[T cmp.Ordered](xs []T) []T {
	if len(xs) == 0 {
		return []T{}
	}
	if len(xs) == 1 {
		return xs
	}
	return append([]T{xs[0]}, everyOther(xs[2:])...)
}

func compress[T cmp.Ordered](kll *KLLSketch[T]) {
	h := 0
	for {
		row := kll.Sketch[h]
		if len(row) >= getSize(kll.K, h, len(kll.Sketch)) {
			if len(kll.Sketch) == h+1 {
				kll.Sketch = append(kll.Sketch, make([]T, 0))
			}
			slices.Sort(row)
			even := rand.Int() % 2
			if even == 0 {
				kll.Sketch[h+1] = append(kll.Sketch[h+1], everyOther(row)...)
			} else {
				kll.Sketch[h+1] = append(kll.Sketch[h+1], everyOther(row[1:])...)
			}
			kll.Sketch[h] = make([]T, 0)
		}
		if len(kll.Sketch) == h+1 {
			return
		}
		h++

	}
}

func (kll *KLLSketch[T]) Merge(sketch KLLSketch[T]) {
	H := max(len(kll.Sketch), len(sketch.Sketch))
	diff := H - len(kll.Sketch)
	kll.Sketch = append(kll.Sketch, make([][]T, diff)...)
	for h := 0; h < len(sketch.Sketch); h++ {
		kll.Sketch[h] = append(kll.Sketch[h], sketch.Sketch[h]...)
	}
	kll.N += sketch.N
	compress(kll)
}

func (kll *KLLSketch[T]) Query(val T) int {
	sum := 0
	for h, row := range kll.Sketch {
		for _, elem := range row {
			if elem <= val {
				sum += int(math.Pow(2.0, float64(h)))
			}
		}
	}
	return sum
}

func (kll *KLLSketch[T]) QueryQuantile(phi float64) T {
	q := int(phi * float64(kll.N))

	quantileSum := 0
	sketch := make([][]T, len(kll.Sketch))
	copy(sketch, kll.Sketch)
	var smallestH int
	var smallestVal T
	for _, row := range sketch {
		sort.Slice(row, func(i, j int) bool {
			return row[i] < row[j]
		})
	}

	for quantileSum < q {
		smallestH = -1
		for h, row := range sketch {
			if len(row) == 0 {
				continue
			}
			if smallestH == -1 {
				smallestVal = row[0]
				smallestH = h
			}
			if row[0] < smallestVal {
				smallestVal = row[0]
				smallestH = h
			}
		}
		if smallestH == -1 {
			return smallestVal
		}
		sketch[smallestH] = sketch[smallestH][1:]
		quantileSum += int(math.Pow(2.0, float64(smallestH)))
	}
	return smallestVal
}

func (kll *KLLSketch[T]) Print() {
	fmt.Println("KLL sketch")
	for h := len(kll.Sketch) - 1; h >= 0; h-- {
		fmt.Println("Level ", h, "= ", kll.Sketch[h])
	}
	fmt.Println("K = ", kll.K)
	fmt.Println("N = ", kll.N)
}
