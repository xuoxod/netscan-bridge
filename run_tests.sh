#!/bin/bash

# Load the exact same development environment variables as our local runner
export NETSCAN_AUTH_TOKEN="dev_token_secret_123"
export PORT="8081"
export ALLOWED_ORIGINS="http://localhost:8080"

echo "================================================="
echo "🧪 Executing Automated Test Suite (Dev Mode)"
echo "🔑 Auth Token: $NETSCAN_AUTH_TOKEN"
echo "================================================="

# Execute all Go tests with verbosity
go test -v ./...
