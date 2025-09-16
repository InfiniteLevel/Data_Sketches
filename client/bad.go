package client

import (
	"fmt"
	"reflect"

	pb "github.com/bruhng/distributed-sketching/proto"
	"github.com/bruhng/distributed-sketching/shared"
	"github.com/bruhng/distributed-sketching/stream"
	"google.golang.org/protobuf/proto"
)

func BadKllClient[T shared.Number](mergeAfter int, dataStream stream.Stream[T], addr string, startConnection connectionStarter) {
	var reconAttempt *int = new(int)
	*reconAttempt = 0
	c, conn, err := startConnection(addr)
	if err != nil {
		fmt.Println(err)
		panic("could not start connection")
	}
	buff := make([]T, 0)
	i := 0
	for data := range dataStream.Data {
		buff = append(buff, data)
		i++

		if i%mergeAfter == 0 {
			protoArr := ConvertToProtoArr(buff)
			MakeRequest(protoArr, addr, c.BadKll, conn, &c, startConnection, reconAttempt)
			buff = make([]T, 0)
		}
	}
	if conn != nil {
		conn.Close()

	}
	blackhole = buff
}
func BadCountClient[T shared.Number](mergeAfter int, dataStream stream.Stream[T], addr string, startConnection connectionStarter) {
	var reconAttempt *int = new(int)
	*reconAttempt = 0
	c, conn, err := startConnection(addr)
	if err != nil {
		fmt.Println(err)
		panic("could not start connection")
	}
	buff := make([]T, 0)
	i := 1
	for data := range dataStream.Data {
		buff = append(buff, data)
		i++

		if i%mergeAfter == 0 {
			protoArr := ConvertToProtoArr(buff)
			if b, err := proto.Marshal(protoArr); err == nil {
				fmt.Printf("Message compressed size: %d bytes\n", len(b))
			} else {
				fmt.Printf("Failed to marshal message for stats: %v\n", err)
			}
			MakeRequest(protoArr, addr, c.BadCount, conn, &c, startConnection, reconAttempt)
			buff = make([]T, 0)

		}

	}
	if conn != nil {
		conn.Close()

	}
	return
}

func GetBad[T shared.Number](mergeAfter int, dataStream stream.Stream[T]) *pb.BadArray {
	buff := make([]T, 0)
	i := 1
	for data := range dataStream.Data {
		buff = append(buff, data)
		i++

		if i%mergeAfter == 0 {
			protoArr := ConvertToProtoArr(buff)
			return protoArr
		}
	}
	return nil
}

func ConvertToProtoArr[T shared.Number](arr []T) *pb.BadArray {
	t := fmt.Sprintf("%T", arr)[2:]
	protoRow := pb.NumericRow{}

	for _, val := range arr {
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
	return &pb.BadArray{Arr: &protoRow, Type: t}
}
