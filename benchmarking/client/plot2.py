import matplotlib.pyplot as plt
import re

# Initialize lists to store data for each file
data = {
    'pi-count-parsed': {'stream_rates': [], 'processing_rates': [], 'label': 'Count'},
    'pi-kll-parsed': {'stream_rates': [], 'processing_rates': [], 'label': 'Kll'},
    'pi-center-parsed': {'stream_rates': [], 'processing_rates': [], 'label': 'Centeralized'},
}

# Regular expression to match lines like "CountThroughput/StreamRate:_X-4 Y[m] ± Z%" or "CountThroughput/StreamRate:_X-4 Y ± Z%"
pattern = r"^(?:CountThroughput|KllThroughput|CenterThroughput)/StreamRate:_(\d+)-(4|12)\s+(\d+\.\d+m?)\s*±\s*\d+%$"

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
                value_str = match.group(3)
                # Check if value is in milliseconds (has 'm') or seconds (no 'm')
                if value_str.endswith('m'):
                    # Remove 'm' and convert to float (already in ms)
                    value = float(value_str[:-1])
                else:
                    # No 'm', value is in seconds, convert to ms by multiplying by 1000
                    value = float(value_str) * 1000
                
                # Calculate processing rate (tuples/sec): (7000 / value_ms) * 1000
                processing_rate_tuples_per_sec = (7000 / value) * 1000
                
                data[filename]['stream_rates'].append(stream_rate_tuples_per_sec)
                data[filename]['processing_rates'].append(processing_rate_tuples_per_sec)

# Create the plot
plt.figure(figsize=(12, 6))

# Plot data from each file with different colors
colors = ['b-', 'r-', 'g-', 'y-']
for (filename, file_data), color in zip(data.items(), colors):
    plt.plot(file_data['stream_rates'], file_data['processing_rates'], color, label=f"{file_data['label']} Processing Rate")

plt.xlabel('Stream Rate (tuples/sec)')
plt.xlim(0, 0.15 * 10**8)  # Set y-axis limit to 0.17
plt.ylabel('Processing Rate (tuples/sec)')
plt.title('Client Throughput Stream Rate Performance Comparison')
plt.grid(True)
plt.legend()

# Save the plot
plt.savefig("throughput_comparison_tuples_per_sec.png")
plt.show()

