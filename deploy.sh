#!/bin/bash
set -euo pipefail

APP_DIR="$(dirname "$(realpath "$0")")"
BINARY="$APP_DIR/ahri-health-bridge"
PORT=8081
PID_FILE="$APP_DIR/app.pid"
LOG_FILE="$APP_DIR/app.log"

graceful_kill() {
    local pid=$1
    local label=$2
    if ps -p "$pid" > /dev/null 2>&1; then
        echo "Stopping $label (PID $pid)..."
        kill "$pid" 2>/dev/null || true
        local i=0
        while ps -p "$pid" > /dev/null 2>&1 && [ $i -lt 10 ]; do
            sleep 1; i=$((i+1))
        done
        if ps -p "$pid" > /dev/null 2>&1; then
            echo "Process did not exit gracefully, sending SIGKILL..."
            kill -9 "$pid" 2>/dev/null || true
        fi
    fi
}

echo "--- Starting Deployment for ahri-health-bridge ---"

# 1. Stop process from PID file
if [ -f "$PID_FILE" ]; then
    OLD_PID=$(cat "$PID_FILE")
    graceful_kill "$OLD_PID" "old process from pidfile"
    rm "$PID_FILE"
fi

# 2. Stop any remaining process on the port (fuser works on both Linux and macOS)
echo "Checking for processes on port $PORT..."
if fuser "$PORT/tcp" > /dev/null 2>&1; then
    PIDS=$(fuser "$PORT/tcp" 2>/dev/null)
    for pid in $PIDS; do
        graceful_kill "$pid" "process on port $PORT"
    done
fi

# 3. Build binary
echo "Building ahri-health-bridge..."
cd "$APP_DIR"
go build -o "$BINARY" .

# 4. Start the application
echo "Starting ahri-health-bridge on port $PORT..."
PORT=$PORT nohup "$BINARY" > "$LOG_FILE" 2>&1 &
NEW_PID=$!
echo $NEW_PID > "$PID_FILE"
echo "Application started with PID $NEW_PID"

# 5. Verify the process is still running and port is open
echo "Waiting for application to come up..."
for i in $(seq 1 15); do
    if ! ps -p "$NEW_PID" > /dev/null 2>&1; then
        echo "Verification FAILED: Process exited. Check $LOG_FILE"
        tail -n 20 "$LOG_FILE"
        exit 1
    fi
    if fuser "$PORT/tcp" > /dev/null 2>&1; then
        echo "Verification SUCCESS: Process is running and listening on port $PORT."
        echo "Tail of logs:"
        tail -n 10 "$LOG_FILE"
        echo "--- Deployment Complete ---"
        exit 0
    fi
    sleep 1
done

echo "Verification FAILED: Process running but port $PORT not open after 15s. Check $LOG_FILE"
tail -n 20 "$LOG_FILE"
exit 1
