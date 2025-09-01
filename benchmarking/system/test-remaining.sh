#!/usr/bin/env bash
# pssh -i -t 1000000000 -H "sketch@10.42.0.2 sketch@10.42.0.3 sketch@10.42.0.4" "source .profile && cd distributed-sketching/client && git pull && go test . -bench BenchmarkSystemKll -benchtime=1x -count 6 -timeout 8h -run=^$" >> "./kll"
# sleep 5
# pssh -i -t 1000000000 -H "sketch@10.42.0.2 sketch@10.42.0.3 sketch@10.42.0.4" "source .profile && cd distributed-sketching/client && git pull && go test . -bench BenchmarkSystemBadKll -benchtime=1x -count 6 -timeout 8h -run=^$" >> "./center-kll"
# sleep 5
# pssh -i -t 1000000000 -H "sketch@10.42.0.2 sketch@10.42.0.3 sketch@10.42.0.4" "source .profile && cd distributed-sketching/client && git pull && go test . -bench BenchmarkSystemBadCount -benchtime=1x -count 6 -timeout 8h -run=^$" >> "./center-count"
# sleep 5
pssh -i -t 1000000000 -H "sketch@10.42.0.2 sketch@10.42.0.3 sketch@10.42.0.4" "source .profile && cd distributed-sketching/client && git pull && go test . -bench BenchmarkSystemCount -benchtime=1x -count 6 -timeout 8h -run=^$" >> "./count"
