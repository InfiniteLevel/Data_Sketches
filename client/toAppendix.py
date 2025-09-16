import re
import sys
import uuid

def parse_time_to_ms(time_str):
    """Convert time string (e.g., '1.002', '991.6m', '500.0µ') to milliseconds."""
    time_str = time_str.strip()
    if time_str.endswith('µ'):
        return float(time_str[:-1]) / 1000  # Microseconds to milliseconds
    elif time_str.endswith('m'):
        return float(time_str[:-1])  # Milliseconds
    else:
        return float(time_str) * 1000  # Seconds to milliseconds

def generate_latex_table(input_file, output_file=None):
    """Generate a LaTeX longtable from benchmark data in the input file."""
    metadata = {}
    results = []
    footnote = ""

    metadata_re = re.compile(r'^(goos|goarch|pkg):\s*(.+)$')
    result_re = re.compile(r'^(\w+)/StreamRate:_(\d+)-4\s+([\d\.]+[mµ]?)\s*±\s*∞\s*(¹)?$')
    footnote_re = re.compile(r'¹\s*(.*)$')

    try:
        with open(input_file, 'r') as f:
            lines = f.readlines()
    except FileNotFoundError:
        print(f"Error: Input file '{input_file}' not found.")
        sys.exit(1)

    for line in lines:
        line = line.strip()
        metadata_match = metadata_re.match(line)
        if metadata_match:
            key, value = metadata_match.groups()
            metadata[key] = value
            continue

        result_match = result_re.match(line)
        if result_match:
            benchmark_name, stream_rate, time, has_footnote = result_match.groups()
            stream_rate = 10**9/int(stream_rate)
            time_ms = 1000/parse_time_to_ms(time) * 1000
            results.append((benchmark_name, stream_rate, time_ms))
            continue

        footnote_match = footnote_re.match(line)
        if footnote_match:
            footnote = footnote_match.group(1)

    results.sort(key=lambda x: (x[0], x[1]), reverse=True)
    table_label = f"tab:benchmark_{uuid.uuid4().hex[:8]}"

    latex = [
        "\\begin{center}",
        "\\begin{longtable}{|l|r|r|}",
        "\\caption{"
        f"Benchmark results for Throughput on {metadata.get('goos', 'unknown')} "
        f"({metadata.get('goarch', 'unknown')}) for package \\texttt{{{metadata.get('pkg', 'unknown')}}}."
        + f"}} \\label{{{table_label}}} \\\\",
        "\\hline",
        "\\textbf{Benchmark} & \\textbf{Stream Rate} & \\textbf{Throughput (ms/op)} \\\\",
        "\\hline",
        "\\endfirsthead",
        "\\hline \\textbf{Benchmark} & \\textbf{Stream Rate} & \\textbf{Throughput (ms/op)} \\\\ \\hline",
        "\\endhead",
        "\\hline \\multicolumn{3}{r}{{Continued on next page}} \\\\",
        "\\endfoot",
        "\\hline",
        "\\endlastfoot",
    ]

    for benchmark_name, stream_rate, time_ms in results:
        latex.append(f"{benchmark_name} & {stream_rate:,} & {time_ms:.3f} \\\\")

    latex.append("\\end{longtable}")
    latex.append("\\end{center}")

    latex_code = "\n".join(latex)

    if output_file:
        try:
            with open(output_file, 'w') as f:
                f.write(latex_code)
            print(f"LaTeX longtable written to '{output_file}'")
        except Exception as e:
            print(f"Error writing to output file: {e}")
            print("Printing LaTeX code to console instead:")
            print(latex_code)
    else:
        print(latex_code)

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python script.py <input_file> [output_file]")
        sys.exit(1)

    input_file = sys.argv[1]
    output_file = sys.argv[2] if len(sys.argv) > 2 else None

    print("Required LaTeX packages: \\usepackage{longtable}, \\usepackage{booktabs}, \\usepackage{float}")
    generate_latex_table(input_file, output_file)

