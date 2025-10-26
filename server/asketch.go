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
	asketchStateMap sync.Map
	asketchMutex    sync.Mutex
)

func getOrCreateASketchState[T shared.Number](field string) *asketch.ASketch[T] {
	key := fmt.Sprintf("%s | %T", field, *new(T)) // "int" or "float64"

	if v, ok := asketchStateMap.Load(key); ok {
		return v.(*asketch.ASketch[T])
	}
	sk := asketch.NewASketch[T](shared.ASketchSeed, shared.ASketchWidth, shared.ASketchDepth, shared.ASketchSlots)
	actual, _ := asketchStateMap.LoadOrStore(key, sk)
	return actual.(*asketch.ASketch[T])
}

// Convert protobuf ASketch to internal ASketch
func convertProtoASToAS[T shared.Number](protoData *pb.ASketch) *asketch.ASketch[T] {
	var filter []asketch.FilterSlot[T]
	var rows [][]int
	var seeds []uint32

	// Convert filter entries
	for _, entry := range protoData.GetFilter() {
		var item T
		switch v := entry.GetItem().GetValue().(type) { // oneof
		case *pb.NumericValue_IntVal:
			item = T(v.IntVal)
		case *pb.NumericValue_FloatVal:
			item = T(v.FloatVal)
		default:
			continue
		}

		filter = append(filter, asketch.FilterSlot[T]{
			Item: item,
			Old:  int(entry.Old),
			New:  int(entry.New),
		})
	}

	// Convert CountMin data
	if cm := protoData.GetCountMin(); cm != nil {
		for _, row := range cm.GetRows() {
			intRow := make([]int, 0, len(row.GetVal()))
			for _, v := range row.GetVal() {
				intRow = append(intRow, int(v))
			}
			rows = append(rows, intRow)
		}
		seeds = append(seeds, cm.GetSeeds()...)
	}

	return asketch.NewASketchFromState(filter, rows, seeds)
}

// Merge the incoming ASketch into the server's ASketch state
func (s *Server) MergeASketch(_ context.Context, in *pb.ASketch) (*pb.MergeReply, error) {
	fld := ""
	fmt.Printf("[SERVER] MergeASketch type=%s filter=%d rows=%d\n", in.GetType(), len(in.GetFilter()), len(in.GetCountMin().GetRows()))
	switch in.Type {
	case "int":
		asketchState := getOrCreateASketchState[int](fld)
		sketch := convertProtoASToAS[int](in)
		asketchMutex.Lock()
		asketchState.MergeSketch(sketch)
		asketchMutex.Unlock()
	case "float64":
		asketchState := getOrCreateASketchState[float64](fld)
		sketch := convertProtoASToAS[float64](in)
		asketchMutex.Lock()
		asketchState.MergeSketch(sketch)
		asketchMutex.Unlock()

		if len(in.GetFilter()) > 0 {
			switch v := in.GetFilter()[0].GetItem().GetValue().(type) {
			case *pb.NumericValue_FloatVal:
				got := asketchState.Query(v.FloatVal)
				fmt.Printf("[SERVER][POST-MERGE] value=%.10g -> %d  sketch=%p\n", v.FloatVal, got, asketchState)
			}
		}

	default:
		return nil, fmt.Errorf("%s is not supported, please submit a valid type", in.GetType())
	}

	return &pb.MergeReply{Status: 0}, nil
}

// Query the ASketch state for the given value
func (s *Server) QueryASketch(_ context.Context, in *pb.NumericValue) (*pb.CountQueryReply, error) {
	switch v := in.GetValue().(type) {
	case *pb.NumericValue_IntVal:
		asketchState := getOrCreateASketchState[int]("")
		ret := asketchState.Query(int(v.IntVal))
		fmt.Printf("[SERVER][QUERY] type=int v=%d -> %d sketch=%p\n", v.IntVal, ret, asketchState)
		return &pb.CountQueryReply{Res: int64(ret)}, nil

	case *pb.NumericValue_FloatVal:
		asketchState := getOrCreateASketchState[float64]("")
		ret := asketchState.Query(v.FloatVal)
		fmt.Printf("[SERVER][QUERY] type=float64 v=%.10g -> %d sketch=%p\n", v.FloatVal, ret, asketchState)
		return &pb.CountQueryReply{Res: int64(ret)}, nil

	default:
		return nil, fmt.Errorf("unsupported NumericValue variant")
	}
}
func (s *Server) TopKASketch(_ context.Context, in *pb.TopKRequest) (*pb.TopKReply, error) {
	fld := ""

	switch in.GetType() {
	case "int":
		st := getOrCreateASketchState[int](fld)
		slots := st.TopK(int(in.GetK()))
		out := &pb.TopKReply{Entries: make([]*pb.TopKEntry, len(slots))}
		for i, sl := range slots {
			out.Entries[i] = &pb.TopKEntry{
				Key:     &pb.NumericValue{Value: &pb.NumericValue_IntVal{IntVal: int64(sl.Item)}},
				EstFreq: int64(sl.New),
			}
		}
		return out, nil

	case "float64":
		st := getOrCreateASketchState[float64](fld)
		slots := st.TopK(int(in.GetK()))
		out := &pb.TopKReply{Entries: make([]*pb.TopKEntry, len(slots))}
		for i, sl := range slots {
			out.Entries[i] = &pb.TopKEntry{
				Key:     &pb.NumericValue{Value: &pb.NumericValue_FloatVal{FloatVal: float64(sl.Item)}},
				EstFreq: int64(sl.New),
			}
		}
		return out, nil

	default:
		return nil, fmt.Errorf("%s is not supported, please submit a valid type", in.GetType())
	}
}
