package main

import (
"log"
"net/http"

"netscan_bridge/api"
)

func main() {
mux := http.NewServeMux()
mux.HandleFunc("GET /api/status", api.StatusHandler)

log.Println("Starting NetScan API Bridge on :8080...")
if err := http.ListenAndServe(":8080", mux); err != nil {
log.Fatalf("Server failed: %v", err)
}
}
