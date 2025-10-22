package server

import (
	"context"
	"fmt"

	pb "github.com/bruhng/distributed-sketching/proto"
	"github.com/bruhng/distributed-sketching/shared"
)

// Convert protoBuf and feed into internal ASketch
func convertProtoBufToBuf[T shared.Number](protoData *pb.BufBatch) []T {
	localBuf := make([]T, len(protoData.Items)) // allocate according length first

	// Convert filter entries
	for i, item := range protoData.Items {
		switch protoData.Type {
		case "int":
			if intVal, ok := item.Value.(*pb.NumericValue_IntVal); ok {
				localBuf[i] = T(intVal.IntVal)
			}
		case "float64":
			if floatVal, ok := item.Value.(*pb.NumericValue_FloatVal); ok {
				localBuf[i] = T(floatVal.FloatVal)
			}
		}
	}

	return localBuf
}

// Merge the incoming Buf into the server's ASketch state
func (s *Server) MergeBufIntoASketch(_ context.Context, in *pb.BufBatch) (*pb.MergeReply, error) {
	switch in.Type {
	case "int":
		asketchState := getOrCreateASketchState[int]()
		buf := convertProtoBufToBuf[int](in)
		asketchMutex.Lock()
		asketchState.MergeBuf(buf)
		asketchMutex.Unlock()
	case "float64":
		asketchState := getOrCreateASketchState[float64]()
		buf := convertProtoBufToBuf[float64](in)
		asketchMutex.Lock()
		asketchState.MergeBuf(buf)
		asketchMutex.Unlock()
	default:
		return nil, fmt.Errorf("%s is not supported, please submit a valid type", in.Type)
	}

	return &pb.MergeReply{Status: 0}, nil
}
