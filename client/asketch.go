package client

import (
	"fmt"
	"reflect"

	pb "github.com/bruhng/distributed-sketching/proto"
	"github.com/bruhng/distributed-sketching/shared"
	"github.com/bruhng/distributed-sketching/sketches/asketch"
	"github.com/bruhng/distributed-sketching/stream"
)

func ASketchClient[T shared.Number](mergeAfter int, dataStream stream.Stream[T], fieldName string, addr string, startConnection connectionStarter) {
	var reconAttempt *int = new(int)
	*reconAttempt = 0
	c, conn, err := startConnection(addr)
	if err != nil {
		fmt.Printf("Connection failed with error: %v\n", err)
		panic("could not start connection")
	}
	sketch := asketch.NewASketch[T](shared.ASketchSeed, shared.ASketchWidth, shared.ASketchDepth, shared.ASketchSlots)
	i := 0
	for data := range dataStream.Data {
		sketch.Add(data)
		i++

		if i%mergeAfter == 0 {
			protoSketch := ConvertToProtoASketch(sketch, fieldName)

			MakeRequest(protoSketch, addr, c.MergeASketch, conn, &c, startConnection, reconAttempt)
			sketch = asketch.NewASketch[T](shared.ASketchSeed, shared.ASketchWidth, shared.ASketchDepth, shared.ASketchSlots)
		}
	}
	if conn != nil {
		conn.Close()
	}
	blackhole = sketch
}

func GetASketch[T shared.Number](mergeAfter int, dataStream stream.Stream[T], fieldName string) *pb.ASketch {
	sketch := asketch.NewASketch[T](shared.ASketchSeed, shared.ASketchWidth, shared.ASketchDepth, shared.ASketchSlots)
	i := 0
	for data := range dataStream.Data {
		sketch.Add(data)
		i++

		if i%mergeAfter == 0 {
			return ConvertToProtoASketch(sketch, fieldName)
		}
	}
	return nil
}

func ConvertToProtoASketch[T shared.Number](sketch *asketch.ASketch[T], fieldName string) *pb.ASketch {
	t := fmt.Sprintf("%T", *new(T))

	// Get snapshot of the sketch
	filter, rows, seeds := sketch.Snapshot()

	protoASketch := &pb.ASketch{
		Type:  t,
		Field: fieldName,
	}
	// Convert filter entries
	for _, slot := range filter {
		var protoValue *pb.NumericValue

		if reflect.ValueOf(slot.Item).Kind() == reflect.Int {
			protoValue = &pb.NumericValue{
				Value: &pb.NumericValue_IntVal{IntVal: int64(slot.Item)},
				Type:  "int",
			}
		} else {
			protoValue = &pb.NumericValue{
				Value: &pb.NumericValue_FloatVal{FloatVal: float64(slot.Item)},
				Type:  "float64",
			}
		}

		protoEntry := &pb.ASketchFilterEntry{
			Item: protoValue,
			Old:  int64(slot.Old),
			New:  int64(slot.New),
		}

		protoASketch.Filter = append(protoASketch.Filter, protoEntry)
	}

	// Convert CountMin data
	protoCountMin := &pb.CountMin{}

	for _, row := range rows {
		protoRow := &pb.IntRow{}
		for _, val := range row {
			protoRow.Val = append(protoRow.Val, int64(val))
		}
		protoCountMin.Rows = append(protoCountMin.Rows, protoRow)
	}

	protoCountMin.Seeds = append(protoCountMin.Seeds, seeds...)
	protoASketch.CountMin = protoCountMin

	return protoASketch
}
