import re

curated_rmediatech = """```text
📦 rmediatech (Core Platform)
├── ⚙️ cmd/                # Server entrypoints & middleware
├── 🗄️ db/                 # Database SQLite schemas, migrations & seeds
├── 📚 docs/               # Core design specifications & logic architecture
├── 🧠 internal/           # Go backend engine (Auth, WebRTC, Database Store, Scanning)
├── 🛠️ scripts/            # CI/CD deployment & colorful DB Developer CLI suite
├── 🎨 static/             # Frontend assets (Bootstrap, pure CSS, downloads payload)
│   └── 📜 js/             # Zero-Trust WebRTC, Analytics, & UI Controllers
└── 🖼️ views/              # Jet templated server-side rendered HTML
```"""

def fix_file(path):
    with open(path, 'r') as f:
        content = f.read()
    
    # Target code blocks that contain a massive tree output
    new_content = re.sub(r'```text\n\.\n├──.*?```', curated_rmediatech, content, flags=re.DOTALL)
    
    # Also target the mhcet terminal block if present
    new_content = re.sub(r'mhcet@.*?tree -La.*?\n\.\n├──.*', curated_rmediatech, new_content, flags=re.DOTALL)
    
    with open(path, 'w') as f:
        f.write(new_content)
    print(f"Fixed {path}")

fix_file("/home/emhcet/private/projects/desktop/golang/rmediatech/README.md")
fix_file("/home/emhcet/private/projects/desktop/golang/rmediatech/info/repo.md")

