#!/usr/bin/env bash
fileName="count$(date +"%H-%M-%S@%F")"

cd ../../client && go test . -timeout 8h -bench BenchmarkCountThroughput -benchtime=30x -run=^$ | tee "../benchmarking/client/${fileName}"


