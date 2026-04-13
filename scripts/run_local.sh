#!/bin/bash
set -e

echo "================================================="
echo "        🚀 RMediaTech Developer Bridge           "
echo "================================================="

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BRIDGE_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$BRIDGE_ROOT"

# Put back the developer defaults
export DEV_EMAIL="$1"

if [ -z "$RMEDIA_ROOT" ]; then
    RMEDIA_ROOT="$(dirname "$BRIDGE_ROOT")/rmediatech"
fi

if [ -f "$RMEDIA_ROOT/.env" ]; then
    echo "💡 Loading backend configuration from $RMEDIA_ROOT/.env"
    set -a
    source "$RMEDIA_ROOT/.env"
    set +a
fi

export SIGNALING_URL="${RMT_SIGNALING_URL:-http://localhost:8080/api/signal}"
export TOKEN="${2:-${RMT_BRIDGE_SECRET:-dev_token_secret_123}}"

HMAC_OUT=$(printf "%s" "$DEV_EMAIL" | openssl dgst -sha256 -hmac "$TOKEN" | sed 's/^.* //')
export ROOM_ID="bridge-${HMAC_OUT:0:12}"

if [ -z "$NETSCAN_BIN_PATH" ]; then
    SIBLING_JAVA_PATH="$(dirname "$(dirname "$BRIDGE_ROOT")")/java/netscan/target/release/netscan"
    if [ -x "$SIBLING_JAVA_PATH" ]; then
        export NETSCAN_BIN_PATH="$SIBLING_JAVA_PATH"
    elif [ -x "$(pwd)/netscan" ]; then
        export NETSCAN_BIN_PATH="$(pwd)/netscan"
    else
        export NETSCAN_BIN_PATH="$(pwd)/netscan"
    fi
fi

echo "✅ Dev User     : $DEV_EMAIL"
echo "✅ Room ID      : $ROOM_ID"
echo "✅ Signaling URL: $SIGNALING_URL"
echo "✅ Netscan Path : $NETSCAN_BIN_PATH"
echo "================================================="

echo "🔨 Building the bridge agent from source..."
go build -o bridge_dev main.go

echo "🛡️  Running the bridge agent as ROOT (required for netscan)..."
while true; do
    sudo -E ./bridge_dev
    echo "🔄 Bridge agent terminated. Restarting in 2 seconds... (Ctrl+C to quit)"
    sleep 2
done
