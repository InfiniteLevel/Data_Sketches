#!/usr/bin/env bash
fileName="kll$(date +"%H-%M-%S@%F")"

cd ../../client && go test . -timeout 8h -bench BenchmarkKllThroughput -count=10 -run=^$ | tee "../benchmarking/client/${fileName}" 


