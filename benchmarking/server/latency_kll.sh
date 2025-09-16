#!/usr/bin/env bash
fileName="kll-Latency-$(date +"%H-%M-%S@%F")"

cd ../../server && go test -timeout 8h -run TestServerLatencyKll | tee "../benchmarking/server/${fileName}"

