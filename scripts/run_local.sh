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
export TOKEN="${TOKEN:-}"
export ROOM_ID="${ROOM_ID:-bridge-2}"

# Smart resolution for production vs development environments
DEV_NETSCAN="/home/emhcet/private/projects/desktop/java/netscan/target/release/netscan"
if [ -f "$DEV_NETSCAN" ]; then
    export NETSCAN_BIN_PATH="$DEV_NETSCAN"
    echo "🔍 Using Development NetScan binary: $NETSCAN_BIN_PATH"
else
    # Real-world fallback: assuming the user downloaded the zipped suite
    # containing both the bridge and netscan binaries in the same folder.
    export NETSCAN_BIN_PATH="./netscan"
    echo "🔍 Using Production NetScan binary: $NETSCAN_BIN_PATH"
fi

echo "================================================="
echo "🚀 Starting Intelligence Bridge (Local Dev Mode)"
echo "🔑 Auth Token: $TOKEN"
echo "🏠 Room ID: $ROOM_ID"
echo "📡 Signaling URL: $SIGNALING_URL"
echo "📡 Port: $PORT"
echo "🌐 Allowed Origins (CORS): $ALLOWED_ORIGINS"
echo "================================================="

# Workaround for 'sudo go: command not found' issue on Linux.
# Since sudo strips user PATH variables, we build the binary first,
# then execute the resulting compiled binary.
echo "🔨 Building the bridge agent..."
go build -o bridge_app main.go

echo "🟢 Running the bridge agent ($NETSCAN_BIN_PATH) as root..."
# Ensure execution happens without Go toolchain dependency for the root session
# sudo -E preserves the environment variables (PORT, ROOM_ID, NETSCAN_BIN_PATH, etc.)
while true; do
    sudo -E ./bridge_app
    echo "🔄 Bridge agent disconnected. Restarting in 2 seconds..."
    sleep 2
done
