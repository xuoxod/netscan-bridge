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
