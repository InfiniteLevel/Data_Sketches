#!/usr/bin/env bash
#
# # Default number of clients
CLIENTS=10
#
# # Check if an argument is provided
if [ ! -z "$1" ]; then
    CLIENTS=$1
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
# # Run one normal instance
echo "Starting main instance..."
./sketcher &
#
# # Wait for 1 second to allow the server to start
sleep 1
#
# # Start specified number of headless clients
echo "Starting $CLIENTS headless clients..."
for ((i=1; i<=CLIENTS; i++)); do
    ./sketcher -client  &
done
#
# # Provide an option to stop all clients
echo
echo "Press any key to stop all clients..."
read -n 1 -s
#
# # Kill all instances of the program
echo "Stopping all instances..."
pkill -f sketcher

echo "All clients stopped."
exit 0
# 
