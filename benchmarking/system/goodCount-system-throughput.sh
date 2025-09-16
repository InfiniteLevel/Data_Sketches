#!/usr/bin/env bash

fileName="count.old"

pssh -i -A -t 1000000000 -H "sketch@10.42.0.2 sketch@10.42.0.3 sketch@10.42.0.4" "source .profile && cd distributed-sketching/client && git pull && go test . -bench BenchmarkSystemCount -benchtime=1x -count 3 -timeout 8h -run=^$" >> "./${fileName}"
