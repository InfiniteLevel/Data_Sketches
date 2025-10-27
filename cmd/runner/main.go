package main

import (
	"flag"
	"sync"

	"github.com/bruhng/distributed-sketching/client"
)

func main() {
	clients := flag.Int("clients", 2, "number of concurrent clients")
	addr := flag.String("a", "127.0.0.1", "server addr")
	port := flag.String("port", "8080", "port")
	sketch := flag.String("sketch", "asketch", "asketch|count|kll")
	data := flag.String("d", "./data/test_zipf.csv", "dataset")
	field := flag.String("name", "item_id", "field")
	typ := flag.String("type", "int", "float|int")
	merge := flag.Int("merge", 500, "merge batch size")
	stream := flag.Int("stream", 100000, "send interval ms")
	flag.Parse()

	var wg sync.WaitGroup
	for i := 0; i < *clients; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if *typ == "float" {
				client.Init[float64](*port, *addr, *sketch, *data, *field, -1, *stream, *merge)
			} else {
				client.Init[int](*port, *addr, *sketch, *data, *field, -1, *stream, *merge)
			}
		}()
	}
	wg.Wait()
}
