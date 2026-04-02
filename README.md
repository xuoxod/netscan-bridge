# NetScan API Bridge (Go Commander)

A robust, isolated REST API written in purely native Go. It serves as the **Universal Remote Broker** between connected clients (Mobile Apps, UI Dashboards) and the natively compiled execution engine (NetScan).

## Purpose 🧠
The core scanner engine interacts natively with the OS network stack requiring deep system privileges (promiscuous mode, raw sockets). Modern mobile phones strictly sandbox UI apps, preventing direct networking.

This Bridge runs locally on a trusted node (e.g., Linode server, Raspberry Pi), exposing a secure REST API. Clients trigger scans purely via HTTP requests, the Bridge executes the native core securely via subprocess, parses the intelligence, and returns purely structured JSON logic.

## Architecture
- **Pure Native Go**: Zero unneeded web frameworks. Powered purely by Go 1.22+ `net/http` standard library.
- **Strictly Sandboxed Execution**: Employs deep string sanitization and `os/exec` isolated arguments to completely prevent bash injection tactics via malicious web requests.
- **Microservice Design**: Does not share code with the core engine avoiding structural dependencies.

## Usage 🛠️

1. Start the API listener:
   ```bash
   go run main.go
   ```
2. The Bridge boots locally to `http://localhost:8080` awaiting commands.
