package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type FlowKey struct {
	SrcIP, DstIP     string
	SrcPort, DstPort string
	Proto            string
}

type FlowStat struct {
	Packets    int
	Bytes      int
	FirstSeen  time.Time
	LastSeen   time.Time
	Timestamps []float64
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run pcap2csv.go input.pcap output.csv")
		os.Exit(1)
	}
	input := os.Args[1]
	output := os.Args[2]

	handle, err := pcap.OpenOffline(input)
	if err != nil {
		log.Fatalf("Error opening pcap: %v", err)
	}
	defer handle.Close()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	var mu sync.Mutex
	flows := make(map[FlowKey]*FlowStat)
	count := 0

	start := time.Now()
	for packet := range packetSource.Packets() {
		count++
		netLayer := packet.NetworkLayer()
		transLayer := packet.TransportLayer()
		if netLayer == nil || transLayer == nil {
			continue
		}

		var key FlowKey
		var proto string
		var srcPort, dstPort string

		switch l := transLayer.(type) {
		case *layers.TCP:
			proto = "TCP"
			srcPort = strconv.Itoa(int(l.SrcPort))
			dstPort = strconv.Itoa(int(l.DstPort))
		case *layers.UDP:
			proto = "UDP"
			srcPort = strconv.Itoa(int(l.SrcPort))
			dstPort = strconv.Itoa(int(l.DstPort))
		default:
			proto = "Other"
			srcPort = "0"
			dstPort = "0"
		}

		srcIP, dstIP := netLayer.NetworkFlow().Endpoints()
		key = FlowKey{
			SrcIP: srcIP.String(), DstIP: dstIP.String(),
			SrcPort: srcPort, DstPort: dstPort, Proto: proto,
		}

		ts := packet.Metadata().Timestamp
		length := len(packet.Data())

		mu.Lock()
		f, ok := flows[key]
		if !ok {
			f = &FlowStat{
				Packets:    0,
				Bytes:      0,
				FirstSeen:  ts,
				LastSeen:   ts,
				Timestamps: []float64{},
			}
			flows[key] = f
		}
		f.Packets++
		f.Bytes += length
		f.LastSeen = ts
		f.Timestamps = append(f.Timestamps, float64(ts.UnixNano())/1e9)
		mu.Unlock()

		if count%100000 == 0 {
			fmt.Printf("\rProcessed %d packets...", count)
		}
	}
	fmt.Printf("Finished reading %d packets in %v\n", count, time.Since(start))

	writeCSV(flows, output)
	fmt.Printf("Output written to %s\n", output)
}

func writeCSV(flows map[FlowKey]*FlowStat, output string) {
	outDir := filepath.Dir(output)
	os.MkdirAll(outDir, os.ModePerm)

	file, err := os.Create(output)
	if err != nil {
		log.Fatalf("Error creating output CSV: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{
		"src_ip", "dst_ip", "src_port", "dst_port", "protocol",
		"packet_count", "byte_count", "flow_duration",
		"avg_interarrival_time", "avg_byte_rate",
	}
	writer.Write(header)

	for key, f := range flows {
		dur := f.LastSeen.Sub(f.FirstSeen).Seconds()
		if dur <= 0 {
			dur = 0.000001
		}
		// Calculate average inter-arrival times
		avgInter := 0.0
		if len(f.Timestamps) > 1 {
			sum := 0.0
			for i := 1; i < len(f.Timestamps); i++ {
				sum += f.Timestamps[i] - f.Timestamps[i-1]
			}
			avgInter = sum / float64(len(f.Timestamps)-1)
		}
		avgRate := float64(f.Bytes) / dur

		row := []string{
			key.SrcIP, key.DstIP, key.SrcPort, key.DstPort, key.Proto,
			strconv.Itoa(f.Packets),
			strconv.Itoa(f.Bytes),
			fmt.Sprintf("%.6f", dur),
			fmt.Sprintf("%.6f", avgInter),
			fmt.Sprintf("%.3f", avgRate),
		}
		writer.Write(row)
	}
}
