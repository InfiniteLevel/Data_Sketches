import re

def parse_go_benchmark(log: str) -> list[dict[str, str | int]]: 
    parsed_results:list[dict[str, str | int]]  = [] 
    
    pattern = re.compile(
        r'Benchmark([A-Za-z0-9]+)/((?:[A-Za-z0-9]+:_\d+,?_?)+)-\d+\s+([0-9]+)\s+([0-9]+) ns/op\s+([0-9]+) B/op\s+([0-9]+) allocs/op',
        re.MULTILINE
    )

    for match in pattern.finditer(log):
        benchmark_name = match.group(1)
        params_string = match.group(2)
        iterations = int(match.group(3))
        time_ns_per_op = int(match.group(4))
        bytes_per_op = int(match.group(5))
        allocs_per_op = int(match.group(6))

        params = {}
        for param in params_string.split(","):
            key, value = param.split(":_")
            params[key] = int(value)

        parsed_results.append({
            "Benchmark": benchmark_name,
            **params,
            "Iterations": iterations,
            "Time_ns_per_op": time_ns_per_op,
            "Bytes_per_op": bytes_per_op,
            "Allocs_per_op": allocs_per_op,
        })
    return parsed_results

# Example usage
log_data = """
BenchmarkSystemCount/Clients:_5,MergeRate:_100,StreamRate:_200-14                    1        8342760500 ns/op       179960240 B/op    2156157 allocs/op
BenchmarkOtherTest/Test:_100,Rate:_50-14                      1         223676000 ns/op       179237176 B/op    2153734 allocs/op
BenchmarkSystemCount/Param1:_5,Param2:_200,AnotherParam:_300-14                    1        8311962900 ns/op       112735448 B/op    1840107 allocs/op
BenchmarkNewBenchmark/Speed:_10,Threads:_4-14                      1         110819400 ns/op       112804160 B/op    1840557 allocs/op
BenchmarkSystemCount/X:_50,Y:_75,Z:_25-14                   1        8581953600 ns/op       104282752 B/op    1800362 allocs/op
BenchmarkTestX/A:_500,B:_1000,C:_2000-14                     1          96352700 ns/op       104376464 B/op    1800742 allocs/op
BenchmarkSystemCount/Threads:_8,Load:_90-14                   1        8352640300 ns/op       97442824 B/op     1768114 allocs/op
BenchmarkAnotherOne/Factor:_3,Multiplier:_7-14                     1          78123500 ns/op       97481720 B/op     1768331 allocs/op
"""

parsed = parse_go_benchmark(log_data)
for entry in parsed:
    print(entry)

