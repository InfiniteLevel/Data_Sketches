import re
import pandas as pd
import matplotlib.pyplot as plt

def parse_mergerate_file(file_path):
    data = []
    mergerate_pattern = re.compile(r'(?:float64)?Mergerate:\s+(\d+)$')
    time_pattern = re.compile(r'^(\d+\.?\d*)\s*(ms|µs)$')  # More flexible pattern
    with open(file_path, 'r', encoding='utf-8') as file:
        lines = file.readlines()
        # Add this inside the parse_mergerate_file() function, just after reading lines:
        if 'federated' in file_path:
            print(f"\n--- Preview of {file_path} ---")
            for line in lines[:10]:  # Show first 10 lines
                print(line.strip())
        current_mergerate = None
        for line in lines:
            line = line.strip()
            if not line:
                continue
            mergerate_match = mergerate_pattern.match(line)
            if mergerate_match:
                current_mergerate = int(mergerate_match.group(1))
                continue
            time_match = time_pattern.match(line)
            if time_match and current_mergerate is not None:
                time_value = float(time_match.group(1))
                unit = time_match.group(2)
                time_ms = time_value / 1000 if unit == 'µs' else time_value
                data.append({'Mergerate': current_mergerate, 'Time_ms': time_ms})
    df = pd.DataFrame(data)
    if df.empty:
        print(f"[Warning] No data parsed from {file_path}")
    return df

files = [
    './centeralized-count-server-latency',
    './centeralized-kll-server-latency',
    './federated-count-server-latency',
    './federated-kll-server-latency'
]
styles = {
    'Centralized Count': {'linestyle': '-', 'marker': 'o'},
    'Centralized KLL': {'linestyle': '--', 'marker': 's'},
    'Federated Count': {'linestyle': '-.', 'marker': '^'},
    'Federated KLL': {'linestyle': ':', 'marker': 'x'}
}
name_map = {
    'centeralized-kll-server-latency': 'Centralized KLL',
    'centeralized-count-server-latency': 'Centralized Count',
    'federated-kll-server-latency': 'Federated KLL',
    'federated-count-server-latency': 'Federated Count'
}

dfs = []
for file in files:
    df = parse_mergerate_file(file)
    df['Dataset'] = name_map.get(file.split('/')[-1], file)
    dfs.append(df)

combined_df = pd.concat(dfs, ignore_index=True)

plt.rcParams['font.size'] = 9
# Create the plot
width_pt = 426.79135
width_in = width_pt / 72
height = width_in/1.5

for dataset in combined_df['Dataset'].unique():
    subset = combined_df[combined_df['Dataset'] == dataset]
    if subset.empty:
        continue
    avg_times = subset.groupby('Mergerate')['Time_ms'].mean()
    mergerate = avg_times.index.astype(float)
    normalized_inverse_y = mergerate / avg_times.values * 1000

    style = styles.get(dataset, {})
    plt.plot(mergerate, normalized_inverse_y, label=dataset, **style)


plt.xlabel('Merge Size')
plt.xlim(0,10000)
plt.ylim(0,150000000)
plt.title("Server Throughput")
plt.ylabel('Throughput (tuples/sec)')
plt.ticklabel_format(style='sci', axis='both', scilimits=(0,0))
plt.grid(True)
plt.legend()
plt.tight_layout()
plt.savefig('throughput_plot.png')
