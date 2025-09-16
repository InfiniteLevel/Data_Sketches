package print

import (
	"encoding/csv"
	"fmt"
	"image/color"
	"log"
	"os"
	"strconv"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func plotHist(path string, targetHeader string, buckets int) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	header := records[0]
	var dataIndex int
	for i, field := range header {
		if field == targetHeader {
			dataIndex = i
			break
		}
	}

	var data plotter.Values
	for _, record := range records[1:] {
		data2, err := strconv.ParseFloat(record[dataIndex], 64)
		if err != nil {
			continue
		}
		data = append(data, data2)
	}

	createHistogram(data, buckets, "histogram_from_data.png")
}

func createHistogram(data plotter.Values, bins int, filename string) {
	p := plot.New()
	p.Title.Text = fmt.Sprintf("Histogram with %d buckets", bins)
	p.X.Label.Text = "Speed (m/s)"
	p.Y.Label.Text = "Frequency"

	hist, err := plotter.NewHist(data, bins)
	if err != nil {
		log.Fatal(err)
	}

	hist.LineStyle.Width = vg.Points(2)
	hist.FillColor = color.RGBA{R: 135, G: 206, B: 250, A: 255}
	hist.Color = color.RGBA{R: 0, B: 0, G: 0, A: 255}

	p.Add(hist)

	if err := p.Save(8*vg.Inch, 4*vg.Inch, filename); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Saved", filename)
}
