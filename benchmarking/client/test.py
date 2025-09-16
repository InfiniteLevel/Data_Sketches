import re

text = """CountThroughput/StreamRate:_4780-4        924.4m ±   14%
CountThroughput/StreamRate:_4790-4        942.3m ±    7%
CountThroughput/StreamRate:_4810-4         1.033 ±   11%
geomean                                   294.1m"""

pattern = r"^CountThroughput/StreamRate:_(\d+-\d)\s+(\d+\.\d+(?:m|ms)?)\s*±\s*(\d+)%$"
for line in text.splitlines():
    match = re.match(pattern, line)
    if match:
        print(f"Stream: {match.group(1)}, Value: {match.group(2)}, Error: {match.group(3)}%")
