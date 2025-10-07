package stream

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/bruhng/distributed-sketching/shared"
	"github.com/google/gopacket/pcap"
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
		data := record[columnIndex]
		parsedData, err := parseNumber(data)
		if err != nil {
			fmt.Println(err)
			panic("Data is not int or float")
		}
		parsed, ok := parsedData.(T)
		if ok {
			streamArr = append(streamArr, parsed)
		} else {
			fmt.Println(err)
			panic("Data is not int or float")
		}

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

func NewStreamFromPcap[T shared.Number](pcapPath string, field string, delayNano int, runAmount int, optional_cutoff ...int) *Stream[T] {
	dataStream := NewStream((make[]T,0),delayNano)
	handle, err := pcap.OpenOffline(pcapPath)
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
