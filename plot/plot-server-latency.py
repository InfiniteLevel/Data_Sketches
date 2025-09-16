import matplotlib.pyplot as plt

def readData(path: str):
    with open(path,"r") as f:
        return f.readlines()

def parseData(rawArr: list[str]) -> dict[str, list[int]]:
    dic : dict[str,list[int]] = {}
    num = ""
    for i in rawArr:
        if i[0] == "m":
            s = i.split(" ")
            num = s[2].strip()
            dic[num] = []
        elif i == "\n":
            continue
        else:
            dic[num].append(int(int(i.strip())/1000))
    return dic

def verticalBoxPlots(data: dict[str,list[int]],xl,yl,tit):
    keys = sorted(data.keys())
    values = [data[k] for k in keys]

    _ = plt.boxplot(values, positions=range(len(keys)), vert=True, whis=1.5)

    _ = plt.xticks(range(len(keys)), labels=keys)
    plt.xlabel(xl)
    plt.ylabel(yl)
    plt.title(tit)
    plt.show()


    


def main():
    rawKLL = readData("../benchmarking/server/kll-Latency-14-02-11@2025-04-15")
    kllData = parseData(rawKLL)
    verticalBoxPlots(kllData, "Merge Size", "Time μs", "Request latency for different merge sizes using KLL")

    rawCount = readData("../benchmarking/server/count-Latency-11-54-17@2025-04-15")
    countData = parseData(rawCount)
    verticalBoxPlots(countData, "Merge Size", "Time μs", "Request latency for different merge sizes using Count")
    

if __name__ == "__main__":
    main()
