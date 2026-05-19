#!/bin/bash

# Configuration
APP_DIR="/root/projects/ahri/tools/ahri-health-bridge"
PORT=8081
PID_FILE="$APP_DIR/app.pid"
LOG_FILE="$APP_DIR/app.log"

echo "--- Starting Deployment for ahri-health-bridge ---"

# 1. Kill any process listening on the target port
echo "Checking for processes on port $PORT..."
PID_ON_PORT=$(lsof -t -i:$PORT)

if [ ! -z "$PID_ON_PORT" ]; then
    echo "Killing process $PID_ON_PORT on port $PORT..."
    kill -9 $PID_ON_PORT
    sleep 2
fi

# 2. Kill based on PID file if it exists and process is still running
if [ -f "$PID_FILE" ]; then
    OLD_PID=$(cat "$PID_FILE")
    if ps -p $OLD_PID > /dev/null; then
        echo "Killing old process $OLD_PID from pidfile..."
        kill -9 $OLD_PID
    fi
    rm "$PID_FILE"
fi

# 3. Clean start - navigate to directory
cd "$APP_DIR" || { echo "Directory $APP_DIR not found"; exit 1; }

# 4. Start the application in the background
# We use PORT=8081 env var to override default if the app supports it, 
# or ensure the Go app is configured to listen on $PORT.
echo "Starting ahri-health-bridge on port $PORT..."
PORT=$PORT nohup go run . > "$LOG_FILE" 2>&1 &
NEW_PID=$!

# 5. Save PID
echo $NEW_PID > "$PID_FILE"
echo "Application started with PID $NEW_PID"

# 6. Verification
sleep 3
if ps -p $NEW_PID > /dev/null; then
    echo "Verification SUCCESS: Process is running."
    echo "Tail of logs:"
    tail -n 10 "$LOG_FILE"
else
    echo "Verification FAILED: Process is not running. Check $LOG_FILE"
    exit 1
fi

echo "--- Deployment Complete ---"
