package client

import (
	"fmt"
	"time"

	pb "github.com/bruhng/distributed-sketching/proto"
	"github.com/bruhng/distributed-sketching/shared"
	"github.com/bruhng/distributed-sketching/sketches/count"
	"github.com/bruhng/distributed-sketching/sketches/kll"
	"github.com/bruhng/distributed-sketching/stream"
)

func LatencyCountClient[T shared.Number](mergeAfter int, dataStream stream.Stream[T], addr string, startConnection connectionStarter) []time.Duration {
	var reconAttempt *int = new(int)
	*reconAttempt = 0
	c, conn, err := startConnection(addr)
	if err != nil {
		fmt.Println(err)
		panic("could not start connection")
	}
	var latencies []time.Duration
	sketch := count.NewCountSketch[T](157, 100, 10)
	i := 0
	for data := range dataStream.Data {
		sketch.Add(data)
		i++

		if i%mergeAfter == 0 {
			protoSketch := ConvertToProtoCount(sketch)
			prev := time.Now()
			MakeRequest(protoSketch, addr, c.MergeCount, conn, &c, startConnection, reconAttempt)
			diff := time.Since(prev)
			sketch = count.NewCountSketch[T](157, 100, 10)
			latencies = append(latencies, diff)

		}

	}
	if conn != nil {
		conn.Close()
	}
	return latencies
}

func LatencyKllClient[T shared.Number](k int, mergeAfter int, dataStream stream.Stream[T], addr string, startConnection connectionStarter) []time.Duration {
	var reconAttempt *int = new(int)
	*reconAttempt = 0
	c, conn, err := startConnection(addr)
	if err != nil {
		fmt.Println(err)
		panic("could not start connection")
	}
	var latencies []time.Duration
	sketch := kll.NewKLLSketch[T](k)
	i := 0
	for data := range dataStream.Data {

		sketch.Add(data)
		i++

		if i%mergeAfter == 0 {
			protoSketch := ConvertToProtoKLL(sketch)
			prev := time.Now()
			MakeRequest[pb.KLLSketch](protoSketch, addr, c.MergeKll, conn, &c, startConnection, reconAttempt)
			diff := time.Since(prev)
			sketch = kll.NewKLLSketch[T](k)
			latencies = append(latencies, diff)
			i = 0
		}
	}
	if conn != nil {
		conn.Close()

	}
	return latencies
}
