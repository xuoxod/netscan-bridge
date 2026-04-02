#!/bin/bash

# Generates a random token for local dev, or uses a hardcoded one
export NETSCAN_AUTH_TOKEN="dev_token_secret_123"
export PORT="8081"

echo "================================================="
echo "🚀 Starting NetScan Bridge (Local Dev Mode)"
echo "🔑 Auth Token: $NETSCAN_AUTH_TOKEN"
echo "📡 Port: $PORT"
echo "================================================="

# Use go run so any changes pick up immediately without needing to separately build
go run main.go
