# The API Bridge (Universal Remote Broker)

A robust, isolated REST API written in a statically typed, compiled systems language. It serves as the **Universal Remote Broker** between connected clients (Mobile Apps, Web Dashboards) and the natively compiled execution engine (The Native Core).

## Purpose 🧠
The core scanner engine interacts natively with the OS network stack requiring deep system privileges (promiscuous mode, raw sockets). Modern mobile sandboxes prevent these direct networking operations.

This Bridge runs locally on a trusted node (e.g., cloud server, local intelligence hub), exposing a secure API. Clients trigger actions purely via HTTP requests; the Bridge executes the native core securely via an isolated subprocess, parses the raw output, and returns purely structured JSON intelligence.

## Architecture
- **Zero-Dependency Build**: No third-party web frameworks. Powered entirely by the native standard library to drastically reduce the application's attack surface.
- **Strictly Sandboxed Execution**: Employs deep string sanitization and isolated parameter arrays to prevent shell injection tactics via malicious web payloads.
- **Air-Gapped Design**: Operates in complete isolation from the native core engine, avoiding architectural coupling or shared memory spaces.

## System Requirements ⚙️
To compile and run the API Bridge, the following developer toolchain is required:
- **Go 1.22+** (Leveraged for its robust, native `net/http` pattern-matching capabilities)

## Usage 🛠️

1. Start the API listener (or compile to an executable binary):
   ```bash
   go run main.go
   ```
2. The Bridge boots locally to `http://localhost:8080` awaiting commands.
