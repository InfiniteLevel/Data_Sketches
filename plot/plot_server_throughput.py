from matplotlib import pyplot as plt
from collections import defaultdict
import re

def parse_go_benchmark(log: str) -> list[dict[str, str | int]]:
    parsed_results: list[dict[str, str | int]] = []

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
    results = defaultdict(lambda: defaultdict(list))  # stream_rate -> merge_rate -> list of (clients, ops/sec)

    for row in parsed:
        merge_rate = row.get("MergeRate")
        stream_rate = row.get("StreamRate")
        clients = row.get("Clients")
        time_per_op = row.get("Time_ns_per_op")

        if None not in (merge_rate, stream_rate, clients, time_per_op):
            ops_per_sec = clients / (float(time_per_op) / (14000.0 * 1_000_000_000.0))
            results[stream_rate][merge_rate].append((clients, ops_per_sec))

    return results

def plot_benchmark(results, name):
    all_stream_rates = sorted(results.keys())
    for stream_rate in all_stream_rates:
        plt.figure(figsize=(10, 6))
        plt.title(f"{name} Throughput vs Clients (StreamRate={stream_rate})")
        plt.xlabel("Number of Clients")
        plt.ylabel("Data points per second")
        plt.grid(True, which="both", linestyle="--", linewidth=0.5)

        merge_map = results[stream_rate]
        for merge_rate, points in merge_map.items():
            points = sorted(points)  
            x_vals, y_vals = zip(*points)
            plt.plot(x_vals, y_vals, marker='x', label=f"MergeRate {merge_rate}")

        plt.xscale('log')
        plt.legend()
        plt.tight_layout()
        plt.show()

import matplotlib.pyplot as plt
from collections import defaultdict

def plot_benchmark_grouped_subplots(results, name, axs):
    all_stream_rates = sorted(results.keys())
    num_plots = len(all_stream_rates)
    rows, cols = 4, 1  # Create a 2x2 grid

    all_handles = {}  # To store handles for the legend
    for i, stream_rate in enumerate(all_stream_rates):
        ax = axs[i % rows]  # Correct indexing for a 2D grid
        ax.set_title(f"StreamRate = {stream_rate}")
        ax.set_xlabel("Number of Clients")
        if i % cols == 0:
            ax.set_ylabel("Data points per second")
        ax.grid(True, which="both", linestyle="--", linewidth=0.5)

        merge_map = results[stream_rate]
        for merge_rate, points in merge_map.items():
            points = sorted(points)  
            x_vals, y_vals = zip(*points)
            label = f"MergeRate {merge_rate}"
            line, = ax.plot(x_vals, y_vals, marker='x', label=label)
            all_handles[label] = line

        ax.set_xscale('log')

    # Sort legend entries
    sorted_handles = sorted(
        all_handles.items(),
        key=lambda x: (int(x[0].split()[-1]))  # Sort by merge rate number
    )

    # Adjust the legend position and place it outside the plot
    axs[0].figure.legend(
        handles=[h for _, h in sorted_handles],
        labels=[l for l, _ in sorted_handles],
        loc='center left',
        bbox_to_anchor=(0.01, 0.5),
        ncol=1,
        fontsize='large',
        frameon=False
    )

    # Set the main title for the entire figure
    axs[0].figure.suptitle(f"{name} Throughput vs Clients", fontsize=16)


def plotFunc():
    # Read in KLL and Count data
    kll_data = open("../benchmarking/server/kll-Throughput-09-54-44@2025-04-15").read()
    count_data = open("../benchmarking/server/count-Throughput-10-40-20@2025-04-15").read()

    # Process and extract results from both files
    kll_results = extract_data(kll_data)
    count_results = extract_data(count_data)

    # Create a single figure for the subplots
    fig, axs = plt.subplots(4, 1, figsize=(8, 10), sharey=True)

    # Plot KLL
    plot_benchmark_grouped_subplots(kll_results, "KLL", axs)  # Pass axs to the function
    
    # Plot Count
    plot_benchmark_grouped_subplots(count_results, "Count", axs)  # Pass axs to the function

    fig.tight_layout(rect=[0.25, 0, 1, 0.95])  # Adjust layout to fit the legend
    plt.savefig("plot")

if __name__ == "__main__":
    plotFunc()

