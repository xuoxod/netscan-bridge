# RMediaTech: Headless WebRTC Daemon

The `netscan_bridge` acts as a secure execution layer bridging Web Browsers (specifically Mobile Phones) to local network architectures. Instead of relying on a fragile and insecure local HTTP server (which fails against browser HTTPS Mixed-Content limitations), this daemon implements a **pure WebRTC DataChannel** payload mechanism.

## Features

- **P2P STUN/TURN**: Creates a NAT-traversing pipeline enabling mobile devices on 5G networks to securely drive desktop processes locked behind home routers.
- **Headless Command Output**: Streams raw, stdout JSON string artifacts straight over a realtime WebRTC DataChannel, bypassing network proxies entirely. *Note: Massive JSON artifacts (e.g., from deep discovery or Action Studio tools) are chunked into native DataChannel binary `Blob` / `ArrayBuffer` payloads to prevent string-limit truncation and ensure zero-corruption delivery to the browser dashboard.*
- **Rigorous Clean Teardown**: Hooks into OS interruptions (`SIGINT`) and DataChannel close events (`pagehide` / browser tab closed) to aggressively `$ kill -9` any OS-level `netscan` runaway executions.
- **Action Studio Support**: Natively translates WebRTC Action Studio payloads directly to executing `netscan` flags:
  - Deep Protocol Recon (`scan`)
  - Universal Discovery (`discover`)
  - IoT Privacy Auditor (`audit` w/ spoofing/stealth)
  - Firewall Stress Tester (`weirdpackets`)
  - Specter Harvester (`specter` w/ stealth)

## Running the Agent

Typically, users download the compiled `netscan-bridge-linux-x64.zip` from the main Dashboard, meaning it will come pre-injected with an initialized `config.yaml`.

```bash
# Example Run with Config
./bridge_agent
```

If you are developing locally or running in a Docker container, override WebRTC signaling URIs using environment flags:

```bash
TOKEN="user123_bridgeToken" \
ROOM_ID="bridge-123" \
SIGNALING_URL="http://localhost:8080/api/signal" \
./bridge_agent
```

## Integration

See the `rmediatech` main repository docs on `MOBILE_BRIDGE_SETUP_GUIDE.md` for specific instructions on painting a mobile user through the WebRTC handshake.

## Edge Node Architecture

```text
📦 netscan_bridge (Edge Execution Node)
├── 🔒 constants/          # Shared system mappings & WebRTC definitions
├── 🧰 executor/           # Edge process management & reverse shells
├── 📡 logger/             # Telemetry & P2P payload chunking logic
├── 🛠️ scripts/            # Build automation & multi-platform compilation wrappers
├── 🚀 main.go             # The daemon entrypoint (Headless WebRTC)
└── 📜 README.md           # This Edge Node documentation
```

<- **Panic Recovery:**  and  sandboxes recover from WebAssembly/subprocess panics, returning generic JSON errors rather than crashing the main MILESTONE: FULL SUITE SYNC. WebRTC persistence fixed, strict argument translations fixed. -->


## Ecosystem Architecture: Decoupled by Design

This project serves as a critical standalone layer but natively ties into the **RMediaTech** web application.

Following strict Separation of Concerns (SoC) and Decoupling principles, each core module (`netscan`, `netscan_bridge`, and `bastion`) can and should be run entirely independently:
*   **CLI**: Fully functional standalone CLI utilities.
*   **Daemon**: Extendable for background, telemetry, or radio surveillance.
*   **FFI/Scripts**: Can be easily integrated, scripted, or consumed by external programs.

This ensures zero vendor-lock-in while acting as a core microservice for RMediaTech's Universal Ingestion Pipelines.



## Ecosystem Architecture: Decoupled by Design

This project serves as a critical standalone layer but natively ties into the **RMediaTech** web application.

Following strict Separation of Concerns (SoC) and Decoupling principles, each core module (`netscan`, `netscan_bridge`, and `bastion`) can and should be run entirely independently:
*   **CLI**: Fully functional standalone CLI utilities.
*   **Daemon**: Extendable for background, telemetry, or radio surveillance.
*   **FFI/Scripts**: Can be easily integrated, scripted, or consumed by external programs.

This ensures zero vendor-lock-in while acting as a core microservice for RMediaTech's Universal Ingestion Pipelines.

