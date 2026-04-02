<!-- markdownlint-disable-file MD041 -->

# The API Bridge (Universal Remote Broker)

![Architecture](https://img.shields.io/badge/Architecture-Universal_Broker-00ADD8?style=for-the-badge)
![Transport](https://img.shields.io/badge/Transport-Secure_REST_API-FF6C37?style=for-the-badge)
![Security](https://img.shields.io/badge/Security-Air_Gapped_Subprocess-7952B3?style=for-the-badge)

A robust, isolated REST API written in a statically typed, compiled systems language. It serves as the **Universal Remote Broker** between connected clients (Mobile Apps, Web Dashboards) and the natively compiled execution engine (The Native Core).

## 🧠 Purpose

The core scanner engine interacts natively with the OS network stack requiring deep system privileges (promiscuous mode, raw sockets). Modern mobile sandboxes prevent these direct networking operations.

This Bridge runs locally on a trusted node (e.g., cloud server, local intelligence hub), exposing a secure API. Clients trigger actions purely via HTTP requests; the Bridge executes the native core securely via an isolated subprocess, parses the raw output, and returns purely structured JSON intelligence.

## 🌳 The Ecosystem Topology

```text
[ The Infrastructure ]
├── 📱 The Client (Mobile/Web)      <- The Commander
│   └── Triggers remote execution via standard HTTP
└── 🌉 The API Bridge               <- The Proxy
    ├── 🛡️ Token-Based Authorization
    ├── 🧹 Input Sanitization (Anti-Shell Injection)
    └── 🦀 The Native Core          <- The Execution Engine (Isolated subprocess)
```

## 🏗️ Architecture

- **Zero-Dependency Build**: No third-party web frameworks. Powered entirely by the native standard library to drastically reduce the application's attack surface.
- **Strictly Sandboxed Execution**: Employs deep string sanitization and isolated parameter arrays to prevent shell injection tactics via malicious web payloads.
- **Air-Gapped Design**: Operates in complete isolation from the native core engine, avoiding architectural coupling or shared memory spaces.

## ⚙️ System Requirements & Developer Environment

To compile and run the API Bridge, the following developer toolchain is required:

```text
[ The Environment ]
├── 🐹 Runtime Compiler   <- Go 1.22+ (Leveraged for robust, native HTTP pattern-matching).
└── 🦀 The Native Core    <- Pre-compiled executable available in the system $PATH.
```

## 🛠️ Usage

1. Compile the API listener:

   ```bash
   go build -o api-bridge main.go
   ./api-bridge
   ```

2. The Bridge boots locally to `http://localhost:8080` awaiting commands.

## Mobile Remote Control (Intelligence Bridge Architecture)

A critical design feature of the `rmediatech` and `netscan` ecosystem is its **"Mobile Remote Control"** capability, explicitly built to bypass the restrictions of mobile browsers and eliminate the need for native iOS/Android App Store applications.

### How It Works:
1. **The Intelligence Bridge (`netscan_bridge`)**: A user downloads the lightweight Bridge API executable (which we configure to embed the native Rust engine) and runs it on a desktop, laptop, or Raspberry Pi connected to their local home network. 
2. **The Smart UI (`rmediatech`)**: The user visits the `rmediatech` web dashboard on their mobile phone. The web application detects the device type and dynamically adapts its UI to function as a sleek remote control rather than a standard web app.
3. **Execution**: The user commands a "Live Recon" via their phone. The web application fires a secure webhook directly to the Bridge API running on their desktop on port `:8081`. The desktop executes the native network sweep on the local network and returns the raw JSON payload back to the phone.
4. **Ingestion**: The `rmediatech` frontend natively ingests this payload, parsing the JSON report and automatically rendering the results via the Universal Ingestion Pipeline.

### Device-Specific User Flows:
- **Desktop/Laptop Users:** These users have the native capability to run CLIs or use `localhost`. The web UI elegantly exposes standard file-upload modalities alongside a `localhost:8081` Live Recon trigger.
- **Mobile Users:** Mobile users cannot run low-level packet sweeps themselves. Therefore, the web UI automatically hides standard upload buttons and exclusively provides "Remote Control" inputs. Mobile users simply provide the local IP address of their computer running the Intelligence Bridge (e.g., `192.168.1.50`) and the UI handles the entire orchestration.

## Mobile Remote Control (Intelligence Bridge Architecture)

A critical design feature of the `rmediatech` and `netscan` ecosystem is its **"Mobile Remote Control"** capability, explicitly built to bypass the restrictions of mobile browsers and eliminate the need for native iOS/Android App Store applications.

### How It Works:
1. **The Intelligence Bridge (`netscan_bridge`)**: A user downloads the lightweight Bridge API executable (which we configure to embed the native Rust engine) and runs it on a desktop, laptop, or Raspberry Pi connected to their local home network. 
2. **The Smart UI (`rmediatech`)**: The user visits the `rmediatech` web dashboard on their mobile phone. The web application detects the device type and dynamically adapts its UI to function as a sleek remote control rather than a standard web app.
3. **Execution**: The user commands a "Live Recon" via their phone. The web application fires a secure webhook directly to the Bridge API running on their desktop on port `:8081`. The desktop executes the native network sweep on the local network and returns the raw JSON payload back to the phone.
4. **Ingestion**: The `rmediatech` frontend natively ingests this payload, parsing the JSON report and automatically rendering the results via the Universal Ingestion Pipeline.

### Device-Specific User Flows:
- **Desktop/Laptop Users:** These users have the native capability to run CLIs or use `localhost`. The web UI elegantly exposes standard file-upload modalities alongside a `localhost:8081` Live Recon trigger.
- **Mobile Users:** Mobile users cannot run low-level packet sweeps themselves. Therefore, the web UI automatically hides standard upload buttons and exclusively provides "Remote Control" inputs. Mobile users simply provide the local IP address of their computer running the Intelligence Bridge (e.g., `192.168.1.50`) and the UI handles the entire orchestration.

## Phase 2: Intelligence Action Studio & Deep Recon
The UI now features a unified Single Page Application (SPA) Action Studio that dynamically aggregates Discovery Sweeps and enables targeted Deep Protocol Recon (Port Scans) with automated rust-native artifact ingestion natively through the bridge.
