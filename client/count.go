package client

import (
	"fmt"

	pb "github.com/bruhng/distributed-sketching/proto"
	"github.com/bruhng/distributed-sketching/shared"
	"github.com/bruhng/distributed-sketching/sketches/count"
	"github.com/bruhng/distributed-sketching/stream"
)

var blackhole interface{}

func CountClient[T shared.Number](mergeAfter int, dataStream stream.Stream[T], addr string, startConnection connectionStarter) {
	var reconAttempt *int = new(int)
	*reconAttempt = 0
	c, conn, err := startConnection(addr)
	if err != nil {
		fmt.Println(err)
		panic("could not start connection")
	}
	sketch := count.NewCountSketch[T](157, 100, 10)
	i := 0
	for data := range dataStream.Data {
		sketch.Add(data)
		i++

		if i%mergeAfter == 0 {
			protoSketch := ConvertToProtoCount(sketch)

			MakeRequest(protoSketch, addr, c.MergeCount, conn, &c, startConnection, reconAttempt)
			sketch = count.NewCountSketch[T](157, 100, 10)
		}

	}
	if conn != nil {
		conn.Close()
	}
	blackhole = sketch
}

func GetCount[T shared.Number](mergeAfter int, dataStream stream.Stream[T]) *pb.CountSketch {
	sketch := count.NewCountSketch[T](157, 100, 10)
	i := 0
	for data := range dataStream.Data {

		sketch.Add(data)
		i++

		if i%mergeAfter == 0 {
			return ConvertToProtoCount(sketch)
		}
	}
	return nil
}

func ConvertToProtoCount[T shared.Number](sketch *count.CountSketch[T]) *pb.CountSketch {
	t := fmt.Sprintf("%T", sketch.Sketch)[4:]
	protoArray := &pb.CountSketch{Type: t}
	data := sketch.Sketch
	seeds := sketch.Seeds

	for _, row := range data {
		protoRow := &pb.IntRow{} // Create a new row

		for _, val := range row {
			protoRow.Val = append(protoRow.Val, int64(val))
		}
		protoArray.Rows = append(protoArray.Rows, protoRow)
	}

	protoArray.Seeds = append(protoArray.Seeds, seeds...)
	return protoArray
}
