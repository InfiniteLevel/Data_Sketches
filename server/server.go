package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	pb "github.com/bruhng/distributed-sketching/proto"
	"github.com/bruhng/distributed-sketching/sketches/count"
	"github.com/bruhng/distributed-sketching/sketches/kll"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedSketcherServer
}

func (s *Server) TestLatency(_ context.Context, in *pb.EmptyMessage) (*pb.EmptyMessage, error) {
	return &pb.EmptyMessage{}, nil
}

var grpcServer *grpc.Server
var listener net.Listener
var savedPort string

var wg sync.WaitGroup
var waiting bool = false
var mu sync.Mutex
var restarts int = 0

func (s *Server) RestartServer(ctx context.Context, in *pb.RestartMessage) (*pb.EmptyMessage, error) {
	mu.Lock()
	if !waiting {
		waiting = true
		wg.Add(int(in.NumMsg))
		mu.Unlock()
		wg.Done()
		wg.Wait()
		waiting = false
		go func() {
			fmt.Println("Restarting...", restarts)
			restarts++
			time.Sleep(1 * time.Second)
			restartServer()
		}()
		return &pb.EmptyMessage{}, nil
	}
	mu.Unlock()
	wg.Done()
	wg.Wait()
	return &pb.EmptyMessage{}, nil

}

func restartServer() {

	if grpcServer != nil {
		grpcServer.Stop()
	}

	if listener != nil {
		listener.Close()
	}

	resetState()
	startServer()

}

func resetState() {
	kllStateMap.Store("int", kll.NewKLLSketch[int](200))
	kllStateMap.Store("float64", kll.NewKLLSketch[float64](200))
	badCountStateMap.Store("int", count.NewCountSketch[int](157, 100, 10))
	badCountStateMap.Store("float64", count.NewCountSketch[float64](157, 100, 10))
	badKllStateMap.Store("int", kll.NewKLLSketch[int](200))
	badKllStateMap.Store("float64", kll.NewKLLSketch[float64](200))
	asketchStateMap = sync.Map{}
}

func PanicRecoveryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	// Use defer + recover to catch panics
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
			err = status.Errorf( // Convert panic into a gRPC error response
				codes.Internal,
				"internal server error: %v",
				r,
			)
		}
	}()

	// Call the actual gRPC method
	return handler(ctx, req)
}

func Init(port string) {
	savedPort = port
	startServer()
	for {
	}

}

func startServer() {

	var err error
	// if listener == nil {
	listener, err = net.Listen("tcp", ":"+savedPort)
	if err != nil {
		panic(fmt.Sprint("listen error: ", err))

	}

	// }

	grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(PanicRecoveryInterceptor),
		grpc.MaxConcurrentStreams(100_000),
	)
	pb.RegisterSketcherServer(grpcServer, &Server{})
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
