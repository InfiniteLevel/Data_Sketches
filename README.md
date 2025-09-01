# Distributed Sketching

This repository contains the implementation and research for our **Master’s thesis on distributed sketching**.

The project provides a distributed system consisting of a **server**, multiple **clients**, and optional **consumers**. Clients process streaming data using sketching algorithms, periodically merging results into a global state maintained by the server. Consumers can query this global state.

---

## 🚀 Getting Started

### Start the Server

Run the server to listen for client merge requests:

```bash
go run . -port a
```

**Arguments**

- `-port` *(optional)* — Port on which the server listens (default: `8080`).

The server maintains a **global sketch state** and merges data sent from clients.

---

### Start a Client

Run a client to process a dataset, sketch it, and periodically send results to the server:

```bash
go run . -client -port a -address b -sketchType c -dataSetPath d -dataSetName e -dataSetType f -mergeRate g -streamRate h
```

**Arguments**

| Flag            | Default     | Description                                                                 |
|---------------------|-----------------|-----------------------------------------------------------------------------|
| `-port`         | `8080`      | Port of the server to connect to.                                           |
| `-address`      | `127.0.0.1` | Server IP address.                                                          |
| `-sketchType`   | `kll`       | Sketching algorithm: `kll` (KLL Sketch, default) or `count` (Count Sketch). |
| `-dataSetPath`  | `./data/PVS 1/dataset_gps.csv` | Path to the dataset `.csv` file.                                |
| `-dataSetName`  |  `speed_meters_per_second`         | Name of the dataset column to process (required if `-dataSetPath` is set).  |
| `-dataSetType`  | `float`     | Data type of the column.                                                    |
| `-mergeRate`    | `1000`      | After how many processed elements the client sends a merge request.         |
| `-streamRate`   | `10`        | Controls how quickly data is streamed. Actual rate is `10^9 / streamRate` Hz.|

---

### Start a Consumer

Run a consumer to query the server’s global state:

```bash
go run . -consumer -port a -address b
```

**Arguments**

- `-port` *(optional)* — Port of the server to connect to (default: `8080`).  
- `-address` *(optional)* — Server IP address (default: `127.0.0.1`).  

Once running, type `help` to see available commands.

---

## 🧩 Sketch Types

- **KLL Sketch (`kll`)** — Approximate quantile sketch (default).  
- **Count Sketch (`count`)** — Approximate frequency sketch.

---

## 📊 Default Dataset (not included in repo)

By default, the system uses the dataset:  
[Passive Vehicular Sensors Dataset (Kaggle)](https://www.kaggle.com/datasets/jefmenegazzo/pvs-passive-vehicular-sensors-datasets)

---

## ✨ Example Usage

Start the server:

```bash
go run . -port 8080
```

Start two clients with the Count sketch:

```bash
go run . -client -port 8080 -address 127.0.0.1 -sketchType count -dataSetPath ./data/pvs.csv -dataSetName speed_meters_per_second -dataSetType float -mergeRate 1000 -streamRate 10

go run . -client -port 8080 -address 127.0.0.1 -sketchType count -dataSetPath ./data/pvs.csv -dataSetName speed_meters_per_second -dataSetType float -mergeRate 100000 -streamRate 100
```

Start a consumer to query results:

```bash
go run . -consumer -port 8080 -address 127.0.0.1
```
