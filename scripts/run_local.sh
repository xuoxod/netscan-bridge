#!/bin/bash
set -e

# Resolve the absolute path of the project root regardless of where this script is called from
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Navigate to project root to ensure 'go run' finds main.go and module paths
cd "$PROJECT_ROOT"

export PORT="8081"
export ALLOWED_ORIGINS="http://localhost:8080"

# Bridge Linkage Variables (Required for the Go Bridge to join WebRTC bounds)
export SIGNALING_URL="${SIGNALING_URL:-http://localhost:8080/api/signal}"
export TOKEN="${TOKEN:-dev_token_secret_123}"
export ROOM_ID="${ROOM_ID:-bridge-2}"

# Smart resolution for production vs development environments
if [ -f "$HOME/private/projects/desktop/java/netscan/target/release/netscan" ]; then
    # DEV Environment (emhcet layout)
    export NETSCAN_BIN_PATH="$HOME/private/projects/desktop/java/netscan/target/release/netscan"
    echo "🔍 Using Development NetScan binary: $NETSCAN_BIN_PATH"
elif [ -f "$HOME/private/projects/java/netscan/target/release/netscan" ]; then
    # PROD Environment (rick layout)
    export NETSCAN_BIN_PATH="$HOME/private/projects/java/netscan/target/release/netscan"
    echo "🔍 Using Production NetScan binary: $NETSCAN_BIN_PATH"
else
    # Fallback / Bundled Suite
    export NETSCAN_BIN_PATH="./netscan"
    echo "🔍 Using Fallback NetScan binary: $NETSCAN_BIN_PATH"
fi

echo "================================================="
echo "🚀 Starting Intelligence Bridge"
echo "🔑 Auth Token: $TOKEN"
echo "🏠 Room ID: $ROOM_ID"
echo "📡 Signaling URL: $SIGNALING_URL"
echo "📡 Port: $PORT"
echo "🌐 Allowed Origins (CORS): $ALLOWED_ORIGINS"
echo "================================================="

echo "🔨 Building the bridge agent..."
go build -o bridge_app main.go

echo "🟢 Running the bridge agent ($NETSCAN_BIN_PATH) as root..."
while true; do
    sudo -E ./bridge_app
    echo "🔄 Bridge agent disconnected. Restarting in 2 seconds..."
    sleep 2
done
