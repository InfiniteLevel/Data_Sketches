package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/bruhng/distributed-sketching/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func mustDial(addr string, timeout time.Duration) *grpc.ClientConn {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	conn, err := grpc.DialContext(
		ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("dial %s failed: %v", addr, err)
	}
	return conn
}

func writeCSV(filename string, header []string, rows [][]string, appendMode bool) error {
	var f *os.File
	var err error

	if appendMode {
		f, err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	} else {
		f, err = os.Create(filename)
	}
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if !appendMode {
		if err := w.Write(header); err != nil {
			return err
		}
	}

	for _, r := range rows {
		if err := w.Write(r); err != nil {
			return err
		}
	}
	return nil
}

func runTopK(c pb.SketcherClient, typ, field string, k uint32, csvPath string, appendMode bool) error {
	req := &pb.TopKRequest{
		Type:  typ,
		K:     k,
		Field: field,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	resp, err := c.TopKASketch(ctx, req)
	if err != nil {
		return fmt.Errorf("TopKASketch: %w", err)
	}

	fmt.Printf("\n[RESULT] %s | field=%s | top-%d | %s\n", time.Now().Format("15:04:05"), field, k, csvPath)
	rows := make([][]string, len(resp.GetEntries()))
	for i, e := range resp.GetEntries() {
		var key string
		switch v := e.GetKey().GetValue().(type) {
		case *pb.NumericValue_IntVal:
			key = fmt.Sprintf("%d", v.IntVal)
		case *pb.NumericValue_FloatVal:
			key = fmt.Sprintf("%g", v.FloatVal)
		default:
			key = "?"
		}
		fmt.Printf("%-4d %-12s %d\n", i+1, key, e.GetEstFreq())
		rows[i] = []string{
			time.Now().Format(time.RFC3339),
			fmt.Sprintf("%d", i+1),
			field,
			key,
			fmt.Sprintf("%d", e.GetEstFreq()),
		}
	}

	header := []string{"timestamp", "rank", "field", "key", "estimated_frequency"}
	if err := writeCSV(csvPath, header, rows, appendMode); err != nil {
		return fmt.Errorf("write csv: %w", err)
	}
	return nil
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "gRPC server address (host:port)")
	typ := flag.String("type", "int", "value type: int|float64")
	field := flag.String("field", "", "column/element name for which to compute Top-K")
	k := flag.Uint("topk", 10, "K for Top-K")
	csvPath := flag.String("out", "test.csv", "path to output CSV file")
	watch := flag.Duration("watch", 0, "repeat every duration (e.g. 2s, 1m); 0 disables")
	flag.Parse()

	if *field == "" {
		log.Fatal("--field must be specified")
	}

	conn := mustDial(*addr, 5*time.Second)
	defer conn.Close()
	client := pb.NewSketcherClient(conn)

	do := func(first bool) {
		if err := runTopK(client, *typ, *field, uint32(*k), *csvPath, !first); err != nil {
			log.Printf("error: %v", err)
		}
	}

	if *watch > 0 {
		first := true
		t := time.NewTicker(*watch)
		defer t.Stop()
		for {
			do(first)
			first = false
			<-t.C
		}
	} else {
		do(true)
	}
}
