#!/usr/bin/env python3
import numpy as np
import argparse
from collections import Counter

def generate_zipf_data(stream_size, num_distinct, skew):
    """Generate Zipf-distributed data matching ASketch paper parameters."""
    print(f"Generating Zipf data: size={stream_size:,}, distinct={num_distinct:,}, skew={skew}")
    
    # Generate Zipf distribution
    zipf_samples = np.random.zipf(skew, stream_size)
    
    # Map to range [0, num_distinct)
    data = (zipf_samples - 1) % num_distinct
    
    return data

def save_to_csv(data, filename):
    """Save data to CSV format compatible with the Go client."""
    print(f"Saving to {filename}...")
    
    with open(filename, 'w') as f:
        f.write('timestamp,item_id\n')
        for i, item in enumerate(data):
            f.write(f'{i},{int(item)}\n')
    
    print(f"✓ Saved {len(data):,} items")

def analyze_data(data, top_k=10):
    """Analyze the generated data distribution."""
    print("\n=== Data Analysis ===")
    
    # Count frequencies
    counter = Counter(data)
    total = len(data)
    distinct = len(counter)
    
    print(f"Total items: {total:,}")
    print(f"Distinct items: {distinct:,}")
    
    # Top-k most frequent
    most_common = counter.most_common(top_k)
    print(f"\nTop-{top_k} most frequent items:")
    for rank, (item, count) in enumerate(most_common, 1):
        percentage = (count / total) * 100
        print(f"  #{rank}: Item {int(item)} → {count:,} times ({percentage:.2f}%)")
    
    # Heavy hitter statistics
    top_32 = counter.most_common(32)
    top_32_count = sum(count for _, count in top_32)
    top_32_percentage = (top_32_count / total) * 100
    print(f"\nTop-32 items account for: {top_32_percentage:.2f}% of all items")

def main():
    parser = argparse.ArgumentParser(description='Generate Zipf-distributed data')
    parser.add_argument('-s', '--size', type=int, default=1000000,
                        help='Stream size (default: 1M)')
    parser.add_argument('-d', '--distinct', type=int, default=250000,
                        help='Number of distinct items (default: 250K)')
    parser.add_argument('-z', '--skew', type=float, default=1.5,
                        help='Zipf skew parameter (default: 1.5)')
    parser.add_argument('-o', '--output', type=str, required=True,
                        help='Output CSV file')
    parser.add_argument('--analyze', action='store_true',
                        help='Print data analysis')
    
    args = parser.parse_args()
    
    # Generate data
    data = generate_zipf_data(args.size, args.distinct, args.skew)
    
    # Analyze if requested
    if args.analyze:
        analyze_data(data)
    
    # Save to CSV
    save_to_csv(data, args.output)
    
    print(f"\n✓ Done!")

if __name__ == '__main__':
    main()
