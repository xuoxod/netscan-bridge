#!/bin/bash

# Task 3: Auditing Non-Source Directories
# rmediatech
RMT_DIRS=("info" "db" "scripts" "docs" "static")
for dir in "${RMT_DIRS[@]}"; do
  cd /home/emhcet/private/projects/desktop/golang/rmediatech
  mkdir -p "$dir"
  echo "[$dir Directory](https://github.com/xuoxod/rmediatech/tree/main/$dir)" > "$dir/README.md"
done

# netscan_bridge
NB_DIRS=("scripts" "executor" "logger" "constants")
for dir in "${NB_DIRS[@]}"; do
  cd /home/emhcet/private/projects/desktop/golang/netscan_bridge
  mkdir -p "$dir"
  echo "[$dir Directory](https://github.com/xuoxod/netscan-bridge/tree/main/$dir)" > "$dir/README.md"
done

