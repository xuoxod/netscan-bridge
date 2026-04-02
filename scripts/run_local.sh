#!/bin/bash
set -e

# Resolve the absolute path of the project root regardless of where this script is called from
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Navigate to project root to ensure 'go run' finds main.go and module paths
cd "$PROJECT_ROOT"

export NETSCAN_AUTH_TOKEN="dev_token_secret_123"
export PORT="8081"
export ALLOWED_ORIGINS="http://localhost:8080"

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
echo "🔑 Auth Token: $NETSCAN_AUTH_TOKEN"
echo "📡 Port: $PORT"
echo "🌐 Allowed Origins (CORS): $ALLOWED_ORIGINS"
echo "================================================="

# Workaround for 'sudo go: command not found' issue on Linux.
# Since sudo strips user PATH variables, we build the binary first,
# then execute the resulting compiled binary.
echo "🔨 Building the bridge agent..."
go build -o bridge_app main.go

echo "🟢 Running the bridge agent ($NETSCAN_BIN_PATH).."
# Ensure execution happens without Go toolchain dependency for the root session
./bridge_app
