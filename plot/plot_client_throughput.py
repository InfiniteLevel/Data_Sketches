from matplotlib import pyplot as plt
import re

def parse_go_benchmark(log: str) -> list[dict[str, str | int]]: 
    parsed_results:list[dict[str, str | int]]  = [] 
    
    pattern = re.compile(
        r'Benchmark([A-Za-z0-9]+)/((?:[A-Za-z0-9]+:_\d+,?_?)+)-\d+\s+([0-9]+)\s+([0-9]+) ns/op\s',
        re.MULTILINE
    )

    for match in pattern.finditer(log):
        benchmark_name = match.group(1)
        params_string = match.group(2)
        iterations = int(match.group(3))
        time_ns_per_op = int(match.group(4))

        params = {}
        for param in params_string.split(","):
            key, value = param.split(":_")
            params[key] = int(value)

        parsed_results.append({
            "Benchmark": benchmark_name,
            **params,
            "Iterations": iterations,
            "Time_ns_per_op": time_ns_per_op
        })
    return parsed_results

def extract_data(data: str):
    parsed = parse_go_benchmark(data)
    x_vals = []
    y_vals = []

    for row in parsed:
        merge_rate = row.get("MergeRate")
        time_per_op = row.get("Time_ns_per_op")

        if merge_rate is not None and time_per_op is not None:
            ops_per_sec = 1.0 / (float(time_per_op) / (1400000.0 * 1_000_000_000.0))
            x_vals.append(merge_rate)
            y_vals.append(ops_per_sec)

    return x_vals, y_vals

def plotFunc():
    kll = open("../benchmarking/client/kll14-47-02@2025-04-14").read()
    count = open("../benchmarking/client/count15-37-11@2025-04-14").read()

    x_kll, y_kll = extract_data(kll)
    x_count, y_count = extract_data(count)

    plt.figure(figsize=(10, 6))
    plt.plot(x_kll, y_kll, marker='x', linestyle='-', linewidth=1, label="kll")
    plt.plot(x_count, y_count, marker='o', linestyle='--', linewidth=2, label="count")

    plt.xscale('log')
    plt.xlim(0.1, 10000000)
    
    # Increase font size for title, labels, and ticks
    plt.xlabel("Merge Size (log scale)", fontsize=14)
    plt.ylabel("Data points per second", fontsize=14)
    plt.title("Normalized Benchmark Time vs Merge Size", fontsize=16)
    
    # Larger ticks on the axes
    plt.xticks(fontsize=12)
    plt.yticks(fontsize=12)
    
    # Increase font size for the legend
    plt.legend(fontsize=12)
    
    plt.grid(True, which="both", linestyle="--", linewidth=0.5)
    plt.tight_layout()
    plt.show()

if __name__ == "__main__":
    plotFunc()
