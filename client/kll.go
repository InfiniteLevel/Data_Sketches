package client

import (
	"fmt"
	"reflect"

	pb "github.com/bruhng/distributed-sketching/proto"
	"github.com/bruhng/distributed-sketching/shared"
	"github.com/bruhng/distributed-sketching/sketches/kll"
	"github.com/bruhng/distributed-sketching/stream"
	"google.golang.org/grpc"
)

type connectionStarter func(string) (pb.SketcherClient, *grpc.ClientConn, error)

func KllClient[T shared.Number](k int, mergeAfter int, dataStream stream.Stream[T], addr string, startConnection connectionStarter) {
	var reconAttempt *int = new(int)
	*reconAttempt = 0
	c, conn, err := startConnection(addr)
	if err != nil {
		fmt.Println(err)
		panic("could not start connection")
	}
	sketch := kll.NewKLLSketch[T](k)
	i := 0
	for data := range dataStream.Data {

		sketch.Add(data)
		i++

		if i%mergeAfter == 0 {
			protoSketch := ConvertToProtoKLL(sketch)
			
			MakeRequest[pb.KLLSketch](protoSketch, addr, c.MergeKll, conn, &c, startConnection, reconAttempt)
			sketch = kll.NewKLLSketch[T](k)
			i = 0
		}
	}
	if conn != nil {
		conn.Close()

	}
	blackhole = sketch
}

func GetKll[T shared.Number](k int, mergeAfter int, dataStream stream.Stream[T]) *pb.KLLSketch {
	sketch := kll.NewKLLSketch[T](k)
	i := 0
	for data := range dataStream.Data {

		sketch.Add(data)
		i++

		if i%mergeAfter == 0 {
			return ConvertToProtoKLL(sketch)
		}
	}
	return nil
}

func ConvertToProtoKLL[T shared.Number](sketch *kll.KLLSketch[T]) *pb.KLLSketch {
	t := fmt.Sprintf("%T", sketch.Sketch)[4:]
	orderedArray := &pb.KLLSketch{N: int64(sketch.N), Type: t}
	data := sketch.Sketch

	for _, row := range data {
		protoRow := &pb.NumericRow{} // Create a new row

		for _, val := range row {
			if reflect.ValueOf(val).Kind() == reflect.Int {
				protoRow.Values = append(protoRow.Values, &pb.NumericValue{
					Value: &pb.NumericValue_IntVal{IntVal: int64(val)}, // Wrap value properly
				})

			} else {
				protoRow.Values = append(protoRow.Values, &pb.NumericValue{
					Value: &pb.NumericValue_FloatVal{FloatVal: float64(val)}, // Wrap value properly
				})

			}
		}
		orderedArray.Rows = append(orderedArray.Rows, protoRow)
	}

	return orderedArray
}
