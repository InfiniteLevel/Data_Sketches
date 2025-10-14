package client

import (
	"context"
	"fmt"
	"time"

	pb "github.com/bruhng/distributed-sketching/proto"
	"github.com/bruhng/distributed-sketching/shared"
	"github.com/bruhng/distributed-sketching/stream"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

var MAX_RECONN_ATTEMPTS int = 20

func Init[T shared.Number](port string, adr string, sketchType string, dataSetPath string, headerName string, numStreamRuns int, streamDelayms int, mergeAfter int) {
	dataStream := *stream.NewStreamFromCsv[T](dataSetPath, headerName, streamDelayms, numStreamRuns)

	switch sketchType {
	case "kll":
		KllClient(100, mergeAfter, dataStream, adr+":"+port, startRealConnection)
	case "count":
		CountClient(mergeAfter, dataStream, adr+":"+port, startRealConnection)
	case "asketch":
		ASketchClient(mergeAfter, dataStream, adr+":"+port, startRealConnection)
	case "badCount":
		BadCountClient(mergeAfter, dataStream, adr+":"+port, startRealConnection)
	case "badKll":
		BadKllClient(mergeAfter, dataStream, adr+":"+port, startRealConnection)
	case "streamClient":
		StreamClient(dataStream, adr+":"+port, startRealConnection)
	default:
		panic("No sketch provided or invalid sketch")
	}
}

func startRealConnection(adr string) (pb.SketcherClient, *grpc.ClientConn, error) {

	conn, err := grpc.NewClient(adr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	c := pb.NewSketcherClient(conn)
	conn.Connect()
	// wait for the connection to be established
	for {
		state := conn.GetState()
		if state == connectivity.Idle || state == connectivity.Connecting {
			continue
		} else if state == connectivity.Ready {
			break
		} else {
			return c, conn, fmt.Errorf("Could not establish connection to server")
		}
	}
	return c, conn, nil
}

func RestartServer(addr string, port string, numMsg int64) {

	c, conn, err := startRealConnection(addr + ":" + port)
	if err != nil {
		fmt.Println(err)
		panic("could not start connection")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	_, err = c.RestartServer(ctx, &pb.RestartMessage{NumMsg: numMsg})
	if err != nil {
		fmt.Println(err)
		panic("could not turn of server")

	}
	conn.Close()

}

type mergeFunction[T any] func(context.Context, *T, ...grpc.CallOption) (*pb.MergeReply, error)

func MakeRequest[T any](protoSketch *T, addr string, merge mergeFunction[T], conn *grpc.ClientConn, c *pb.SketcherClient, startConnection connectionStarter, attempt *int) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)

	// Log protobuf size before sending

	_, err := merge(ctx, protoSketch)
	cancel()
	if err != nil {
		fmt.Println(err)
		conn.Close()
		if *attempt > MAX_RECONN_ATTEMPTS {
			fmt.Printf("Could not reconnect after %d attempts shutting down\n", MAX_RECONN_ATTEMPTS)
			panic("Could not reestablish connection")
			// TODO: Maybe remove panic
		}
		tc, tconn, err := startConnection(addr)
		if err != nil {
			fmt.Printf("%d faild reconnection attempt, will try again later\n", *attempt)
			(*attempt)++
		}
		c = &tc
		conn = tconn
		return

	}
	*attempt = 0
}
