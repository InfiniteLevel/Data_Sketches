// auto_query.go — Automated A-Sketch Query (CSV logging)
// Command Instance：
//   go run auto_query.go -addr 127.0.0.1:8080 -value 25.5 -n 1000 -interval 50ms -out asketch_25_5.csv
//
// Start server and client first：
//   go run . -port 8080
//   go run . -client -port 8080 -address 127.0.0.1 -sketch asketch -data ./data/dataset_gps.csv -name accuracy -type float -merge 50 -stream 1000000000
//
// Then run the script

package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "github.com/bruhng/distributed-sketching/proto"
)

func queryASketch(ctx context.Context, cli pb.SketcherClient, req *pb.NumericValue) (int64, error) {
	resp, err := cli.QueryASketch(ctx, req)
	if err != nil {
		return 0, err
	}
	return resp.Res, nil
}

func queryCount(ctx context.Context, cli pb.SketcherClient, req *pb.NumericValue) (int64, error) {
	resp, err := cli.QueryCount(ctx, req)
	if err != nil {
		return 0, err
	}
	return resp.Res, nil
}

func main() {
	var (
		dtype    = flag.String("dtype", "", "override NumericValue.Type (e.g., int|int64|float)")
		mode     = flag.String("mode", "float", "value type: float|int")
		q        = flag.String("q", "4.0", "query value as text")
		addr     = flag.String("addr", "127.0.0.1:8080", "gRPC server address, host:port")
		n        = flag.Int("n", 1000, "number of queries")
		interval = flag.Duration("interval", 0, "interval between queries (e.g. 50ms); 0 = as fast as possible")
		timeout  = flag.Duration("timeout", 5*time.Second, "per-request timeout")
		out      = flag.String("out", "asketch_queries.csv", "output CSV path")
		api      = flag.String("api", "ASketch", "which API to call: ASketch|count")
		warmup   = flag.Int("warmup", 5, "warmup queries (not recorded)")
	)
	flag.Parse()

	chooseType := func() string {
		if *dtype != "" {
			return *dtype
		}
		apiLower := strings.ToLower(*api)
		if apiLower == "asketch" {
			if *mode == "int" {
				return "int"
			}
			return "float64"
		}
		if *mode == "int" {
			return "int"
		}
		return "float64"
	}

	// construct NumericValue（oneof + Type）
	makeReq := func() (*pb.NumericValue, error) {
		t := chooseType()
		if *mode == "int" {
			iv, err := strconv.ParseInt(*q, 10, 64)
			if err != nil {
				return nil, err
			}
			return &pb.NumericValue{
				Value: &pb.NumericValue_IntVal{IntVal: iv},
				Type:  t,
			}, nil
		}
		fv, err := strconv.ParseFloat(*q, 64)
		if err != nil {
			return nil, err
		}
		return &pb.NumericValue{
			Value: &pb.NumericValue_FloatVal{FloatVal: fv},
			Type:  t,
		}, nil
	}
	// connect gRPC
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial %s: %v", *addr, err)
	}
	defer conn.Close()
	cli := pb.NewSketcherClient(conn)

	// select the function
	call := queryASketch
	if *api == "count" {
		call = queryCount
	}

	req, err := makeReq()
	if err != nil {
		log.Fatalf("build request from -q=%q failed: %v", *q, err)
	}

	// warm up (non-count)
	for i := 0; i < *warmup; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		_, _ = call(ctx, cli, req)
		cancel()
		if *interval > 0 {
			time.Sleep(*interval)
		}
	}

	// open CSV
	f, err := os.Create(*out)
	if err != nil {
		log.Fatalf("create csv: %v", err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	// the CSV sheet heading
	if err := w.Write([]string{"timestamp", "query_value", "estimated_frequency", "latency_ms"}); err != nil {
		log.Fatalf("write header: %v", err)
	}

	// formal query
	start := time.Now()
	fail := 0
	for i := 0; i < *n; i++ {
		ts := time.Now().UTC().Format("2006-01-02 15:04:05.000Z")
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		t0 := time.Now()
		freq, err := call(ctx, cli, req)
		lat := time.Since(t0)
		cancel()
		if err != nil {
			fail++
			if fail <= 3 {
				if st, ok := status.FromError(err); ok {
					fmt.Printf("[debug] #%d grpc error: code=%s msg=%q\n", i, st.Code(), st.Message())
				} else {
					fmt.Printf("[debug] #%d error: %v\n", i, err)
				}
			}
			_ = w.Write([]string{ts, *q, "-1", fmt.Sprintf("%.3f", float64(lat.Microseconds())/1000.0)})
		} else {
			_ = w.Write([]string{ts, *q, fmt.Sprintf("%d", freq), fmt.Sprintf("%.3f", float64(lat.Microseconds())/1000.0)})
		}
		if *interval > 0 {
			time.Sleep(*interval)
		}
	}
	elapsed := time.Since(start).Seconds()
	qps := float64(*n) / elapsed
	w.Flush()

	fmt.Printf("Done %d queries (fail=%d) in %.3fs  =>  QPS = %.1f\nCSV -> %s\n",
		*n, fail, elapsed, qps, *out)
}
