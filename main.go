package main

import (
	"flag"

	"github.com/bruhng/distributed-sketching/client"
	"github.com/bruhng/distributed-sketching/consumer"
	"github.com/bruhng/distributed-sketching/server"
)

func main() {
	isClient := flag.Bool("client", false, "use flag if you want to create a client process instead of server")
	isConsumer := flag.Bool("consumer", false, "use flag if you want to create a consumer process instead of a server")
	port := flag.String("port", "8080", "Choose what port to use")
	address := flag.String("a", "127.0.0.1", "Choose what ip to connect to")
	sketchType := flag.String("sketch", "kll", "Choose what sketch to use")
	dataSetPath := flag.String("d", "./data/PVS 1/dataset_gps.csv", "Choose what data set path to use as data stream")
	dataSetName := flag.String("name", "speed_meters_per_second", "Choose what part of the data set to use as data stream")
	dataSetType := flag.String("type", "float", "Choose what type the data set is")
	mergeRate := flag.Int("merge", 1000, "merge rate for clients")
	streamRate := flag.Int("stream", 10, "stream rate for clients")

	flag.Parse()
	if *isClient {
		if *dataSetType == "float" {
			client.Init[float64](*port, *address, *sketchType, *dataSetPath, *dataSetName, 1, *streamRate, *mergeRate)
		} else if *dataSetType == "int" {
			client.Init[int](*port, *address, *sketchType, *dataSetPath, *dataSetName, 1, *streamRate, *mergeRate)
		}
	} else if *isConsumer {
		consumer.Init(*port, *address)
	} else {
		server.Init(*port)
	}
}
