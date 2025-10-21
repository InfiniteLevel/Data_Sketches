#!/usr/bin/env python3
import pandas as pd
import matplotlib.pyplot as plt
from collections import Counter
import numpy as np
import os
import sys

def main():
    if len(sys.argv) < 2:
        print("Usage: python3 plot_data.py <csv_file>")
        sys.exit(1)
    
    filepath = sys.argv[1]
    print(f"Loading {filepath}...")
    
    # Create plots directory
    os.makedirs('plots', exist_ok=True)
    
    # Load data
    df = pd.read_csv(filepath)
    data = df['item_id'].values
    print(f"Loaded {len(data):,} items")
    
    # Count frequencies
    counter = Counter(data)
    print(f"Found {len(counter):,} distinct items")
    
    # 1. Top-32 bar chart
    print("Generating top-32 bar chart...")
    top_32 = counter.most_common(32)
    items = [f"#{i+1}" for i in range(32)]
    freqs = [count for _, count in top_32]
    
    plt.figure(figsize=(14, 6))
    plt.bar(items, freqs, color='steelblue', edgecolor='black')
    plt.xlabel('Rank')
    plt.ylabel('Frequency')
    plt.title('Top-32 Heavy Hitters')
    plt.xticks(rotation=45)
    plt.tight_layout()
    plt.savefig('plots/top32_bar.png', dpi=300)
    print("✓ Saved plots/top32_bar.png")
    plt.close()
    
    # 2. Rank-frequency (log-log) - validates Zipf
    print("Generating rank-frequency plot...")
    all_items = counter.most_common()
    ranks = np.arange(1, len(all_items) + 1)
    frequencies = [count for _, count in all_items]
    
    plt.figure(figsize=(10, 6))
    plt.loglog(ranks, frequencies, 'b-', linewidth=2)
    plt.xlabel('Rank (log scale)')
    plt.ylabel('Frequency (log scale)')
    plt.title('Rank-Frequency Plot (Zipf Distribution)')
    plt.grid(True, alpha=0.3)
    plt.tight_layout()
    plt.savefig('plots/rank_frequency.png', dpi=300)
    print("✓ Saved plots/rank_frequency.png")
    plt.close()
    
    # 3. Cumulative distribution
    print("Generating cumulative distribution...")
    total = sum(counter.values())
    cumulative = []
    running_sum = 0
    
    for _, count in all_items:
        running_sum += count
        cumulative.append((running_sum / total) * 100)
    
    plt.figure(figsize=(10, 6))
    plt.plot(range(1, len(cumulative) + 1), cumulative, 'b-', linewidth=2)
    plt.axhline(y=80, color='r', linestyle='--', label='80%')
    plt.axhline(y=86.57, color='g', linestyle='--', label='86.57% (top-32)')
    plt.xlabel('Number of Unique Items')
    plt.ylabel('Cumulative Percentage (%)')
    plt.title('Cumulative Frequency Distribution')
    plt.legend()
    plt.grid(True, alpha=0.3)
    plt.tight_layout()
    plt.savefig('plots/cumulative.png', dpi=300)
    print("✓ Saved plots/cumulative.png")
    plt.close()
    
    # Print summary
    print("\n=== Summary ===")
    for i, (item, count) in enumerate(top_32[:10], 1):
        pct = (count / total) * 100
        print(f"  #{i}: Item {item} → {count:,} ({pct:.2f}%)")
    
    top_32_sum = sum(count for _, count in top_32)
    top_32_pct = (top_32_sum / total) * 100
    print(f"\nTop-32 items: {top_32_pct:.2f}% of all data")
    
    print("\n✓ All plots saved to plots/ directory")

if __name__ == '__main__':
    main()
