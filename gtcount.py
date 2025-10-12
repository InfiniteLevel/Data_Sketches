import csv, sys

path, col, val = sys.argv[1], sys.argv[2], sys.argv[3]
cnt = 0
with open(path, newline='') as f:
    r = csv.DictReader(f)
    for row in r:
        v = (row.get(col, "") or "").strip().strip('"')
        if v == val:
            cnt += 1
print(cnt)

