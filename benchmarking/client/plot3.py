import matplotlib.pyplot as plt
import re
import numpy as np

# Initialize lists to store data for each file
data = {
    'pi-count-parsed': {'stream_rates': [], 'processing_rates': [], 'label': 'Count'},
    'pi-count-parsed-pinned': {'stream_rates': [], 'processing_rates': [], 'label': 'Count2'},
    'pi-kll-parsed': {'stream_rates': [], 'processing_rates': [], 'label': 'Kll'},
    'pi-kll-parsed-pinned': {'stream_rates': [], 'processing_rates': [], 'label': 'Kll2'},
    'pi-center-parsed': {'stream_rates': [], 'processing_rates': [], 'label': 'Centeralized'},
    'pi-center-parsed-pinned': {'stream_rates': [], 'processing_rates': [], 'label': 'Centeralized2'}
}

# Regular expression to match lines like "CountThroughput/StreamRate:_X-4 Y[m] ± Z%" or "CountThroughput/StreamRate:_X-4 Y ± Z%"
pattern = r"^(?:CountThroughput|KllThroughput|CenterThroughput)/StreamRate:_(\d+)-4\s+(\d+\.\d+m?)\s*±\s*\d+%$"

# Read and parse data from each file
for filename in data:
    with open(f"./{filename}", "r") as file:
        for line in file:
            match = re.match(pattern, line.strip())
            if match:
                # Extract stream inter-arrival time (ns)
                inter_arrival_time_ns = int(match.group(1))
                # Skip if inter_arrival_time_ns is 0 to avoid division by zero
                if inter_arrival_time_ns == 0:
                    continue
                
                # Convert to stream rate (tuples/sec): 1 / (ns / 10^9)
                stream_rate_tuples_per_sec = 10**9 / inter_arrival_time_ns
                
                # Extract the throughput value
                value_str = match.group(2)
                # Check if value is in milliseconds (has 'm') or seconds (no 'm')
                if value_str.endswith('m'):
                    # Remove 'm' and convert to float (already in ms)
                    value = float(value_str[:-1])
                else:
                    # No 'm', value is in seconds, convert to ms by multiplying by 1000
                    value = float(value_str) * 1000
                
                # Calculate processing rate (tuples/sec): (7000 / value_ms) * 1000
                processing_rate_tuples_per_sec = (7000 / value) * 1000
                if "pinned" in filename:                    
                    processing_rate_tuples_per_sec = (1400*36 / value) * 1000
                
                data[filename]['stream_rates'].append(stream_rate_tuples_per_sec)
                data[filename]['processing_rates'].append(processing_rate_tuples_per_sec)

# Create the plot
plt.figure(figsize=(12, 6))

# Plot data from each file with different colors
colors = ['b-', 'r-', 'g-','c-','m-', 'y-']
for (filename, file_data), color in zip(data.items(), colors):
    plt.plot(file_data['stream_rates'], file_data['processing_rates'], color, label=f"{file_data['label']} Processing Rate")

# Add 1:1 line
# Get the maximum stream rate and processing rate to define the range
all_stream_rates = [rate for file_data in data.values() for rate in file_data['stream_rates']]
all_processing_rates = [rate for file_data in data.values() for rate in file_data['processing_rates']]
max_rate = max(max(all_stream_rates, default=0), max(all_processing_rates, default=0))
plt.plot([0, max_rate], [0, max_rate], 'k--', label='1:1 Line')


plt.xlim(0, 0.15 * 10**8)  # Set y-axis limit to 0.17

plt.ylim(0, 1.3* 10**6)  # Set y-axis limit to 0.17

plt.xlabel('Stream Rate (tuples/sec)')
plt.ylabel('Processing Rate (tuples/sec)')
plt.title('Client Throughput Stream Rate Performance Comparison')
plt.grid(True)
plt.legend()

# Save the plot
plt.savefig("throughput_comparison_tuples_per_sec_with_line.png")
plt.show()
