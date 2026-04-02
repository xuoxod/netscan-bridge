#!/bin/bash
set -e

# Resolve the absolute path of the project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Navigate to project root to execute testing correctly across all packages
cd "$PROJECT_ROOT"

export NETSCAN_AUTH_TOKEN="dev_token_secret_123"
export PORT="8081"
export ALLOWED_ORIGINS="http://localhost:8080"

echo "================================================="
echo "🧪 Executing Automated Test Suite (Dev Mode)"
echo "================================================="

go test -v ./...
