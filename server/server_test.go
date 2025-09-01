package server_test

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	pb "github.com/bruhng/distributed-sketching/proto"
	"github.com/bruhng/distributed-sketching/server"
	"github.com/bruhng/distributed-sketching/stream"

	"github.com/bruhng/distributed-sketching/client"
	"github.com/bruhng/distributed-sketching/sketches/count"
	"github.com/bruhng/distributed-sketching/sketches/kll"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

var DATA_SET_PATH = "../data/PVS 1/dataset_gps.csv"
var HEADER_NAME = "speed_meters_per_second"
var PORT = "8080"
var SERVER_ADR = "127.0.0.1"
var NUM_CLIENTS = 10
var NUM_STREAM_RUNS = 10
var STREAM_DELAY = 0

const bufSize = 1024 * 1024 * 100

var lis *bufconn.Listener

var clientAmounts []int = []int{512, 256, 128, 64, 32}

var streamRates []int = []int{0, 10, 100, 1000}

var samples int = 1000

func BenchmarkThroughputKll(b *testing.B) {
	go server.Init(PORT)
	time.Sleep(500 * time.Millisecond)
	var wg sync.WaitGroup
	var fg sync.WaitGroup
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	for clientAmount := 1; clientAmount <= 10000; clientAmount *= 10 {
		for mergeRate := 1; mergeRate <= 10000; mergeRate *= 10 {
			for _, streamRate := range streamRates {
				b.Run(fmt.Sprintf("Clients: %d,MergeRate: %d,StreamRate: %d", clientAmount, mergeRate, streamRate), func(pb *testing.B) {
					pb.StopTimer()
					for range clientAmount {
						wg.Add(1)
						fg.Add(1)
						go func() {
							wg.Done()
							cond.L.Lock()

							cond.Wait()
							cond.L.Unlock()

							client.Init[float64](PORT, SERVER_ADR, "kll", DATA_SET_PATH, HEADER_NAME, 10, streamRate, mergeRate)
							fg.Done()
						}()
					}
					wg.Wait()
					pb.ResetTimer()
					pb.StartTimer()
					cond.Broadcast()
					fg.Wait()
					pb.StopTimer()
					client.RestartServer(SERVER_ADR, PORT, 1)
					time.Sleep(time.Second)
				})
			}
		}
	}
}

func BenchmarkThroughputCount(b *testing.B) {
	go server.Init(PORT)
	time.Sleep(500 * time.Millisecond)
	var wg sync.WaitGroup
	var fg sync.WaitGroup
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	for clientAmount := 1; clientAmount <= 10000; clientAmount *= 10 {
		for mergeRate := 1; mergeRate <= 10000; mergeRate *= 10 {
			for _, streamRate := range streamRates {
				b.Run(fmt.Sprintf("Clients: %d,MergeRate: %d,StreamRate: %d", clientAmount, mergeRate, streamRate), func(pb *testing.B) {
					for range clientAmount {
						wg.Add(1)
						fg.Add(1)
						go func() {
							wg.Done()
							cond.L.Lock()

							cond.Wait()
							cond.L.Unlock()

							client.Init[float64](PORT, SERVER_ADR, "count", DATA_SET_PATH, HEADER_NAME, 10, streamRate, mergeRate)
							fg.Done()
						}()
					}
					wg.Wait()
					pb.ResetTimer()
					pb.StartTimer()
					cond.Broadcast()
					fg.Wait()
					pb.StopTimer()
					client.RestartServer(SERVER_ADR, PORT, 1)
					time.Sleep(time.Second)
				})
			}
		}
	}
}

func TestServerLatencyKll(t *testing.T) {
	var reconAttempt *int = new(int)
	*reconAttempt = 0
	go server.Init(PORT)
	time.Sleep(500 * time.Millisecond)
	c, conn, err := startRealConnection(SERVER_ADR + ":" + PORT)
	if err != nil {
		panic("Could not start benchmark")
	}

	for mergeRate := 1000; mergeRate <= 1000000; mergeRate *= 2 {
		dataStream := *stream.NewStreamFromCsv[float64](DATA_SET_PATH, HEADER_NAME, 0, 10*mergeRate)
		sketch := client.GetKll(200, mergeRate, dataStream)

		fmt.Println("Mergerate:", mergeRate)
		for range 1000 {
			client.MakeRequest(sketch, SERVER_ADR+":"+PORT, c.MergeKll, conn, &c, startRealConnection, reconAttempt)
		}

		client.RestartServer(SERVER_ADR, PORT, 1)
		time.Sleep(1 * time.Second)
	}
}
func TestServerLatencyCount(t *testing.T) {
	var reconAttempt *int = new(int)
	*reconAttempt = 0
	go server.Init(PORT)
	time.Sleep(500 * time.Millisecond)
	c, conn, err := startRealConnection(SERVER_ADR + ":" + PORT)
	if err != nil {
		panic("Could not start benchmark")
	}

	for mergeRate := 1000; mergeRate <= 1000000; mergeRate *= 2 {
		dataStream := *stream.NewStreamFromCsv[float64](DATA_SET_PATH, HEADER_NAME, 0, 10*mergeRate)
		sketch := client.GetCount(mergeRate, dataStream)

		fmt.Println("Mergerate:", mergeRate)
		for range 1000 {
			client.MakeRequest(sketch, SERVER_ADR+":"+PORT, c.MergeCount, conn, &c, startRealConnection, reconAttempt)
		}

		client.RestartServer(SERVER_ADR, PORT, 1)
		time.Sleep(1 * time.Second)
	}
}
func TestServerLatencyCenteralizedKll(t *testing.T) {
	var reconAttempt *int = new(int)
	*reconAttempt = 0
	go server.Init(PORT)
	time.Sleep(500 * time.Millisecond)
	c, conn, err := startRealConnection(SERVER_ADR + ":" + PORT)
	if err != nil {
		panic("Could not start benchmark")
	}

	for mergeRate := 1000; mergeRate <= 1000000; mergeRate *= 2 {
		dataStream := *stream.NewStreamFromCsv[float64](DATA_SET_PATH, HEADER_NAME, 0, 10*mergeRate)
		sketch := client.GetBad(mergeRate, dataStream)

		fmt.Println("Mergerate:", mergeRate)
		for range 1000 {
			client.MakeRequest(sketch, SERVER_ADR+":"+PORT, c.BadKll, conn, &c, startRealConnection, reconAttempt)
		}

		client.RestartServer(SERVER_ADR, PORT, 1)
		time.Sleep(1 * time.Second)
	}
}
func TestServerLatencyCenteralizedCount(t *testing.T) {
	var reconAttempt *int = new(int)
	*reconAttempt = 0
	go server.Init(PORT)
	time.Sleep(500 * time.Millisecond)
	c, conn, err := startRealConnection(SERVER_ADR + ":" + PORT)
	if err != nil {
		panic("Could not start benchmark")
	}

	for mergeRate := 1000; mergeRate <= 1000000; mergeRate *= 2 {
		dataStream := *stream.NewStreamFromCsv[float64](DATA_SET_PATH, HEADER_NAME, 0, 10*mergeRate)
		sketch := client.GetBad(mergeRate, dataStream)

		fmt.Println("Mergerate:", mergeRate)
		for range 1000 {
			client.MakeRequest(sketch, SERVER_ADR+":"+PORT, c.BadCount, conn, &c, startRealConnection, reconAttempt)
		}

		client.RestartServer(SERVER_ADR, PORT, 1)
		time.Sleep(1 * time.Second)
	}
}

func BenchmarkPinger(b *testing.B) {
	samples = 1000
	conn, err := grpc.NewClient(SERVER_ADR+":"+PORT, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println(err)
		panic("Could not connect to server")
	}
	defer conn.Close()
	c := pb.NewSketcherClient(conn)
	var latencies = make([]int, samples)
	for i := range samples {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		prev := time.Now()
		_, err := c.TestLatency(ctx, &pb.EmptyMessage{})
		after := time.Now()
		time.Sleep(time.Millisecond)
		if err != nil {
			fmt.Println("Could not fetch: ", err)
			continue
		}
		diff := int(after.Sub(prev).Nanoseconds())
		latencies[i] = diff
	}
	client.RestartServer(SERVER_ADR, PORT, 1)
	fmt.Println(latencies)
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

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterSketcherServer(s, &server.Server{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func BenchmarkKllMergeInt(b *testing.B) {
	b.StopTimer()
	ctx := context.Background()
	conn, err := grpc.NewClient("bufnet", grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024)), grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		b.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	server := pb.NewSketcherClient(conn)
	sketch := kll.NewKLLSketch[int](200)
	server.MergeKll(ctx, client.ConvertToProtoKLL(sketch))

	for range 100 {
		sketch.Add(rand.Intn(20))
	}
	b.StartTimer()

	for range b.N {
		server.MergeKll(ctx, client.ConvertToProtoKLL(sketch))
	}
}

func BenchmarkCountMergeInt(b *testing.B) {
	b.StopTimer()
	ctx := context.Background()
	conn, err := grpc.NewClient("bufnet", grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024)), grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		b.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	server := pb.NewSketcherClient(conn)
	sketch := count.NewCountSketch[int](111, 50, 5)
	server.MergeCount(ctx, client.ConvertToProtoCount(sketch))
	for range 100 {
		sketch.Add(rand.Intn(100))
	}
	b.StartTimer()

	for range b.N {
		server.MergeCount(ctx, client.ConvertToProtoCount(sketch))
	}
}
