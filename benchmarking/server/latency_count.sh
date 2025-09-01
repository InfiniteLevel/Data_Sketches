#!/usr/bin/env bash
fileName="count-Latency-$(date +"%H-%M-%S@%F")"

cd ../../server && go test -timeout 8h -run TestServerLatencyCount | tee "../benchmarking/server/${fileName}"