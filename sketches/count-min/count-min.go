package countmin

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math/rand"

	"github.com/bruhng/distributed-sketching/shared"
	"github.com/spaolacci/murmur3"
)

type CountMin[T shared.Number] struct {
	Sketch [][]int
	Seeds  []uint32
}

// construct: width=bucket number of each row, depth=column number(hash number)
func NewCountMin[T shared.Number](seed int64, width uint64, depth int) *CountMin[T] {
	arr := make([][]int, depth)

	for i := 0; i < depth; i++ {
		arr[i] = make([]int, width)
	}

	r := rand.New(rand.NewSource(seed))
	seeds := make([]uint32, depth)
	for i := 0; i < depth; i++ {
		seeds[i] = r.Uint32()
	}
	return &CountMin[T]{Sketch: arr, Seeds: seeds}
}

// build from the existing data
func NewCountMinFromData[T shared.Number](arr [][]int, seeds []uint32) *CountMin[T] {
	return &CountMin[T]{Sketch: arr, Seeds: seeds}
}

func getIndex(data []byte, seed uint32, width uint64) uint64 {
	//create an incrementlly writable object(be called hasher) first
	hash := murmur3.New64WithSeed(seed)
	//feed the byte data into hasher
	hash.Write(data)
	//return the 64-bit hash value based on currently written byte
	return hash.Sum64() % width
}

func (cm *CountMin[T]) AddBy(item T, u int) {
	if u <= 0 {
		return
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(item)
	if err != nil {
		fmt.Println(err)
		panic("Could not convert data to bytes!")
	}
	width := uint64(len(cm.Sketch[0]))
	for i, seed := range cm.Seeds {
		index := getIndex(buf.Bytes(), seed, width)
		cm.Sketch[i][index] += u
	}
}

func (cm *CountMin[T]) Add(item T) {
	cm.AddBy(item, 1)
}

func (cm *CountMin[T]) Query(item T) int {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(item)
	if err != nil {
		fmt.Println(err)
		panic("Could not convert data to bytes")
	}
	width := uint64(len(cm.Sketch[0]))
	min := int(^uint(0) >> 1)
	for i, seed := range cm.Seeds {
		v := cm.Sketch[i][getIndex(buf.Bytes(), seed, width)]
		if v < min {
			min = v
		}
	}
	return min
}

func (cm *CountMin[T]) Merge(other CountMin[T]) {
	if len(cm.Sketch) != len(other.Sketch) || len(cm.Sketch[0]) != len(other.Sketch[0]) {
		panic("Missmatched table shape!")

	}
	for i := range cm.Seeds {
		if cm.Seeds[i] != other.Seeds[i] {
			panic("Missmatched seeds!")
		}
	}
	for i := range cm.Sketch {
		for j := range cm.Sketch[i] {
			cm.Sketch[i][j] += other.Sketch[i][j]
		}
	}
}
