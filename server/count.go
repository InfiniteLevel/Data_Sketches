package server

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/bruhng/distributed-sketching/proto"
	"github.com/bruhng/distributed-sketching/shared"
	"github.com/bruhng/distributed-sketching/sketches/count"
)

var (
	CountStateOnce sync.Once
	CountStateMap  sync.Map
	CountMutex     sync.Mutex
)

func getOrCreateCountState[T shared.Number]() *count.CountSketch[T] {
	key := fmt.Sprintf("%T", *new(T))
	var sketch *count.CountSketch[T]
	CountStateOnce.Do(func() {
		CountStateMap.Store(key, count.NewCountSketch[T](157, 100, 10))
	})
	if val, ok := CountStateMap.Load(key); ok {
		sketch = val.(*count.CountSketch[T])
	}
	return sketch
}

func convertProtoCountToCount[T shared.Number](protoData *pb.CountSketch) *count.CountSketch[T] {
	var data [][]int
	var seeds []uint32

	for _, protoRow := range protoData.Rows {
		var row []int

		for _, protoValue := range protoRow.Val {
			row = append(row, int(protoValue))
		}

		data = append(data, row)
	}
	seeds = append(seeds, protoData.Seeds...)

	return count.NewCountFromData[T](data, seeds)
}

func (s *Server) MergeCount(_ context.Context, in *pb.CountSketch) (*pb.MergeReply, error) {
	if in.Type == "int" {
		countState := getOrCreateCountState[int]()
		sketch := convertProtoCountToCount[int](in)
		CountMutex.Lock()
		countState.Merge(*sketch)
		CountMutex.Unlock()
	} else if in.Type == "float64" {
		countState := getOrCreateCountState[float64]()
		sketch := convertProtoCountToCount[float64](in)
		CountMutex.Lock()
		countState.Merge(*sketch)
		CountMutex.Unlock()
	} else {
		return nil, fmt.Errorf("%s is not supported, please submit a valid type", in.Type)
	}

	return &pb.MergeReply{Status: 0}, nil
}

func (s *Server) QueryCount(_ context.Context, in *pb.NumericValue) (*pb.CountQueryReply, error) {
	if in.Type == "int" {
		countState := getOrCreateCountState[int]()
		val := in.GetIntVal()
		ret := countState.Query(int(val))
		return &pb.CountQueryReply{Res: int64(ret)}, nil
	} else if in.Type == "float64" {
		countState := getOrCreateKllState[float64]()
		val := in.GetFloatVal()
		ret := countState.Query(float64(val))
		return &pb.CountQueryReply{Res: int64(ret)}, nil
	} else {
		return nil, fmt.Errorf("%s is not supported, please submit a valid type", in.Type)
	}
}
