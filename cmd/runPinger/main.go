package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bruhng/distributed-sketching/client"
	pb "github.com/bruhng/distributed-sketching/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var DATA_SET_PATH = "../data/PVS 1/dataset_gps.csv"
var HEADER_NAME = "speed_meters_per_second"
var PORT = "8080"
var SERVER_ADR = "127.0.0.1"
var NUM_CLIENTS = 10
var NUM_STREAM_RUNS = 10
var STREAM_DELAY = 0

func main() {
	samples := 10000
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
