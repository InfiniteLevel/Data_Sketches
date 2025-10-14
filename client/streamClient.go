package client

import (
	"fmt"
	"reflect"

	pb "github.com/bruhng/distributed-sketching/proto"
	"github.com/bruhng/distributed-sketching/shared"
	"github.com/bruhng/distributed-sketching/stream"
)

func StreamClient[T shared.Number](dataStream stream.Stream[T], addr string, startConnection connectionStarter) {
	var reconAttempt *int = new(int)
	*reconAttempt = 0
	c, conn, err := startConnection(addr)
	if err != nil {
		fmt.Printf("Connection failed with error: %v\n", err)
		panic("could not start connection")
	}

	//Instantiate application-buffer
	buf := make([]T, shared.BufSize)
	i := 0
	for data := range dataStream.Data {
		buf[i] = data
		i++

		if i%shared.BufSize == 0 {
			protoBuf := ConvertToProtoBuf(buf)

			MakeRequest(protoBuf, addr, c.MergeBufIntoASketch, conn, &c, startConnection, reconAttempt)
			buf = make([]T, shared.BufSize)
			i = 0
		}
	}
	if conn != nil {
		conn.Close()
	}
	blackhole = buf
}

// rethink this, could not work. Only works with "always-addable" functionality of data sketches
func GetBuf[T shared.Number](dataStream stream.Stream[T]) *pb.BufBatch {
	buf := make([]T, 0, shared.BufSize)
	i := 0
	for data := range dataStream.Data {
		buf[i] = data
		i++

		if i%shared.BufSize == 0 {
			return ConvertToProtoBuf(buf)
		}
	}
	return nil
}

func ConvertToProtoBuf[T shared.Number](buf []T) *pb.BufBatch {
	t := fmt.Sprintf("%T", *new(T))
	bufCopy := make([]T, len(buf))
	copy(bufCopy, buf)

	protoBuf := &pb.BufBatch{Type: t}

	_ = reflect.ValueOf(bufCopy)
	return protoBuf
}
