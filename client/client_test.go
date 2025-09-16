package client_test

import (
	"fmt"
	"testing"
	// "time"

	"github.com/bruhng/distributed-sketching/client"
	// "github.com/bruhng/distributed-sketching/server"
	"github.com/bruhng/distributed-sketching/stream"
)

var PORT = "8080"
var SERVER_ADR = "10.42.0.1"
var DATA_SET_PATH = "../data/PVS 1/dataset_gps.csv"
var HEADER_NAME = "speed_meters_per_second"
var NUM_STREAM_RUNS = 1

func BenchmarkKllThroughput(b *testing.B) {
	for rate := 1000000; rate > 1; rate = int(float64(rate) * 0.99) {
		b.Run(fmt.Sprintf("StreamRate: %d", rate), func(b *testing.B) {
			for b.Loop() {
				b.StopTimer()
				dataStream := *stream.NewStreamFromCsv[float64](DATA_SET_PATH, HEADER_NAME, rate, NUM_STREAM_RUNS, 1000)
				b.StartTimer()
				client.KllClient(200, 100000, dataStream, SERVER_ADR+":"+PORT, startFakeConnection)
				b.StopTimer()
			}
		})
	}
}
func BenchmarkCountThroughput(b *testing.B) {
	// go server.Init(PORT)
	for rate := 1000000; rate > 1; rate = int(float64(rate) * 0.98) {
		b.Run(fmt.Sprintf("StreamRate: %d", rate), func(b *testing.B) {
			for b.Loop() {
				b.StopTimer()
				dataStream := *stream.NewStreamFromCsv[float64](DATA_SET_PATH, HEADER_NAME, rate, NUM_STREAM_RUNS, 1000)
				b.StartTimer()
				client.CountClient(100000, dataStream, SERVER_ADR+":"+PORT, startFakeConnection)
				b.StopTimer()
			}
			// client.RestartServer(SERVER_ADR, PORT, 1)
			// time.Sleep(1 * time.Second)
		})
	}
}
func BenchmarkCenterThroughput(b *testing.B) {
	// go server.Init(PORT)
	for rate := 1000000; rate > 1; rate = int(float64(rate) * 0.99) {
		b.Run(fmt.Sprintf("StreamRate: %d", rate), func(b *testing.B) {
			for b.Loop() {
				b.StopTimer()
				dataStream := *stream.NewStreamFromCsv[float64](DATA_SET_PATH, HEADER_NAME, rate, NUM_STREAM_RUNS, 1000)
				b.StartTimer()
				client.BadCountClient(100000, dataStream, SERVER_ADR+":"+PORT, startFakeConnection)
				b.StopTimer()
			}
			// client.RestartServer(SERVER_ADR, PORT, 1)
			// time.Sleep(1 * time.Second)
		})
	}
}
