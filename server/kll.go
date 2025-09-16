package server

import (
	"cmp"
	"context"
	"fmt"
	"sync"

	pb "github.com/bruhng/distributed-sketching/proto"
	"github.com/bruhng/distributed-sketching/shared"
	"github.com/bruhng/distributed-sketching/sketches/kll"
)

var (
	kllStateOnce sync.Once
	kllStateMap  sync.Map
	KllMutex     sync.Mutex
)

func getOrCreateKllState[T cmp.Ordered]() *kll.KLLSketch[T] {
	key := fmt.Sprintf("%T", *new(T))
	var sketch *kll.KLLSketch[T]
	kllStateOnce.Do(func() {
		kllStateMap.Store(key, kll.NewKLLSketch[T](200))
	})
	if val, ok := kllStateMap.Load(key); ok {
		sketch = val.(*kll.KLLSketch[T])
	}
	return sketch
}

func convertProtoKLLToKLL[T shared.Number](protoData *pb.KLLSketch) *kll.KLLSketch[T] {
	var data [][]T

	for _, protoRow := range protoData.Rows {
		var row []T

		for _, protoValue := range protoRow.Values {
			if intVal, ok := protoValue.Value.(*pb.NumericValue_IntVal); ok {
				row = append(row, T(intVal.IntVal))
			}
			if floatVal, ok := protoValue.Value.(*pb.NumericValue_FloatVal); ok {
				row = append(row, T(any(floatVal.FloatVal).(T)))
			}
		}

		data = append(data, row)
	}

	return kll.NewKLLFromData[T](data, protoData.GetN(), 200)
}

func (s *Server) MergeKll(_ context.Context, in *pb.KLLSketch) (*pb.MergeReply, error) {
	if in.Type == "int" {
		kllState := getOrCreateKllState[int]()
		sketch := convertProtoKLLToKLL[int](in)
		KllMutex.Lock()
		kllState.Merge(*sketch)
		KllMutex.Unlock()
	} else if in.Type == "float64" {
		kllState := getOrCreateKllState[float64]()
		sketch := convertProtoKLLToKLL[float64](in)
		KllMutex.Lock()
		kllState.Merge(*sketch)
		KllMutex.Unlock()
	} else {
		return nil, fmt.Errorf("%s is not supported, please submit a valid type", in.Type)
	}

	return &pb.MergeReply{Status: 0}, nil
}

func (s *Server) QueryKll(_ context.Context, in *pb.NumericValue) (*pb.QueryReturn, error) {
	if in.Type == "int" {
		kllState := getOrCreateKllState[int]()
		val := in.GetIntVal()
		ret := kllState.Query(int(val))
		return &pb.QueryReturn{N: int64(kllState.N), Phi: int64(ret)}, nil
	} else if in.Type == "float64" {
		kllState := getOrCreateKllState[float64]()
		val := in.GetFloatVal()
		ret := kllState.Query(float64(val))
		return &pb.QueryReturn{N: int64(kllState.N), Phi: int64(ret)}, nil
	} else {
		return nil, fmt.Errorf("%s is not supported, please submit a valid type", in.Type)
	}
}

func (s *Server) ReverseQueryKll(_ context.Context, in *pb.ReverseQuery) (*pb.NumericValue, error) {
	phi := in.Phi
	if in.Type == "int" {
		kllState := getOrCreateKllState[int]()
		ret := kllState.QueryQuantile(float64(phi))
		return &pb.NumericValue{Value: &pb.NumericValue_IntVal{IntVal: int64(ret)}}, nil
	} else if in.Type == "float64" {
		kllState := getOrCreateKllState[float64]()
		ret := kllState.QueryQuantile(float64(phi))
		return &pb.NumericValue{Value: &pb.NumericValue_FloatVal{FloatVal: float64(ret)}}, nil
	} else {
		return nil, fmt.Errorf("%s is not supported, please submit a valid type", in.Type)
	}
}

func (s *Server) PlotKll(_ context.Context, in *pb.PlotRequest) (*pb.PlotKllReply, error) {
	if in.Type == "int" {
		kllState := getOrCreateKllState[int]()
		numBins := int(in.GetNumBins())
		xmin := kllState.QueryQuantile(0.0)
		xmax := kllState.QueryQuantile(1.0)
		step := float64(xmax-xmin) / float64(numBins)

		splits := make([]int, numBins+1)
		for i := 0; i <= numBins; i++ {
			splits[i] = xmin + int(step*float64(i))
		}

		pmf := make([]float64, numBins)
		for i := 0; i < numBins; i++ {
			pmf[i] = float64(kllState.Query(splits[i+1]) - kllState.Query(splits[i]))
		}
		return &pb.PlotKllReply{Step: float64(step), Pmf: pmf}, nil
	} else if in.Type == "float64" {
		kllState := getOrCreateKllState[float64]()
		numBins := int(in.GetNumBins())
		xmin := kllState.QueryQuantile(0.0)
		xmax := kllState.QueryQuantile(1.0)
		step := float64(xmax-xmin) / float64(numBins)

		splits := make([]float64, numBins+1)
		for i := 0; i <= numBins; i++ {
			splits[i] = xmin + step*float64(i)
		}

		pmf := make([]float64, numBins)
		for i := 0; i < numBins; i++ {
			pmf[i] = float64(kllState.Query(splits[i+1]) - kllState.Query(splits[i]))
		}
		return &pb.PlotKllReply{Step: float64(step), Pmf: pmf}, nil

	} else {
		return nil, fmt.Errorf("%s is not supported, please submit a valid type", in.Type)
	}
}
