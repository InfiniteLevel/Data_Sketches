from matplotlib import pyplot as plt
from collections import defaultdict
import re

def parse_go_benchmark(log: str) -> list[dict[str, str | int]]:
    # Map from param signature -> list of (iterations, time)
    grouped_results = defaultdict(list)

    pattern = re.compile(
        r'Benchmark(?P<name>[A-Za-z0-9]+)/(?P<params>(?:[A-Za-z0-9]+:_\d+,?_?)+)-\d+\s+(?P<iter>[0-9]+)\s+(?P<ns>[0-9]+) ns/op\s',
        re.MULTILINE
    )

    for match in pattern.finditer(log):
        benchmark_name = match.group("name")
        params_string = match.group("params")
        iterations = int(match.group("iter"))
        time_ns_per_op = int(match.group("ns"))

        # Turn param string into sorted tuple so grouping is deterministic
        params = {}
        for param in params_string.split(","):
            key, value = param.split(":_")
            params[key] = int(value)

        # Group key based on all relevant param values
        key = (benchmark_name,) + tuple(sorted(params.items()))
        grouped_results[key].append(time_ns_per_op)

    # Build final averaged result list
    parsed_results = []
    for key, times in grouped_results.items():
        benchmark_name = key[0]
        param_pairs = key[1:]
        avg_time = sum(times) // len(times)

        entry = {
            "Benchmark": benchmark_name,
            "Time_ns_per_op": avg_time
        }
        entry.update(dict(param_pairs))
        parsed_results.append(entry)

    return parsed_results

def extract_data(data: str, label: str):
    parsed = parse_go_benchmark(data)
    results = defaultdict(lambda: defaultdict(lambda: defaultdict(list)))

    for row in parsed:
        merge_rate = row.get("MergeRate")
        stream_rate = row.get("StreamRate")
        clients = 3*row.get("Clients")
        time_per_op = row.get("Time_ns_per_op")

        if None not in (merge_rate, stream_rate, clients, time_per_op):
            ops_per_sec = clients / (float(time_per_op) / (14000.0 * 1_000_000_000.0))
            results[stream_rate][merge_rate][label].append((clients, ops_per_sec))

    return results

def merge_results(*results_list):
    merged = defaultdict(lambda: defaultdict(lambda: defaultdict(list)))
    for results in results_list:
        for stream_rate, merge_map in results.items():
            for merge_rate, system_map in merge_map.items():
                for label, data in system_map.items():
                    merged[stream_rate][merge_rate][label].extend(data)
    return merged

def plot_benchmark(results, name):
    all_stream_rates = sorted(results.keys())
    for stream_rate in all_stream_rates:
        plt.figure(figsize=(10, 6))
        plt.title(f"{name} Throughput vs Clients (StreamRate={stream_rate})")
        plt.xlabel("Number of Clients")
        plt.ylabel("Data points per second")
        plt.grid(True, which="both", linestyle="--", linewidth=0.5)

        handles_good = []
        labels_good = []
        handles_bad = []
        labels_bad = []

        for merge_rate, system_map in results[stream_rate].items():
            for label in sorted(system_map.keys()):  # 'bad' > 'good' by default so we sort
                points = sorted(system_map[label])
                x_vals, y_vals = zip(*points)
                style = {
                    "marker": "o" if label == "good" else "x",
                    "linestyle": "-" if label == "good" else "--",
                    "label": f"{label} - Merge Size {merge_rate}",
                }
                line, = plt.plot(x_vals, y_vals, **style)

                if label == "good":
                    handles_good.append(line)
                    labels_good.append(style["label"])
                else:
                    handles_bad.append(line)
                    labels_bad.append(style["label"])

        plt.legend(handles_good + handles_bad, labels_good + labels_bad)
        plt.xscale('log')
        plt.tight_layout()
        plt.show()

def plot_benchmark_grouped_subplots(results, name):
    all_stream_rates = sorted(results.keys())
    num_plots = len(all_stream_rates)
    rows, cols = 4, 1

    fig, axs = plt.subplots(rows, cols, figsize=(8, 10),  sharex=True)
    axs = axs.flatten()

    all_handles = {}
    for i, stream_rate in enumerate(all_stream_rates):
        ax = axs[i]
        ax.set_title(f"StreamRate = {stream_rate}")
        ax.set_xlabel("Number of Clients")
        if i % cols == 0:
            ax.set_ylabel("Data points per second")
        ax.grid(True, which="both", linestyle="--", linewidth=0.5)

        for merge_rate, system_map in results[stream_rate].items():
            for label in sorted(system_map.keys()): 
                points = sorted(system_map[label])
                if not points:
                    continue
                x_vals, y_vals = zip(*points)
                style = {
                    "marker": "o" if label == "good" else "x",
                    "linestyle": "-" if label == "good" else "--",
                    "label": f"{"Federated" if label == "good" else "naive"} - MergeRate {merge_rate}",
                }
                line, = ax.plot(x_vals, y_vals, **style)
                all_handles[style["label"]] = line

        ax.set_xscale("log")

    for j in range(num_plots, rows * cols):
        fig.delaxes(axs[j])

    sorted_handles = sorted(
        all_handles.items(),
        key=lambda x: (0 if "good" in x[0].lower() else 1, x[0])
    )

    fig.legend(
        handles=[h for _, h in sorted_handles],
        labels=[l for l, _ in sorted_handles],
        loc='center left',
        bbox_to_anchor=(0.01, 0.5),
        ncol=1,
        fontsize='large',
        frameon=False
    )


    fig.suptitle(f"{name} Throughput vs Clients", fontsize=16)
    fig.tight_layout(rect=[0.35, 0, 1, 0.95])
    # plt.savefig("systemThroughputPlot")
    plt.show()

def plotFunc():
    good_kll_data = open("../benchmarking/system/goodKll10-24-15@2025-04-16").read()
    bad_kll_data = open("../benchmarking/system/badKll13-38-42@2025-04-16").read()
    good_count_data = open("../benchmarking/system/goodCount11-32-53@2025-04-16").read()
    bad_count_data = open("../benchmarking/system/badCount14-37-32@2025-04-16").read()

    kll_good = extract_data(good_kll_data, "good")
    kll_bad = extract_data(bad_kll_data, "bad")
    count_good = extract_data(good_count_data, "good")
    count_bad = extract_data(bad_count_data, "bad")

    merged_kll = merge_results(kll_good, kll_bad)
    merged_count = merge_results(count_good, count_bad)

    plot_benchmark_grouped_subplots(merged_kll, "KLL")
    plot_benchmark_grouped_subplots(merged_count, "Count")


if __name__ == "__main__":
    plotFunc()
