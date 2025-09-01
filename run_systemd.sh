#!/bin/bash

# Variables
CLIENTS=10
SERVICE_NAME="sketcher"  # Change this to match your service name
PROJECT_PATH="$(pwd)"  # Get the current directory (where the script is run)
SERVICE_FILE="$PROJECT_PATH/server/$SERVICE_NAME.service"  # Service file inside project
SYSTEMD_PATH="/etc/systemd/system/$SERVICE_NAME.service"  # Target systemd path

# Check if an argument for clients is provided
if [ ! -z "$1" ]; then
    CLIENTS=$1
fi

echo "📂 Project Path: $PROJECT_PATH"
echo "🔍 Looking for service file: $SERVICE_FILE"

# Ensure the service file exists in the project directory
if [ ! -f "$SERVICE_FILE" ]; then
    echo "❌ Service file $SERVICE_FILE not found! Please create it first."
    exit 1
fi

# Compile the Go program
echo "Compiling Go program..."
go build -o sketcher main.go
if [ $? -ne 0 ]; then
    echo "Compilation failed!"
    exit 1
fi

# Remove existing symlink (if any)
if [ -L "$SYSTEMD_PATH" ]; then
    sudo rm "$SYSTEMD_PATH"
    echo "🔄 Removed existing systemd symlink."
elif [ -f "$SYSTEMD_PATH" ]; then
    echo "⚠️ A file already exists at $SYSTEMD_PATH. Remove it manually before running this script."
    exit 1
fi

# Create symlink for systemd service
sudo ln -s "$SERVICE_FILE" "$SYSTEMD_PATH"
echo "✅ Symlink created: $SERVICE_FILE → $SYSTEMD_PATH"

# Reload systemd
sudo systemctl daemon-reload
echo "🔄 Systemd reloaded."

# Enable the service (start on boot)
# sudo systemctl enable "$SERVICE_NAME"
# echo "🚀 Service enabled to start on boot."

# Start the main server via systemd
sudo systemctl restart "$SERVICE_NAME"
echo "✅ Main server started via systemd."

# Wait for 1 second to allow the server to start
sleep 1

# Start specified number of headless clients
echo "Starting $CLIENTS headless clients..."
for ((i=1; i<=CLIENTS; i++)); do
    ./sketcher -client >/dev/null 2>&1 &
done

# Provide an option to stop all clients
echo
echo "Press any key to stop all clients..."
read -n 1 -s

# Kill all instances of the program
echo "Stopping all instances..."
pkill -f sketcher

echo "All clients stopped."
exit 0
