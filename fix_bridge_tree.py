import re

curated_bridge = """```text
📦 netscan_bridge (Edge Execution Node)
├── 🔒 constants/          # Shared system mappings & WebRTC definitions
├── 🧰 executor/           # Edge process management & reverse shells
├── 📡 logger/             # Telemetry & P2P payload chunking logic
├── 🛠️ scripts/            # Build automation & multi-platform compilation wrappers
├── 🚀 main.go             # The daemon entrypoint (Headless WebRTC)
└── 📜 README.md           # This Edge Node documentation
```"""

def fix_file(path):
    with open(path, 'r') as f:
        content = f.read()
    
    new_content = re.sub(r'```text\n\.\s*\n├──.*?```', curated_bridge, content, flags=re.DOTALL)
    
    with open(path, 'w') as f:
        f.write(new_content)

fix_file("/home/emhcet/private/projects/desktop/golang/netscan_bridge/README.md")
print("Done fixing bridge README")
