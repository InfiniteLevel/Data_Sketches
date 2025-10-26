package consumer

import (
	"bufio"
	"context"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"
	"time"

	pb "github.com/bruhng/distributed-sketching/proto"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Init(port string, adr string) {
	conn, err := grpc.NewClient(adr+":"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println(err)
		panic("Could not connect to server")
	}
	defer conn.Close()
	c := pb.NewSketcherClient(conn)

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Write help for help")
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Could not read string. Please try again")
			continue
		}
		input = strings.TrimSpace(input)
		words := strings.Split(input, " ")

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		switch words[0] {
		case "TestLatency":
			if len(words) < 2 {
				fmt.Println("Please provide an amount of tests")
				continue
			}
			amount, err := strconv.Atoi(words[1])
			if err != nil {
				fmt.Println(err)
				fmt.Println("the amount has to be a int")
				continue
			}
			var avrageTime time.Duration
			for range amount {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				start := time.Now()
				_, err := c.TestLatency(ctx, &pb.EmptyMessage{})
				duration := time.Since(start)
				if err != nil {
					fmt.Println("Could not fetch: ", err)
					continue
				}
				avrageTime = (avrageTime + duration) / 2
			}
			if avrageTime == 0 {
				fmt.Println("No tests could be performed")
				continue
			}
			fmt.Println("Avrage response time: ", avrageTime)
		case "QueryKll":
			if len(words) < 2 {
				fmt.Println("QueryKll requires an int or float")
				continue
			}
			x, err := strconv.Atoi(words[1])
			if err != nil {
				x, err := strconv.ParseFloat(words[1], 64)
				if err != nil {
					fmt.Println("QueryKll requires an int or float")
					continue
				}
				res, err := c.QueryKll(ctx, &pb.NumericValue{Value: &pb.NumericValue_FloatVal{FloatVal: float64(x)}, Type: "float64"})
				if err != nil {
					fmt.Println("Could not fetch: ", err)
					continue
				}
				fmt.Println(res)
				fmt.Println("Quantile", float64(res.Phi)/float64(res.N))
			} else {
				res, err := c.QueryKll(ctx, &pb.NumericValue{Value: &pb.NumericValue_IntVal{IntVal: int64(x)}, Type: "int"})
				if err != nil {
					fmt.Println("Could not fetch: ", err)
					continue
				}
				fmt.Println(res)
			}
		case "ReverseQueryKll":
			if len(words) < 3 {
				fmt.Println("ReverseQueryKll requires a float and a type")
				continue
			}
			x, err := strconv.ParseFloat(words[1], 32)
			if err != nil {
				fmt.Printf("%s is not a float", words[1])
				continue
			}
			var res *pb.NumericValue
			if words[2] == "float" {
				res, err = c.ReverseQueryKll(ctx, &pb.ReverseQuery{Phi: float64(x), Type: "float64"})
				if err != nil {
					fmt.Println("Could not fetch: ", err)
					continue
				}
			} else if words[2] == "int" {
				res, err = c.ReverseQueryKll(ctx, &pb.ReverseQuery{Phi: float64(x), Type: "int"})
				if err != nil {
					fmt.Println("Could not fetch: ", err)
					continue
				}
			} else {
				fmt.Printf("%s is not a valid type", words[2])
			}

			if err != nil {
				fmt.Println("Could not fetch: ", err)
				continue
			}
			fmt.Println(res)

		case "PlotKll":
			if len(words) < 3 {
				fmt.Println("PlotKll requires an int and a type")
				continue
			}

			numBins, err := strconv.Atoi(words[1])
			if err != nil {
				fmt.Printf("%s is not an int", words[1])
				continue
			}
			var res *pb.PlotKllReply

			if "float" == words[2] {
				res, err = c.PlotKll(ctx, &pb.PlotRequest{NumBins: int64(numBins), Type: "float64"})
				if err != nil {
					fmt.Println("Could not fetch: ", err)
					continue
				}
			} else if "int" == words[2] {
				res, err = c.PlotKll(ctx, &pb.PlotRequest{NumBins: int64(numBins), Type: "int"})
				if err != nil {
					fmt.Println("Could not fetch: ", err)
					continue
				}
			} else {
				fmt.Printf("%s is not a valid type", words[2])
				continue
			}
			pmf := res.Pmf
			pHist := plot.New()
			pHist.Title.Text = "KLL Sketch Histogram"
			pHist.X.Label.Text = "Speed (m/s)"
			pHist.Y.Label.Text = "Frequency"

			bars := make(plotter.Values, numBins)
			labels := make([]string, len(pmf))
			for i, v := range pmf {
				bars[i] = float64(v)
				labels[i] = strconv.FormatFloat(res.Step*float64(i), 'f', 1, 64)
			}

			hist, err := plotter.NewBarChart(bars, vg.Points(float64(res.Step)))
			if err != nil {
				fmt.Println("Something went wrong when creating the chart")
				fmt.Println(err)
			}

			hist.Width = vg.Points(res.Step * 29.0)
			hist.LineStyle.Width = vg.Points(2)
			hist.LineStyle.Color = color.RGBA{R: 0, B: 0, G: 0, A: 255}
			hist.Color = color.RGBA{R: 135, G: 206, B: 250, A: 255}

			pHist.Add(hist)
			pHist.NominalX(labels...)
			if err := pHist.Save(12*vg.Inch, 6*vg.Inch, "histogram.png"); err != nil {
				fmt.Println("Something went wrong when saving the chart")
				fmt.Println(err)
			}

			fmt.Println("Histogram saved as histogram.png")

		case "QueryASketch":
			if len(words) < 2 {
				fmt.Println("QueryASketch requires an int or float")
				continue
			}
			x, err := strconv.Atoi(words[1])
			if err != nil {
				x, err := strconv.ParseFloat(words[1], 64)
				if err != nil {
					fmt.Println("QueryASketch requires an int or float")
					continue
				}
				res, err := c.QueryASketch(ctx, &pb.NumericValue{Value: &pb.NumericValue_FloatVal{FloatVal: float64(x)}, Type: "float64"})
				if err != nil {
					fmt.Println("Could not fetch: ", err)
					continue
				}
				fmt.Printf("Frequency of %.2f: %d\n", x, res.Res)
			} else {
				res, err := c.QueryASketch(ctx, &pb.NumericValue{Value: &pb.NumericValue_IntVal{IntVal: int64(x)}, Type: "int"})
				if err != nil {
					fmt.Println("Could not fetch: ", err)
					continue
				}
				fmt.Printf("Frequency of %d: %d\n", x, res.Res)
			}
		case "TopKASketch":
			if len(words) < 3 {
				fmt.Println("TopKASketch requires an int and a type")
				continue
			}
			k, err := strconv.Atoi(words[1])
			if err != nil {
				fmt.Printf("%s is not an int", words[1])
				continue
			}
			var res *pb.TopKReply
			if words[2] == "float" {
				res, err = c.TopKASketch(ctx, &pb.TopKRequest{K: uint32(k), Type: "float64"})
				if err != nil {
					fmt.Println("Could not fetch: ", err)
					continue
				}
			} else if words[2] == "int" {
				res, err = c.TopKASketch(ctx, &pb.TopKRequest{K: uint32(k), Type: "int"})
				if err != nil {
					fmt.Println("Could not fetch: ", err)
					continue
				}
			} else {
				fmt.Printf("%s is not a valid type", words[2])
				continue
			}
			fmt.Println("Top", k, "entries in ASketch:")
			for _, entry := range res.Entries {
				switch v := entry.Key.GetValue().(type) {
				case *pb.NumericValue_IntVal:
					fmt.Printf("Value: %d, Estimated Frequency: %d\n", v.IntVal, entry.EstFreq)
				case *pb.NumericValue_FloatVal:
					fmt.Printf("Value: %.2f, Estimated Frequency: %d\n", v.FloatVal, entry.EstFreq)
				}
			}
		case "help":
			fmt.Println("The valid types are [int, float]\n")

			fmt.Println("TestLatency [int]")
			fmt.Println("Tests the latency for the avrage of [int] requests\n")

			fmt.Println("ReverseQueryKll [float] [string]")
			fmt.Println("Returns value of type [string] at quantile [float]\n")

			fmt.Println("QueryKll x")
			fmt.Println("Returns quantlie of value [int/float]\n")

			fmt.Println("QueryASketch x")
			fmt.Println("Returns frequency count of value [int/float] from ASketch\n")

			fmt.Println("PlotKll [int] [string]")
			fmt.Println("Returns a histogram with [int] buckets of sketch of type [string]\n")

			fmt.Println("help")
			fmt.Println("Prints Help")

		default:
			fmt.Printf("%s is not a command\n", words[0])
			continue
		}

	}

}
