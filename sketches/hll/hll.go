package hll

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"math"
	"math/bits"
	"math/rand"

	"github.com/bruhng/distributed-sketching/shared"
	"github.com/spaolacci/murmur3"
)

type HLLSketch[T shared.Number] struct {
	C []int
	m uint64
	h uint32
	g uint32
}

func NewHLLSketch[T shared.Number](m uint64, seed int64) *HLLSketch[T] {
	arr := make([]int, m)

	r := rand.New(rand.NewSource(seed))
	h := r.Uint32()
	g := r.Uint32()

	return &HLLSketch[T]{C: arr, m: m, h: h, g: g}
}

func hashWithSeed(data []byte, seed uint32) uint64 {
	hash := murmur3.New64WithSeed(seed)
	hash.Write(data)
	return hash.Sum64()
}

func (hll HLLSketch[T]) Add(x T) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(x)
	if err != nil {
		fmt.Println(err)
		panic("hi i am not handled (could not convert data to bytes)")
	}
	bytes := buf.Bytes()

	hx := hashWithSeed(bytes, hll.h)
	gx := hashWithSeed(bytes, hll.g)

	C := hll.C
	m := hll.m

	chx := C[hx%m]
	zgx := bits.LeadingZeros(uint(gx))

	hll.C[hx%m] = max(chx, zgx)
}

func (hll HLLSketch[T]) Merge(hll2 HLLSketch[T]) error {
	newC := make([]int, hll.m)
	if hll.g != hll2.g || hll.h != hll2.h || hll.m != hll2.m {
		return errors.New("Missmatched parameters")
	}
	for i, x := range hll.C {
		newC[i] = max(x, hll2.C[i])
	}
	hll.C = newC
	return nil
}

func (hll HLLSketch[T]) Query() float64 {
	x := 0.0
	c := hll.C
	m := float64(hll.m)
	am := 0.7213 / (1 + 1.079*m)
	for _, cj := range c {
		x += math.Pow(2.0, -float64(cj))
	}
	return am * math.Pow(m, 2) / x
}
