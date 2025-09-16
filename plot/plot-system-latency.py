import matplotlib.pyplot as plt
from collections import defaultdict
import json
import re
import numpy as np

def readData(path: str):
    with open(path,"r") as f:
        return f.readlines()

def nested_dict():
    return defaultdict(nested_dict)

def parseData(lines: list[str]): 

    data = nested_dict()


    for i in range(0, len(lines), 2):
        header = lines[i]
        array_line = lines[i + 1]
    
        # Extract numbers using regex
        nums = list(map(int, re.findall(r'\d+', header)))
        clients, merge_rate, stream_rate = nums

        # Parse the array line
        values = list(map(int, re.findall(r'\d+', array_line)))


        # Assign to the nested dict
        data[stream_rate][merge_rate][clients] = [x /1000 for x in values]

    # Convert to normal dicts (optional)
    parsed = json.loads(json.dumps(data))
    return parsed




def verticalBoxPlots(data: dict[str, list[int]], ax, title=""):
    keys = sorted(data.keys())
    values = [data[k] for k in keys]

    ax.boxplot(values, positions=range(len(keys)), vert=True, whis=1.5)

    ax.set_xticks(range(len(keys)))
    ax.set_xticklabels(keys, rotation=30, ha='right')
    ax.set_title(title, fontsize=12)         # Push subplot title upward
    ax.set_xlabel("Clients")             # Push x-label downward a bit
    ax.set_ylabel("Time μs")








def plot(data, sketchType):
    for streamRate, mergeRates in data.items():
        num_plots = len(mergeRates)
        cols = 2
        rows = (num_plots + 1) // cols  # Only enough rows for actual plots

        fig, axs = plt.subplots(rows, cols, figsize=(6 * cols, 5 * rows), sharey=True)
        axs = axs.flatten()

        for ax in axs[num_plots:]:
            fig.delaxes(ax)  # Remove unused axes

        fig.suptitle(f"{sketchType} system queuing latency with Streamrate of {streamRate}", fontsize=16)

        for i, (mergeRate, clients) in enumerate(mergeRates.items()):
            verticalBoxPlots(clients, axs[i], f"Merge Rate Every: {mergeRate}")

        fig.subplots_adjust( hspace=0.5)

        fig.tight_layout(rect=[0.15, 0, 0.15, 0.95])
        plt.show()

        
        
    


def main():
    rawKllData = readData("../benchmarking/system/kll-system-latency")
    parsedKll = parseData(rawKllData)
    plot(parsedKll, "KLL")


    rawCountData = readData("../benchmarking/system/count-system-latency")
    parsedCount = parseData(rawCountData)
    plot(parsedCount, "Count")

    

if __name__ == "__main__":
    main()
