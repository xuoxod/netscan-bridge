#!/bin/bash
set -e

cd "$(dirname "$0")/.."

# Dynamic Environment resolution
# Auto-detects the workspace root by walking up the directory tree
WORKSPACE_ROOT=$(pwd)
while [ "$WORKSPACE_ROOT" != "/" ]; do
    if [ -d "$WORKSPACE_ROOT/desktop/java/netscan" ]; then
        break
    fi
    WORKSPACE_ROOT=$(dirname "$WORKSPACE_ROOT")
done

if [ "$WORKSPACE_ROOT" == "/" ]; then
    echo "❌ FATAL: Could not resolve workspace root automatically."
    echo "Please set RMEDIA_ROOT and NETSCAN_JAVA_ROOT environment variables manually."
    exit 1
fi

if [ -z "$NETSCAN_JAVA_ROOT" ]; then
    NETSCAN_JAVA_ROOT="$WORKSPACE_ROOT/desktop/java/netscan"
fi

if [ -z "$RMEDIA_ROOT" ]; then
    if [ -d "$WORKSPACE_ROOT/golang/rmediatech" ]; then
        # Production server structure
        RMEDIA_ROOT="$WORKSPACE_ROOT/golang/rmediatech"
    elif [ -d "$WORKSPACE_ROOT/desktop/golang/rmediatech" ]; then
        # Local Dev structure
        RMEDIA_ROOT="$WORKSPACE_ROOT/desktop/golang/rmediatech"
    else
        echo "❌ FATAL: Could not find rmediatech in expected locations."
        exit 1
    fi
fi

echo "🔨 Compiling Bridge Agent..."
# Compiling a standalone Linux x64 binary
GOOS=linux GOARCH=amd64 go build -o bridge_agent ./main.go

echo "📦 Creating bundling workspace..."
WORK_DIR=$(mktemp -d)
BUNDLE_DIR="$WORK_DIR/netscan-bridge"
mkdir -p "$BUNDLE_DIR"
mkdir -p "$BUNDLE_DIR/bin"

NETSCAN_BIN="$NETSCAN_JAVA_ROOT/target/release/netscan"
BASE_TAR="$RMEDIA_ROOT/static/downloads/base/netscan-suite-linux-x64.tar.gz"

if [ -f "$BASE_TAR" ]; then
    echo "📥 Extracting base tarball from $BASE_TAR..."
    tar -xzf "$BASE_TAR" -C "$WORK_DIR"
    # Rename extracted directory to match our bridge payload structure
    if [ -d "$WORK_DIR/netscan-suite-linux-x64" ]; then
        cp -r "$WORK_DIR/netscan-suite-linux-x64"/* "$BUNDLE_DIR/"
        rm -rf "$WORK_DIR/netscan-suite-linux-x64"
    fi
else
    echo "⚠️ Base tarball not found in Downloads. Validating local mock structure..."
    if [ -f "$NETSCAN_BIN" ]; then
        echo "📥 Injecting built netscan binary from $NETSCAN_BIN..."
        cp "$NETSCAN_BIN" "$BUNDLE_DIR/bin/"
        chmod +x "$BUNDLE_DIR/bin/netscan"
    else
        echo "❌ FATAL: Cannot find compiled 'netscan' binary at '$NETSCAN_BIN'."
        echo "Please build the Java netscan engine first or set NETSCAN_JAVA_ROOT."
        rm -rf "$WORK_DIR"
        exit 1
    fi
fi

echo "🔌 Injecting bridge_agent..."
cp ./bridge_agent "$BUNDLE_DIR/"

echo "📝 Writing README_FIRST.txt..."
cat << 'DOC' > "$BUNDLE_DIR/README_FIRST.txt"
===========================================================
RMediaTech Intelligence Bridge Agent
===========================================================

USER RESPONSIBILITY & SETUP INSTRUCTIONS:
This bridge agent MUST be run on a device inside the target network.
In order to execute network commands securely, you must compile the
'netscan' native engine binary from source and place the executable
at 'netscan-bridge/bin/netscan'.
DOC

echo "⚙️ Creating start script (run.sh)..."
cat << 'SCRIPT' > "$BUNDLE_DIR/run.sh"
#!/usr/bin/env bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
export PATH="$PATH:$DIR/bin"
export NETSCAN_BIN_PATH="$DIR/bin/netscan"
cd "$DIR" || exit 1
echo "🛡️  Running the bridge agent as ROOT (required for netscan)..."
exec sudo -E ./bridge_agent
SCRIPT
chmod +x "$BUNDLE_DIR/run.sh"

echo "🗜️ Packaging netscan-bridge-linux-x64.zip..."
cd "$WORK_DIR"
zip -r netscan-bridge-linux-x64.zip "netscan-bridge"

echo "🚚 Deploying artifact to rmediatech backend..."
DEST_DIR="$RMEDIA_ROOT/static/downloads"
mkdir -p "$DEST_DIR"
cp netscan-bridge-linux-x64.zip "$DEST_DIR/netscan-bridge-linux-x64.zip"
# Backward compatibility or fallback if parts of the UI still look for the tarball
tar -czf netscan-bridge-linux-x64.tar.gz "netscan-bridge/"
cp netscan-bridge-linux-x64.tar.gz "$DEST_DIR/netscan-bridge-linux-x64.tar.gz"

echo "🧹 Cleaning up..."
rm -rf "$WORK_DIR"
cd - > /dev/null

echo "✅ Success! The new payload has been deployed."
