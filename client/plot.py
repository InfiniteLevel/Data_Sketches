import matplotlib.pyplot as plt
import re
import numpy as np

# Initialize lists to store data for each type
data = {
    'kll': {'stream_rates': [], 'processing_rates': [], 'label': 'Kll'},
    'count': {'stream_rates': [], 'processing_rates': [], 'label': 'Count'},
    'center': {'stream_rates': [], 'processing_rates': [], 'label': 'Buffering'},
}

# Regular expression to match lines like "CountThroughput/StreamRate:_X-4 Y[m] ± Z%" or "CountThroughput/StreamRate:_X-4 Y ± Z%"
pattern = r"^(?:CountThroughput|KllThroughput|CenterThroughput)/StreamRate:_(\d+)-(\d+)\s+(\d+\.\d+(?:m|µ)?)\s*±\s*∞\s*¹$"

# File mappings to data types and units
file_mappings = {
    # 'kll.stat': {'type': 'kll', 'unit': 'µs'},
    # 'count.stat': {'type': 'count', 'unit': 'µs'},
    # 'center.stat': {'type': 'center', 'unit': 'µs'},
    'kll-nano.stat': {'type': 'kll', 'unit': 'ns'},
    'count-nano.stat': {'type': 'count', 'unit': 'ns'},
    'center-nano.stat': {'type': 'center', 'unit': 'ns'},
}

# Read and parse data from each file
for filename, info in file_mappings.items():
    try:
        with open(f"./{filename}", "r", encoding='utf-8') as file:
            for line in file:
                match = re.match(pattern, line.strip())
                if match:
                    # Extract stream inter-arrival time
                    inter_arrival_time = int(match.group(1))
                    # Skip if inter_arrival_time is 0 to avoid division by zero
                    if inter_arrival_time == 0:
                        continue

                    # Convert to stream rate (tuples/sec)
                    if info['unit'] == 'ns':
                        # Nanoseconds: 1 / (ns / 10^9) = 10^9 / ns
                        stream_rate_tuples_per_sec = 10**9 / inter_arrival_time
                    else:
                        # Microseconds: 1 / (µs / 10^6) = 10^6 / µs
                        stream_rate_tuples_per_sec = 10**6 / inter_arrival_time

                    # Extract the throughput value
                    value_str = match.group(3)
                    # Check if value is in milliseconds (has 'm') or seconds (no 'm')
                    if value_str.endswith('m'):
                        # Remove 'm' and convert to float (already in ms)
                        value = float(value_str[:-1])
                    elif value_str.endswith('µ'):
                        value = float(value_str[:-1]) / 1000
                    else:
                        # No 'm', value is in seconds, convert to ms by multiplying by 1000
                        value = float(value_str) * 1000

                    # Calculate processing rate (tuples/sec): (7000 / value_ms) * 1000
                    processing_rate_tuples_per_sec = (1000 / value) * 1000
                    if "pinned" in filename:
                        processing_rate_tuples_per_sec = (1400 * 36 / value) * 1000

                    data[info['type']]['stream_rates'].append(stream_rate_tuples_per_sec)
                    data[info['type']]['processing_rates'].append(processing_rate_tuples_per_sec)
    except FileNotFoundError:
        print(f"File {filename} not found, skipping.")

plt.rcParams['font.size'] = 9
# Create the plot
width_pt = 426.79135
width_in = width_pt / 72
height = width_in/1.5
plt.figure(figsize=(width_in, height))

# Plot data from each type with different colors
colors = ['b-', 'r-', 'g-', 'c-']
# markers = ['bo', 'ro', 'go']  # Corresponding markers for each type
for (data_type, type_data), color in zip(data.items(), colors):
    if type_data['stream_rates']:  # Only plot if data exists
        # Sort data by stream rate to ensure a smooth line
        sorted_indices = np.argsort(type_data['stream_rates'])
        sorted_stream_rates = [type_data['stream_rates'][i] for i in sorted_indices]
        sorted_processing_rates = [type_data['processing_rates'][i] for i in sorted_indices]
        plt.plot(sorted_stream_rates, sorted_processing_rates, color, label=f"{type_data['label']} Throughput",marker='o',markersize=2)

# Add 1:1 line
# Get the maximum stream rate and processing rate to define the range
# all_stream_rates = [rate for type_data in data.values() for rate in type_data['stream_rates']]
# all_processing_rates = [rate for type_data in data.values() for rate in type_data['processing_rates']]
# max_rate = max(max(all_stream_rates, default=0), max(all_processing_rates, default=0))
# plt.plot([0, max_rate], [0, max_rate], 'k--', label='Optimal Throughput')

lim = 7000000
# lim = 10000
plt.xlim(0, lim)
plt.ylim(0, lim-3000000)


plt.xlabel('Stream Rate (T/s)')
plt.ylabel('Throughput (T/s)')
plt.title('Client Throughput')
plt.grid(True)
plt.legend()

# Save the plot
plt.savefig("small.png")
plt.show()
plt.show()
