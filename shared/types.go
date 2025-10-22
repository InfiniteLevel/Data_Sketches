package shared

import "golang.org/x/exp/constraints"

type Number interface {
	constraints.Integer | constraints.Float
}

// ASketch constants
const (
	ASketchSeed  int64  = 157
	ASketchWidth uint64 = 512
	ASketchDepth int    = 4
	ASketchSlots int    = 16
)
