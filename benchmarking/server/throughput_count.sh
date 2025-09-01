#!/usr/bin/env bash
fileName="count-Throughput-$(date +"%H-%M-%S@%F")"

cd ../../server && go test -timeout 8h -run ^$ -bench BenchmarkThroughputCount -benchtime=30x | tee "../benchmarking/server/${fileName}"

