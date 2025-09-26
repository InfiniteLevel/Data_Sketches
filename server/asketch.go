package server

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/bruhng/distributed-sketching/proto"
	"github.com/bruhng/distributed-sketching/shared"
	"github.com/bruhng/distributed-sketching/sketches/asketch"
)

var (
	asketchStateOnce sync.Once
	asketchStateMap  sync.Map
	asketchMutex     sync.Mutex
)

func getOrCreateASketchState[T shared.Number]() *asketch.ASketch[T] {
	key := fmt.Sprintf("%T", *new(T))
	var sketch *asketch.ASketch[T]
	asketchStateOnce.Do(func() {
		asketchStateMap.Store(key, asketch.NewASketch[T](shared.ASketchSeed, shared.ASketchWidth, shared.ASketchDepth, shared.ASketchSlots))
	})
	if val, ok := asketchStateMap.Load(key); ok {
		sketch = val.(*asketch.ASketch[T])
	}
	return sketch
}

// Convert protobuf ASketch to internal ASketch
func convertProtoASToAS[T shared.Number](protoData *pb.ASketch) *asketch.ASketch[T] {
	var filter []asketch.FilterSlot[T]
	var rows [][]int
	var seeds []uint32

	// Convert filter entries
	for _, entry := range protoData.Filter {
		var item T
		switch entry.Item.Type {
		case "int":
			if intVal, ok := entry.Item.Value.(*pb.NumericValue_IntVal); ok {
				item = T(intVal.IntVal)
			}
		case "float64":
			if floatVal, ok := entry.Item.Value.(*pb.NumericValue_FloatVal); ok {
				item = T(any(floatVal.FloatVal).(T))
			}
		}

		filter = append(filter, asketch.FilterSlot[T]{
			Item: item,
			Old:  int(entry.Old),
			New:  int(entry.New),
		})
	}

	// Convert CountMin data
	for _, row := range protoData.CountMin.Rows {
		var intRow []int
		for _, val := range row.Val {
			intRow = append(intRow, int(val))
		}
		rows = append(rows, intRow)
	}

	seeds = protoData.CountMin.Seeds

	return asketch.NewASketchFromState(filter, rows, seeds)
}

// Merge the incoming ASketch into the server's ASketch state
func (s *Server) MergeASketch(_ context.Context, in *pb.ASketch) (*pb.MergeReply, error) {
	switch in.Type {
	case "int":
		asketchState := getOrCreateASketchState[int]()
		sketch := convertProtoASToAS[int](in)
		asketchMutex.Lock()
		asketchState.Merge(sketch)
		asketchMutex.Unlock()
	case "float64":
		asketchState := getOrCreateASketchState[float64]()
		sketch := convertProtoASToAS[float64](in)
		asketchMutex.Lock()
		asketchState.Merge(sketch)
		asketchMutex.Unlock()
	default:
		return nil, fmt.Errorf("%s is not supported, please submit a valid type", in.Type)
	}

	return &pb.MergeReply{Status: 0}, nil
}

// Query the ASketch state for the given value
func (s *Server) QueryASketch(_ context.Context, in *pb.NumericValue) (*pb.CountQueryReply, error) {
	switch in.Type {
	case "int":
		asketchState := getOrCreateASketchState[int]()
		val := in.GetIntVal()
		ret := asketchState.Query(int(val))
		return &pb.CountQueryReply{Res: int64(ret)}, nil
	case "float64":
		asketchState := getOrCreateASketchState[float64]()
		val := in.GetFloatVal()
		ret := asketchState.Query(float64(val))
		return &pb.CountQueryReply{Res: int64(ret)}, nil
	default:
		return nil, fmt.Errorf("%s is not supported, please submit a valid type", in.Type)
	}
}
