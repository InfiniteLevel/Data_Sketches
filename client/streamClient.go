package client

import (
	"fmt"

	pb "github.com/bruhng/distributed-sketching/proto"
	"github.com/bruhng/distributed-sketching/shared"
	"github.com/bruhng/distributed-sketching/stream"
)

func StreamClient[T shared.Number](batchsize int, dataStream stream.Stream[T], addr string, startConnection connectionStarter) {
	var reconAttempt *int = new(int)
	*reconAttempt = 0
	c, conn, err := startConnection(addr)
	if err != nil {
		fmt.Printf("Connection failed with error: %v\n", err)
		panic("could not start connection")
	}
	fmt.Printf("Starting stream client with batch size %d\n", batchsize)
	//Instantiate application-buffer
	buf := make([]T, batchsize)
	i := 0
	for data := range dataStream.Data {
		buf[i] = data
		i++

		if i%batchsize == 0 {
			//fmt.Print("Buffer full, initiating send\n")
			protoBuf := ConvertToProtoBuf(buf)

			MakeRequest(protoBuf, addr, c.MergeBufIntoASketch, conn, &c, startConnection, reconAttempt)
			buf = make([]T, batchsize)
			i = 0
		}
	}
	if conn != nil {
		conn.Close()
	}
	blackhole = buf
}

// rethink this, could not work. Only works with "always-addable" functionality of data sketches
func GetProtoBuf[T shared.Number](batchsize int, dataStream stream.Stream[T]) *pb.BufBatch {
	buf := make([]T, 0, batchsize)
	i := 0
	for data := range dataStream.Data {
		buf[i] = data
		i++

		if i%batchsize == 0 {
			//fmt.Print("Buffer full, initiating send\n")
			return ConvertToProtoBuf(buf)
		}
	}
	return nil
}

func ConvertToProtoBuf[T shared.Number](buf []T) *pb.BufBatch {
	t := fmt.Sprintf("%T", *new(T))
	protoBuf := &pb.BufBatch{Type: t}
	switch t {
	case "int":
		for _, item := range buf {
			protoBuf.Items = append(protoBuf.Items, &pb.NumericValue{Value: &pb.NumericValue_IntVal{IntVal: int64(item)}})
		}
	case "float64":
		for _, item := range buf {
			protoBuf.Items = append(protoBuf.Items, &pb.NumericValue{Value: &pb.NumericValue_FloatVal{FloatVal: float64(item)}})
		}
	default:
		panic("Type not supported")
	}
	return protoBuf
}
