package count

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"math/rand"
	"slices"

	"github.com/bruhng/distributed-sketching/shared"
	"github.com/spaolacci/murmur3"
)

type CountSketch[T shared.Number] struct {
	Sketch [][]int
	Seeds  []uint32
}

func NewCountSketch[T shared.Number](seed int64, size uint64, num_hashes int) *CountSketch[T] {
	arr := make([][]int, num_hashes)

	for i := 0; i < num_hashes; i++ {
		arr[i] = make([]int, size)
	}

	r := rand.New(rand.NewSource(seed))
	seeds := make([]uint32, num_hashes)
	for i := 0; i < num_hashes; i++ {
		seeds[i] = r.Uint32()
	}
	return &CountSketch[T]{Sketch: arr, Seeds: seeds}
}

func NewCountFromData[T shared.Number](arr [][]int, seeds []uint32) *CountSketch[T] {
	return &CountSketch[T]{Sketch: arr, Seeds: seeds}
}

func getSign(data []byte) int {
	hash := murmur3.New64()
	hash.Write(data)
	hashValue := hash.Sum64()

	if hashValue%2 == 0 {
		return 1
	}
	return -1
}

func getIndex(data []byte, seed uint32, size uint64) uint64 {
	hash := murmur3.New64WithSeed(seed)
	hash.Write(data)
	hashVal := hash.Sum64()
	return hashVal % size

}

func (cs *CountSketch[T]) Add(item T) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(item)
	if err != nil {
		fmt.Println(err)
		panic("hi i am not handled (could not convert data to bytes)")
	}
	size := uint64(len(cs.Sketch[0]))
	sign := getSign(buf.Bytes())
	for i, seed := range cs.Seeds {
		index := getIndex(buf.Bytes(), seed, size)
		cs.Sketch[i][index] += sign
	}
}

func (cs *CountSketch[T]) Query(item T) int {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(item)
	if err != nil {
		fmt.Println(err)
		panic("hi i am not handled (could not convert data to bytes)")
	}
	size := uint64(len(cs.Sketch[0]))
	sign := getSign(buf.Bytes())
	result_size := len(cs.Seeds)
	results := make([]int, result_size)
	for i, seed := range cs.Seeds {
		index := getIndex(buf.Bytes(), seed, size)
		results[i] = cs.Sketch[i][index] * sign
	}

	slices.Sort(results)
	if result_size%2 == 0 {
		return (results[result_size/2-1] + results[result_size/2]) / 2
	}
	return results[result_size/2]
}

func (cs *CountSketch[T]) QueryMean() float64 {
	es := make([]float64, len(cs.Sketch))
	esize := len(es)
	for j, row := range cs.Sketch {
		for _, k := range row {
			es[j] = es[j] + math.Pow(float64(k), 2.0)
		}
	}

	slices.Sort(es)
	if esize%2 == 0 {
		return (es[esize/2-1] + es[esize/2]) / 2
	}
	return es[esize/2]
}

func (cs *CountSketch[T]) Merge(sketch CountSketch[T]) {
	if len(cs.Sketch[0]) != len(sketch.Sketch[0]) {
		panic("Missmatched length of second dimension in merged sketches")
	}
	for i, rows := range cs.Sketch {
		for j, elems := range rows {
			cs.Sketch[i][j] = elems + sketch.Sketch[i][j]
		}
	}
}

func (cs *CountSketch[T]) Print() {
	fmt.Println("Count sketch")
	for _, row := range cs.Sketch {
		fmt.Println(row)
	}
	fmt.Println("Seeds : ", cs.Seeds)
}
