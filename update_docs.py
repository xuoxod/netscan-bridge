import os
import subprocess
import re

rmt_path = "/home/emhcet/private/projects/desktop/golang/rmediatech"
nb_path = "/home/emhcet/private/projects/desktop/golang/netscan_bridge"

def get_tree(path):
    result = subprocess.run(["tree", "-L", "3"], cwd=path, capture_output=True, text=True)
    return f"```text\n{result.stdout}\n```"

rmt_tree = get_tree(rmt_path)
nb_tree = get_tree(nb_path)

# Regex to find markdown codeblocks that have tree ASCII characters
tree_regex = re.compile(r'```text.*?├──.*?```', re.DOTALL)

def replace_trees(file_path, tree_text):
    if not os.path.exists(file_path):
        return
    with open(file_path, "r") as f:
        content = f.read()
    content = tree_regex.sub(tree_text, content)
    with open(file_path, "w") as f:
        f.write(content)

replace_trees(os.path.join(rmt_path, "README.md"), rmt_tree)
replace_trees(os.path.join(rmt_path, "info/repo.md"), rmt_tree)
replace_trees(os.path.join(nb_path, "README.md"), nb_tree)

def enhance_edge_computing(file_path):
    if not os.path.exists(file_path):
        return
    with open(file_path, "r") as f:
        content = f.read()
    
    # rmediatech/README.md specific replacements
    content = content.replace("Mobile Remote Control", "Universal Edge Orchestration")
    
    # Inject Edge Computing concepts
    if "Zero-Trust Edge Computing" not in content:
        content.replace("zero-compromise platform", "Zero-Trust Edge Computing platform")
        if "Network Intelligence" in content:
             content = content.replace(
                "platform for **Network Intelligence**", 
                "platform for **Zero-Trust Edge Computing** and **Network Intelligence**"
            )
        
    with open(file_path, "w") as f:
        f.write(content)

enhance_edge_computing(os.path.join(rmt_path, "README.md"))
enhance_edge_computing(os.path.join(nb_path, "README.md"))

print("Done")
