#!/usr/bin/env bash
fileName="$(date +"%H-%M-%S@%F")"

cd ../client && git pull && go test . -timeout 8h -bench BenchmarkKllMemory > "../benchmarking/client/${fileName}"

go tool pprof -web -alloc_space --show_from=KllClient ..\benchmarking\client\memprofile_kll_1.pprof
go tool pprof -web -alloc_space --show_from=KllClient ..\benchmarking\client\memprofile_kll_10.pprof
go tool pprof -web -alloc_space --show_from=KllClient ..\benchmarking\client\memprofile_kll_100.pprof
go tool pprof -web -alloc_space --show_from=KllClient ..\benchmarking\client\memprofile_kll_1000.pprof
go tool pprof -web -alloc_space --show_from=KllClient ..\benchmarking\client\memprofile_kll_10000.pprof
