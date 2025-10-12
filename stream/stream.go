package stream

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bruhng/distributed-sketching/shared"
)

type Stream[T shared.Number] struct {
	Data chan T
}

func preciseSleep(duration time.Duration) {
	start := time.Now()
	for time.Since(start) < duration {
		// Busy-wait
	}
}

func NewStream[T shared.Number](data []T, delayNano int) *Stream[T] {
	ch := make(chan T, 1000)
	go func() {
		for _, item := range data {
			ch <- item
			preciseSleep(time.Duration(delayNano) * time.Nanosecond)
		}
	}()
	return &Stream[T]{Data: ch}
}

func NewStreamFromCsv[T shared.Number](csvPath string, field string, delayNano int, runAmount int, optional_cutoff ...int) *Stream[T] {
	cutoff := -1
	if len(optional_cutoff) > 0 {
		cutoff = optional_cutoff[0]
	}
	dataStream := NewStream(make([]T, 0), delayNano)
	file, err := os.Open(csvPath)
	if err != nil {
		fmt.Println(err)
		panic("Could not read csv")
	}

	defer file.Close()

	reader := csv.NewReader(file)

	header, err := reader.Read()
	if err != nil {
		fmt.Println(err)
		panic("could not reader header")
	}

	columnIndex := -1
	for i, h := range header {
		if h == field {
			columnIndex = i
			break
		}
	}
	if columnIndex == -1 {
		panic("Invalid field name")
	}
	streamArr := make([]T, 0)
	for {
		record, err := reader.Read()
		if err != nil || cutoff == len(streamArr) {
			break
		}
		data := strings.TrimSpace(record[columnIndex])

		var parsed T
		var ok bool

		switch any(*new(T)).(type) {
		case int64:
			// parse int first, or float second
			if iv, err := strconv.ParseInt(data, 10, 64); err == nil {
				parsed, ok = T(iv), true
			} else if fv, err := strconv.ParseFloat(data, 64); err == nil {
				// accept float number（e.g. 4.0、23.000）
				if math.Abs(fv-math.Round(fv)) < 1e-9 {
					parsed, ok = T(int64(math.Round(fv))), true
				} else {
					parsed, ok = T(int64(math.Round(fv))), true
				}
			}

		case float64:
			// float first, or int second
			if fv, err := strconv.ParseFloat(data, 64); err == nil {
				parsed, ok = T(fv), true
			} else if iv, err := strconv.ParseInt(data, 10, 64); err == nil {
				parsed, ok = T(float64(iv)), true
			}
		}

		if !ok {
			continue
		}
		streamArr = append(streamArr, parsed)
	}
	go func() {
		for i := runAmount; i != 0; i-- {
			for _, val := range streamArr {
				dataStream.Data <- val
				preciseSleep(time.Duration(delayNano) * time.Nanosecond)
			}
		}
		close(dataStream.Data)
	}()

	return dataStream
}

func parseNumber(s string) (any, error) {
	if intValue, err := strconv.ParseInt(s, 10, 64); err == nil {
		return intValue, nil
	}

	if floatValue, err := strconv.ParseFloat(s, 64); err == nil {
		return floatValue, nil
	}

	return nil, fmt.Errorf("%s is not a valid number", s)
}
