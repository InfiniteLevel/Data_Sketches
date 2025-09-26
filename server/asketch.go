package server

import (

	pb "github.com/bruhng/distributed-sketching/proto"
	"github.com/bruhng/distributed-sketching/shared"
	"github.com/bruhng/distributed-sketching/sketches/asketch"
)

var(
	asketchStateOnce sync.Once
	asketchStateMap sync.Map
	asketchMutex sync.Mutex
)

const(
	asketchSeed int64 = 157
	asketchWidth uint64 = 2048
	asketchDepth int = 7
	asketchSlots int = 128
)

func getOrCreateASketchState[T cmp.Ordered]() *kll.KLLSketch[T] {
	key := fmt.Sprintf("%T", *new(T))

	var sketch *asketch.ASketch[T]
	asketchStateOnce.Do(func(){
		asketchStateMap.Store(key, asketch.NewASketch[T](asketchSeed, asketchWidth, asketchDepth, asketchSlots))
	})
	if val,ok := asketchStateMap.Load(key); ok{
		sketch = val.(*asketch.ASketch[T])
	}
	return sketch
}

func convertProtoAStoAS()