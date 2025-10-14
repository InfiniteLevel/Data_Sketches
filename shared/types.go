package shared

import "golang.org/x/exp/constraints"

type Number interface {
	constraints.Integer | constraints.Float
}

// ASketch constants
const (
	ASketchSeed  int64  = 157
	ASketchWidth uint64 = 2048
	ASketchDepth int    = 7
	ASketchSlots int    = 128
)

// Primitive buf constants
const (
	BufSize int = 1000 //number of elements in the buf
)
