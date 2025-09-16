import matplotlib.pyplot as plt
import re
from collections import defaultdict
 
# Initialize data structure to store results
data = {}
clients_values = [1, 10,50, 100, 500]
merge_rates = [1000, 10000]
datasets = ['Federated KLL Sketch', 'Centralized KLL Sketch', 'Federated Count Sketch', 'Centralized Count Sketch']
 
# Initialize nested dictionary for each Clients, MergeRate, and dataset combination
for client in clients_values:
    data[client] = {}
    for merge_rate in merge_rates:
        data[client][merge_rate] = {}
        for dataset in datasets:
            data[client][merge_rate][dataset] = {'stream_rates': [], 'processing_rates': [], 'label': f'{dataset} with Merge Size: {merge_rate}'}
 
# Temporary structure to collect all sec/op values for averaging
raw_data = {dataset: defaultdict(list) for dataset in datasets}
 
# Regular expression to match lines like "SystemKll/Clients:_X,MergeRate:_Y,StreamRate:_Z-4 W[m] ± V%"
pattern = r"^(SystemKll|SystemBadKll|SystemCount|SystemBadCount)/Clients:_(\d+),MergeRate:_(\d+),StreamRate:_(\d+),_DataPoints:_(\d+)-4\s+(\d+\.\d+m?|\d+\.\d+)"
 
# Read and parse data from files, collecting all sec/op values
for dataset, filename in [('Federated KLL Sketch', 'parsed-kll'), ('Centralized KLL Sketch', 'parsed-center-kll'), ('Federated Count Sketch', 'parsed-count'), ('Centralized Count Sketch', 'parsed-center-count')]:
    with open(filename, 'r') as file:
        for line in file:
            match = re.match(pattern, line.strip())
            if match:
                clients = int(match.group(2))
                merge_rate = int(match.group(3))
                stream_rate = int(match.group(4))  # Stream rate parameter from file
                if stream_rate == 0:
                    continue
                stream_rate_tps = 10**9 / stream_rate  # Convert to tuples/sec
                # Extract the throughput value and convert to milliseconds
                value_str = match.group(6)
                if value_str.endswith('m'):
                    value_ms = float(value_str[:-1])  # Already in milliseconds
                else:
                    value_ms = float(value_str) * 1000  # Convert seconds to milliseconds
                raw_data[dataset][(clients, merge_rate, stream_rate)].append(value_ms)
 
# Average the sec/op values and compute processing rates
for dataset in datasets:
    for (clients, merge_rate, stream_rate), values in raw_data[dataset].items():
        # Average the milliseconds values
        avg_value_ms = sum(values) / len(values)
        # Compute processing rate using your formula
        processing_rate = (merge_rate * 5 * clients *3 / avg_value_ms) * 1000
        # Store in the final data structure
        data[clients][merge_rate][dataset]['stream_rates'].append(10**9 / stream_rate)
        data[clients][merge_rate][dataset]['processing_rates'].append(processing_rate)
 
# Sort the stream_rates and processing_rates for each combination
for client in clients_values:
    for merge_rate in merge_rates:
        for dataset in datasets:
            stream_rates = data[client][merge_rate][dataset]['stream_rates']
            processing_rates = data[client][merge_rate][dataset]['processing_rates']
            if stream_rates:  # Only sort if there is data
                sorted_pairs = sorted(zip(stream_rates, processing_rates))
                data[client][merge_rate][dataset]['stream_rates'] = [pair[0] for pair in sorted_pairs]
                data[client][merge_rate][dataset]['processing_rates'] = [pair[1] for pair in sorted_pairs]
 
# Create a plot for each Clients value
colors = ['b-', 'r-', 'g-']  # Solid, dashed, dash-dot for newgoodKll; adjust for BadSystemKll
# Helper function to plot sketch type
 
def plot_sketch_type(sketch_type_prefix, output_prefix, xlims=None):
    sketch_labels = {
        f'Federated {sketch_type_prefix} Sketch': 'Federated',
        f'Centralized {sketch_type_prefix} Sketch': 'Centralized'
    }
 
    small_clients = [c for c in clients_values if c != 500]
    fig, axs = plt.subplots(2, 2, figsize=(16, 12))
    axs = axs.flatten()
 
    colors = ['b', 'r']
    linestyles = ['-', '--']
    handles_labels = []
 
    for idx, client in enumerate(small_clients):
        ax = axs[idx]
        xlim = xlims.get(client, (0, 200000)) if xlims else (0, 200000)
 
        for i, merge_rate in enumerate(merge_rates):
            for j, dataset in enumerate(sketch_labels):
                stream_rates = data[client][merge_rate][dataset]['stream_rates']
                processing_rates = data[client][merge_rate][dataset]['processing_rates']
                label = f"{sketch_labels[dataset]} {sketch_type_prefix} (Merge size: {merge_rate} tuples)"
                line, = ax.plot(
                    stream_rates,
                    processing_rates,
                    linestyle=linestyles[j],
                    marker='o',
                    markersize=4,
                    linewidth=2,
                    color=colors[i],
                    label=label
                )
                if idx == 0:
                    handles_labels.append((line, label))
 
        ax.set_title(f'Clients: {client * 3}', fontsize=16)
        ax.set_xlabel('Stream Rate (tuples/sec)', fontsize=14)
        ax.set_ylabel('Throughput (tuples/sec)', fontsize=14)
        ax.set_xlim(left=xlim[0], right=xlim[1])
        ax.set_ylim(bottom=0, top=xlim[1] * 3 * client)
        ax.grid(True, which="both", ls="--")
        ax.tick_params(labelsize=14)
        ax.margins(x=0)
 
    # Remove unused axes if any
    for i in range(len(small_clients), len(axs)):
        fig.delaxes(axs[i])
 
    if handles_labels:
        handles, labels = zip(*handles_labels)
        fig.legend(handles, labels, loc='lower center', ncol=2, fontsize=14, bbox_to_anchor=(0.5, -0.05))
 
    fig.suptitle(f'{sketch_type_prefix} System Throughput', fontsize=20, y=1.02)
    plt.tight_layout(rect=[0, 0.05, 1, 0.95])
    plt.savefig(f"{output_prefix}-combined-zoomed.png", bbox_inches='tight')
    plt.close()
 
 
 
kll_xlims = {
    1: (0, 175000),
    10: (0, 225000),
    50: (0, 650000),
    100: (0, 450000),
    500: (0, 450000)
}
 
count_xlims = {
    1: (0, 150000),
    10: (0, 125000),
    50: (0, 75000),
    100: (0, 60000),
    500: (0, 150000)
}
 
plot_sketch_type("KLL", "results-kll", xlims=kll_xlims)
plot_sketch_type("Count", "results-count", xlims=count_xlims)
