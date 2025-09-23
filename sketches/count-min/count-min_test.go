package countmin

import (
	"math"
	"reflect"
	"testing"
	"time"
)

func newCM(seed int64) *CountMin[uint64] {
	return NewCountMin[uint64](seed, 1<<12, 4) // width=4096, depth=4
}

func TestEmptyIsZero(t *testing.T) {
	cm := newCM(1)
	if got := cm.Query(12345); got != 0 {
		t.Fatalf("empty query should be 0, got=%d", got)
	}
}

func TestMonotonicAndUpperBound(t *testing.T) {
	cm := newCM(2)
	k := uint64(42)
	prev := 0
	trueCnt := 0
	for i := 0; i < 2000; i++ {
		cm.Add(k)
		trueCnt++
		cur := cm.Query(k)
		if cur < prev {
			t.Fatalf("non-decreasing violated: prev=%d cur=%d at i=%d", prev, cur, i)
		}
		prev = cur
	}
	if prev < trueCnt {
		t.Fatalf("upper bound violated: est=%d truth=%d", prev, trueCnt)
	}
}

func TestMergeEqualsUnion(t *testing.T) {
	seed := int64(3)
	a := newCM(seed)
	b := newCM(seed)

	// 构造两条子流（有交集）
	for i := 0; i < 5000; i++ {
		a.Add(uint64(i % 2000))
	}
	for i := 0; i < 7000; i++ {
		b.Add(uint64((i + 7) % 2000))
	}

	// “把两条流合在一起喂一份”
	union := newCM(seed)
	for i := 0; i < 5000; i++ {
		union.Add(uint64(i % 2000))
	}
	for i := 0; i < 7000; i++ {
		union.Add(uint64((i + 7) % 2000))
	}

	// 合并
	a.Merge(*b)

	if !reflect.DeepEqual(a.Sketch, union.Sketch) {
		t.Fatalf("merged matrix != union matrix")
	}

}

func TestMergeCommutativeAndAssociative(t *testing.T) {
	seed := time.Now().UnixNano()
	A1, B1, C1 := newCM(seed), newCM(seed), newCM(seed)
	A2, B2, C2 := newCM(seed), newCM(seed), newCM(seed)

	for i := 0; i < 10000; i++ {
		k := uint64(i % 2500)
		A1.Add(k)
		A2.Add(k)
	}
	for i := 0; i < 8000; i++ {
		k := uint64((i + 7) % 2500)
		B1.Add(k)
		B2.Add(k)
	}
	for i := 0; i < 6000; i++ {
		k := uint64((i + 13) % 2500)
		C1.Add(k)
		C2.Add(k)
	}

	// 交换律：A+B == B+A
	X1 := newCM(seed)
	*X1 = *A1
	X1.Merge(*B1)
	Y1 := newCM(seed)
	*Y1 = *B1
	Y1.Merge(*A1)
	if !reflect.DeepEqual(X1.Sketch, Y1.Sketch) {
		t.Fatalf("A+B != B+A")
	}

	// 结合律：(A+B)+C == A+(B+C)
	L := newCM(seed)
	*L = *A1
	L.Merge(*B1)
	L.Merge(*C1)
	R := newCM(seed)
	*R = *B2
	R.Merge(*C2)
	R.Merge(*A2)
	if !reflect.DeepEqual(L.Sketch, R.Sketch) {
		t.Fatalf("(A+B)+C != A+(B+C)")
	}
}

func TestProbabilisticErrorBound(t *testing.T) {
	// 理论：ε ≈ e/width；误差上界 ≈ ε * N（以概率 ≥ 1-δ）
	seed := int64(7)
	width := uint64(1 << 12) // 4096
	depth := 4
	cm := NewCountMin[uint64](seed, width, depth)

	N := 0
	truth := map[uint64]int{}
	for i := 0; i < 200000; i++ {
		k := uint64(i % 50000) // 5w 个键，长尾
		cm.Add(k)
		truth[k]++
		N++
	}
	eps := math.E / float64(width)
	bound := int(3.0 * eps * float64(N)) // 放宽到 3εN，更稳

	viol := 0
	for k, v := range truth {
		est := cm.Query(k)
		if est < v {
			t.Fatalf("upper bound violated for key=%d est=%d truth=%d", k, est, v)
		}
		if est-v > bound {
			viol++
		}
	}
	// 允许极少数键超过宽松上界（概率事件）
	if viol > 5 {
		t.Fatalf("too many keys exceed bound: %d", viol)
	}
}
