#!/bin/bash
set -e

cd "$(dirname "$0")/.."

echo "🔨 Compiling Bridge Agent..."
# Make sure we compile a standalone Linux x64 binary
GOOS=linux GOARCH=amd64 go build -o bridge_agent ./main.go

echo "📦 Creating bundling workspace..."
WORK_DIR=$(mktemp -d)
BUNDLE_DIR="$WORK_DIR/netscan-suite-linux-x64"
mkdir -p "$BUNDLE_DIR"

if [ -f "$HOME/Downloads/GZs/netscan-suite-linux-x64.tar.gz" ]; then
    echo "📥 Extracting base tarball from Downloads..."
    tar -xzf "$HOME/Downloads/GZs/netscan-suite-linux-x64.tar.gz" -C "$WORK_DIR"
else
    echo "⚠️ Base tarball not found in Downloads. Creating a mock structure..."
    mkdir -p "$BUNDLE_DIR/bin"
    # Copy from the dev server location if available
    cp "$HOME/private/projects/desktop/java/netscan/target/release/netscan" "$BUNDLE_DIR/bin/" || echo "Warning: No netscan binary found locally"
fi

echo "🔌 Injecting bridge_agent..."
cp ./bridge_agent "$BUNDLE_DIR/"

echo "🗜️ Repackaging netscan-suite-linux-x64.tar.gz..."
cd "$WORK_DIR"
tar -czf netscan-suite-linux-x64.tar.gz "netscan-suite-linux-x64/"

echo "🚚 Replacing rmediatech download artifact..."
DEST_DIR="$HOME/private/projects/desktop/golang/rmediatech/static/downloads"
mkdir -p "$DEST_DIR"
cp netscan-suite-linux-x64.tar.gz "$DEST_DIR/netscan-suite-linux-x64.tar.gz"

echo "🧹 Cleaning up..."
rm -rf "$WORK_DIR"
cd - > /dev/null

echo "✅ Success! The new tarball has been deployed to the rmediatech static/downloads folder."
