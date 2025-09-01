import matplotlib.pyplot as plt
import re

# Initialize lists to store data for each file
data = {
    'pi-count-parsed': {'stream_rates': [], 'sec_per_op': [], 'label': 'Count'},
    'pi-kll-parsed': {'stream_rates': [], 'sec_per_op': [], 'label': 'Kll'},
    'pi-center-parsed': {'stream_rates': [], 'sec_per_op': [], 'label': 'Centeralized'}
}

# Regular expression to match lines like "CountThroughput/StreamRate:_X-4 Y[m] ± Z%" or "CountThroughput/StreamRate:_X-4 Y ± Z%"
pattern = r"^(?:CountThroughput|KllThroughput|CenterThroughput)/StreamRate:_(\d+)-4\s+(\d+\.\d+m?)\s*±\s*\d+%$"

# Read and parse data from each file
for filename in data:
    with open(f"./{filename}", "r") as file:
        for line in file:
            match = re.match(pattern, line.strip())
            if match:
                stream_rate = int(match.group(1))
                # Extract the throughput value
                value_str = match.group(2)
                # Check if value is in milliseconds (has 'm') or seconds (no 'm')
                if value_str.endswith('m'):
                    # Remove 'm' and convert to float (already in ms)
                    value = float(value_str[:-1])
                else:
                    # No 'm', value is in seconds, convert to ms by multiplying by 1000
                    value = float(value_str) * 1000
                data[filename]['stream_rates'].append(stream_rate)
                data[filename]['sec_per_op'].append(7000/value)

# Create the plot
plt.figure(figsize=(12, 6))

# Plot data from each file with different colors
colors = ['b-', 'r-', 'g-']
for (filename, file_data), color in zip(data.items(), colors):
    plt.plot(file_data['stream_rates'], file_data['sec_per_op'], color, label=f"{file_data['label']} Datapoints per millisecond")

plt.xlabel('Stream inter-arrival time (ns)')
plt.ylabel('Data points per ms (khz)')
plt.title('Client Throughput Stream Rate Performance Comparison')
plt.ylim(0, 0.17 * 7000)  # Set y-axis limit to 0.17
plt.grid(True)
plt.legend()

# Save the plot
plt.savefig("throughput_comparison.png")
