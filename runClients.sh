#!/usr/bin/env bash
# # Default number of clients
CLIENTS=10
MERGE_RATE=10
STREAM_RATE=10
IP="127.0.0.1"
TYPE="kll"
#
# # Check if an argument is provided
if [ ! -z "$1" ]; then
    CLIENTS=$1
fi
if [ ! -z "$2" ]; then
    MERGE_RATE=$2
fi
if [ ! -z "$3" ]; then
    STREAM_RATE=$3
fi
if [ ! -z "$4" ]; then
    IP=$4
fi
if [ ! -z "$5" ]; then
    TYPE=$5
fi
#
# # Compile the Go program
echo "Compiling Go program..."
go build -o sketcher main.go
if [ $? -ne 0 ]; then
    echo "Compilation failed!"
    exit 1
fi
#
#
# # Start specified number of headless clients
echo "Starting $CLIENTS headless clients..."
for ((i=1; i<=CLIENTS; i++)); do
    ./sketcher -client -merge $MERGE_RATE -stream $STREAM_RATE -a $IP -sketch $TYPE  &
done
