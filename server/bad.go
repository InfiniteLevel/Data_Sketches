package server

import (
	"cmp"
	"context"
	"fmt"
	"sync"

	pb "github.com/bruhng/distributed-sketching/proto"
	"github.com/bruhng/distributed-sketching/shared"
	"github.com/bruhng/distributed-sketching/sketches/count"
	"github.com/bruhng/distributed-sketching/sketches/kll"
)

var (
	badKllStateOnce   sync.Once
	badKllStateMap    sync.Map
	badKllMutex       sync.Mutex
	badCountStateOnce sync.Once
	badCountStateMap  sync.Map
	badCountMutex     sync.Mutex
)

func getOrCreateBadKllState[T cmp.Ordered]() *kll.KLLSketch[T] {
	key := fmt.Sprintf("%T", *new(T))
	var sketch *kll.KLLSketch[T]
	badKllStateOnce.Do(func() {
		badKllStateMap.Store(key, kll.NewKLLSketch[T](200))
	})
	if val, ok := badKllStateMap.Load(key); ok {
		sketch = val.(*kll.KLLSketch[T])
	}
	return sketch
}

func (s *Server) BadKll(_ context.Context, in *pb.BadArray) (*pb.MergeReply, error) {
	if in.Type == "int" {
		sketch := getOrCreateBadKllState[int]()
		badKllMutex.Lock()
		for _, val := range in.Arr.GetValues() {
			sketch.Add(int(val.GetIntVal()))
		}
		badKllMutex.Unlock()
	} else if in.Type == "float64" {
		sketch := getOrCreateBadKllState[float64]()
		badKllMutex.Lock()
		for _, val := range in.Arr.GetValues() {
			sketch.Add(val.GetFloatVal())
		}
		badKllMutex.Unlock()

	} else {
		return nil, fmt.Errorf("%s is not supported, please submit a valid type", in.Type)
	}
	return &pb.MergeReply{Status: 0}, nil
}
func getOrCreateBadCountState[T shared.Number]() *count.CountSketch[T] {
	key := fmt.Sprintf("%T", *new(T))
	var sketch *count.CountSketch[T]
	badCountStateOnce.Do(func() {
		badCountStateMap.Store(key, count.NewCountSketch[T](157, 100, 10))
	})
	if val, ok := badCountStateMap.Load(key); ok {
		sketch = val.(*count.CountSketch[T])
	}
	return sketch
}

func (s *Server) BadCount(_ context.Context, in *pb.BadArray) (*pb.MergeReply, error) {
	if in.Type == "int" {
		sketch := getOrCreateBadCountState[int]()
		badCountMutex.Lock()
		for _, val := range in.Arr.GetValues() {
			sketch.Add(int(val.GetIntVal()))
		}
		badCountMutex.Unlock()
	} else if in.Type == "float64" {
		sketch := getOrCreateBadCountState[float64]()
		badCountMutex.Lock()
		for _, val := range in.Arr.GetValues() {
			sketch.Add(val.GetFloatVal())
		}
		badCountMutex.Unlock()

	} else {
		return nil, fmt.Errorf("%s is not supported, please submit a valid type", in.Type)
	}
	return &pb.MergeReply{Status: 0}, nil
}
