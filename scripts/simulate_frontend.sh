#!/bin/bash
set -e

echo "📡 Simulating API Post from http://localhost:8080..."

curl -i -X POST http://localhost:8081/api/scan \
  -H "Authorization: Bearer dev_token_secret_123" \
  -H "Content-Type: application/json" \
  -H "Origin: http://localhost:8080" \
  -d '{"target":"127.0.0.1"}'

echo -e "\n\n✅ Simulation Complete."
