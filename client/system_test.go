package client_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/bruhng/distributed-sketching/client"
	pb "github.com/bruhng/distributed-sketching/proto"
	"github.com/bruhng/distributed-sketching/shared"
	"github.com/bruhng/distributed-sketching/stream"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

var SYSTEM_NUM_STREAM_RUNS = 1
var NUM_MERGES = 5

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
	//
	return c, conn, nil
}

func startFakeConnection(_ string) (pb.SketcherClient, *grpc.ClientConn, error) {
	return nil, nil, nil
}

var clientAmounts []int = []int{100}

var mergeRates []int = []int{10000}

func startKll[T shared.Number](wg *sync.WaitGroup, fg *sync.WaitGroup, cond *sync.Cond, streamRate int, mergeRate int, sketch *pb.KLLSketch, merges int) {
	c, conn, err := startRealConnection(SERVER_ADR + ":" + PORT)

	var mergesMadeGroup sync.WaitGroup

	var reconAttempt *int = new(int)
	if err != nil {
		panic("test failed because no connection")
	}
	wg.Done()
	cond.L.Lock()
	fg.Add(1)

	cond.Wait()
	cond.L.Unlock()

	for range merges {

		time.Sleep(time.Duration(streamRate*mergeRate) * time.Nanosecond)
		mergesMadeGroup.Add(1)
		go func() {
			client.MakeRequest[pb.KLLSketch](sketch, SERVER_ADR, c.MergeKll, conn, &c, startRealConnection, reconAttempt)
			mergesMadeGroup.Done()
		}()
	}
	mergesMadeGroup.Wait()
	fg.Done()
	return
}

func BenchmarkSystemKll(b *testing.B) {

	var wg sync.WaitGroup
	var fg sync.WaitGroup
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	for _, clientAmmount := range clientAmounts {
		for _, mergeRate := range mergeRates {
			for streamRate := 100000; streamRate >= 500; streamRate = int(float64(streamRate) * 0.9) {

				dataStream := *stream.NewStreamFromCsv[float64](DATA_SET_PATH, HEADER_NAME, streamRate, -1)
				sketch := client.GetKll(100, mergeRate, dataStream)

				b.Run(fmt.Sprintf("Clients: %d,MergeRate: %d,StreamRate: %d, DataPoints: %d", clientAmmount, mergeRate, streamRate, NUM_MERGES*mergeRate), func(pb *testing.B) {
					b.ReportAllocs()

					pb.StopTimer()
					for range clientAmmount {
						wg.Add(1)

						go startKll[float64](&wg, &fg, cond, streamRate, mergeRate, sketch, NUM_MERGES)
					}
					wg.Wait()
					pb.StartTimer()

					cond.Broadcast()

					fg.Wait()

				})
				client.RestartServer("10.42.0.1", "8080", 3)
				time.Sleep(10 * time.Second)
			}
		}
	}
}
func startBadKll[T shared.Number](wg *sync.WaitGroup, fg *sync.WaitGroup, cond *sync.Cond, streamRate int, mergeRate int, sketch *pb.BadArray, merges int) {
	c, conn, err := startRealConnection(SERVER_ADR + ":" + PORT)
	var reconAttempt *int = new(int)
	var mergesMadeGroup sync.WaitGroup
	if err != nil {
		panic("test failed because no connection")
	}
	wg.Done()
	cond.L.Lock()
	fg.Add(1)

	cond.Wait()
	cond.L.Unlock()
	for range merges {

		time.Sleep(time.Duration(streamRate*mergeRate) * time.Nanosecond)
		mergesMadeGroup.Add(1)
		go func() {
			client.MakeRequest[pb.BadArray](sketch, SERVER_ADR, c.BadKll, conn, &c, startRealConnection, reconAttempt)
			mergesMadeGroup.Done()
		}()
	}
	mergesMadeGroup.Wait()
	fg.Done()
	return
}

func BenchmarkSystemBadKll(b *testing.B) {

	var wg sync.WaitGroup
	var fg sync.WaitGroup
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	for _, clientAmmount := range clientAmounts {
		for _, mergeRate := range mergeRates {
			for streamRate := 100000; streamRate >= 500; streamRate = int(float64(streamRate) * 0.9) {
				dataStream := *stream.NewStreamFromCsv[float64](DATA_SET_PATH, HEADER_NAME, streamRate, -1)
				sketch := client.GetBad(mergeRate, dataStream)

				b.Run(fmt.Sprintf("Clients: %d,MergeRate: %d,StreamRate: %d, DataPoints: %d", clientAmmount, mergeRate, streamRate, NUM_MERGES*mergeRate), func(pb *testing.B) {
					b.ReportAllocs()

					pb.StopTimer()
					for range clientAmmount {
						wg.Add(1)
						go startBadKll[float64](&wg, &fg, cond, streamRate, mergeRate, sketch, NUM_MERGES)
					}
					wg.Wait()
					pb.StartTimer()

					cond.Broadcast()

					fg.Wait()

				})
				client.RestartServer("10.42.0.1", "8080", 3)
				time.Sleep(10 * time.Second)
			}
		}
	}
}

func startCount[T shared.Number](wg *sync.WaitGroup, fg *sync.WaitGroup, cond *sync.Cond, streamRate int, mergeRate int, sketch *pb.CountSketch, merges int) {
	c, conn, err := startRealConnection(SERVER_ADR + ":" + PORT)
	var reconAttempt *int = new(int)
	var mergesMadeGroup sync.WaitGroup
	if err != nil {
		panic("test failed because no connection")
	}
	wg.Done()
	cond.L.Lock()
	fg.Add(1)

	cond.Wait()
	cond.L.Unlock()
	for range merges {

		time.Sleep(time.Duration(streamRate*mergeRate) * time.Nanosecond)
		mergesMadeGroup.Add(1)
		go func() {
			client.MakeRequest[pb.CountSketch](sketch, SERVER_ADR, c.MergeCount, conn, &c, startRealConnection, reconAttempt)
			mergesMadeGroup.Done()
		}()
	}
	mergesMadeGroup.Wait()

	fg.Done()
	return
}

func BenchmarkSystemCount(b *testing.B) {
	var wg sync.WaitGroup
	var fg sync.WaitGroup
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	for _, clientAmmount := range clientAmounts {
		for _, mergeRate := range mergeRates {
			for streamRate := 100000; streamRate >= 500; streamRate = int(float64(streamRate) * 0.9) {
				dataStream := *stream.NewStreamFromCsv[float64](DATA_SET_PATH, HEADER_NAME, streamRate, -1)
				sketch := client.GetCount(mergeRate, dataStream)
				b.Run(fmt.Sprintf("Clients: %d,MergeRate: %d,StreamRate: %d, DataPoints: %d", clientAmmount, mergeRate, streamRate, NUM_MERGES*mergeRate), func(pb *testing.B) {
					b.ReportAllocs()

					pb.StopTimer()
					for range clientAmmount {
						wg.Add(1)
						go startCount[float64](&wg, &fg, cond, streamRate, mergeRate, sketch, NUM_MERGES)
					}
					wg.Wait()
					pb.StartTimer()

					cond.Broadcast()

					fg.Wait()

				})
				// add reset
				client.RestartServer("10.42.0.1", "8080", 3)
				time.Sleep(10 * time.Second)
			}
		}
	}
}
func startBadCount[T shared.Number](wg *sync.WaitGroup, fg *sync.WaitGroup, cond *sync.Cond, streamRate int, mergeRate int, sketch *pb.BadArray, merges int) {
	c, conn, err := startRealConnection(SERVER_ADR + ":" + PORT)
	var reconAttempt *int = new(int)
	var mergesMadeGroup sync.WaitGroup
	if err != nil {
		panic("test failed because no connection")
	}
	wg.Done()
	cond.L.Lock()
	fg.Add(1)

	cond.Wait()
	cond.L.Unlock()
	for range merges {

		time.Sleep(time.Duration(streamRate*mergeRate) * time.Nanosecond)
		mergesMadeGroup.Add(1)
		go func() {
			client.MakeRequest[pb.BadArray](sketch, SERVER_ADR, c.BadCount, conn, &c, startRealConnection, reconAttempt)
			mergesMadeGroup.Done()
		}()
	}
	mergesMadeGroup.Wait()
	fg.Done()
	return
}

func BenchmarkSystemBadCount(b *testing.B) {

	var wg sync.WaitGroup
	var fg sync.WaitGroup
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	for _, clientAmmount := range clientAmounts {
		for _, mergeRate := range mergeRates {
			for streamRate := 100000; streamRate >= 500; streamRate = int(float64(streamRate) * 0.90) {
				dataStream := *stream.NewStreamFromCsv[float64](DATA_SET_PATH, HEADER_NAME, streamRate, -1)
				sketch := client.GetBad(mergeRate, dataStream)

				b.Run(fmt.Sprintf("Clients: %d,MergeRate: %d,StreamRate: %d, DataPoints: %d", clientAmmount, mergeRate, streamRate, NUM_MERGES*mergeRate), func(pb *testing.B) {
					b.ReportAllocs()

					pb.StopTimer()
					for range clientAmmount {
						wg.Add(1)
						go startBadCount[float64](&wg, &fg, cond, streamRate, mergeRate, sketch, NUM_MERGES)
					}
					wg.Wait()
					pb.StartTimer()

					cond.Broadcast()

					fg.Wait()

				})
				client.RestartServer("10.42.0.1", "8080", 3)
				time.Sleep(10 * time.Second)
			}
		}
	}
}
