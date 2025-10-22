package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"time"
)

func quantize(v float64, round int) float64 {
	if round < 0 {
		return v
	}
	f := math.Pow10(round)
	return math.Round(v*f) / f
}

func readTruth(csvPath, header string, round int) (map[float64]int, error) {
	f, err := os.Open(csvPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return map[float64]int{}, nil
	}

	col := -1
	for i, h := range rows[0] {
		if h == header {
			col = i
			break
		}
	}
	if col < 0 {
		return nil, fmt.Errorf("header %q not found", header)
	}

	gt := make(map[float64]int)
	for i := 1; i < len(rows); i++ {
		v, err := strconv.ParseFloat(rows[i][col], 64)
		if err != nil {
			continue
		}
		v = quantize(v, round)
		gt[v]++
	}
	return gt, nil
}

type pair struct {
	key float64
	cnt int
}

func topKFromCounts(m map[float64]int, k int) []pair {
	ps := make([]pair, 0, len(m))
	for k0, c := range m {
		ps = append(ps, pair{k0, c})
	}
	sort.Slice(ps, func(i, j int) bool {
		if ps[i].cnt != ps[j].cnt {
			return ps[i].cnt > ps[j].cnt
		}
		return ps[i].key < ps[j].key
	})
	if k > len(ps) {
		k = len(ps)
	}
	return ps[:k]
}

type est struct {
	ts   string
	rank int
	key  float64
	est  int
}

func readASketchBlock(path string, round int) ([]est, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) <= 1 {
		return nil, nil
	}

	blocks := map[string][]est{}
	order := []string{}
	for i := 1; i < len(rows); i++ {
		if len(rows[i]) < 5 {
			continue
		}
		ts := rows[i][0]
		rk, _ := strconv.Atoi(rows[i][1])
		kf, _ := strconv.ParseFloat(rows[i][3], 64)
		kf = quantize(kf, round)
		ef, _ := strconv.Atoi(rows[i][4])
		if _, ok := blocks[ts]; !ok {
			order = append(order, ts)
		}
		blocks[ts] = append(blocks[ts], est{ts, rk, kf, ef})
	}
	if len(order) == 0 {
		return nil, nil
	}
	last := order[len(order)-1]
	sort.Slice(blocks[last], func(i, j int) bool { return blocks[last][i].rank < blocks[last][j].rank })
	return blocks[last], nil
}

func lastK(as []est, k int) []est {
	if k > len(as) {
		k = len(as)
	}
	return as[len(as)-k:]
}

func main() {
	truthCSV := flag.String("truth", "", "path to source CSV (same as producer)")
	header := flag.String("header", "", "column/header name (e.g. vdop)")
	round := flag.Int("round", -1, "quantize float by N decimals; -1 to disable")
	topk := flag.Int("k", 10, "K")
	asketch := flag.String("asketch", "topk_result.csv", "ASketch CSV exported by auto_query")
	flag.Parse()

	if *truthCSV == "" || *header == "" {
		log.Fatal("usage: verify_topk --truth data.csv --header <name> --k 10 --asketch vdop_topk.csv")
	}

	gt, err := readTruth(*truthCSV, *header, *round)
	if err != nil {
		log.Fatal(err)
	}
	gtTop := topKFromCounts(gt, *topk)

	asBlk, err := readASketchBlock(*asketch, *round)
	if err != nil {
		log.Fatal(err)
	}
	if len(asBlk) == 0 {
		log.Fatal("ASketch CSV has no rows (only header?)")
	}

	asTop := lastK(asBlk, *topk)

	gtSet := map[float64]struct{}{}
	for _, p := range gtTop {
		gtSet[p.key] = struct{}{}
	}

	var hit int
	var relErrSum float64
	for _, e := range asTop {
		if _, ok := gtSet[e.key]; ok {
			hit++
		}
		if c, ok := gt[e.key]; ok && c > 0 {
			relErrSum += math.Abs(float64(e.est-c)) / float64(c)
		}
	}
	prec := float64(hit) / float64(len(asTop))
	rec := float64(hit) / float64(len(gtTop))
	avgRE := 0.0
	if hit > 0 {
		avgRE = relErrSum / float64(hit)
	}

	fmt.Printf("=== Verify Top-%d @ %s ===\n", *topk, time.Now().Format(time.RFC3339))
	fmt.Printf("Precision@%d: %.3f  Recall@%d: %.3f  AvgRelErr(overlap): %.3f\n", *topk, prec, *topk, rec, avgRE)

	fmt.Println("\nGroundTruth (key,count):")
	for i, p := range gtTop {
		fmt.Printf("%2d. %-8g %d\n", i+1, p.key, p.cnt)
	}
	fmt.Println("\nASketch (key,est):")
	for i, e := range asTop {
		fmt.Printf("%2d. %-8g %d\n", i+1, e.key, e.est)
	}
}
