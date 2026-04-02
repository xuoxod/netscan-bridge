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

echo "================================================="
echo "🚀 Starting Intelligence Bridge (Local Dev Mode)"
echo "🔑 Auth Token: $NETSCAN_AUTH_TOKEN"
echo "📡 Port: $PORT"
echo "🌐 Allowed Origins (CORS): $ALLOWED_ORIGINS"
echo "================================================="

go run main.go
