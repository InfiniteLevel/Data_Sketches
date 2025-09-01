#!/usr/bin/env bash
fileName="kll-Throughput-$(date +"%H-%M-%S@%F")"

cd ../../server && go test -timeout 8h -run ^$ -bench BenchmarkThroughputKll -benchtime=30x | tee "../benchmarking/server/${fileName}"

