package main

import (
"log"
"net/http"
"os"

"netscan_bridge/api"
)

func main() {
// The API will aggressively drop POSTs without this matching Bearer token
authToken := os.Getenv("NETSCAN_AUTH_TOKEN")
if authToken == "" {
log.Fatal("FATAL: NETSCAN_AUTH_TOKEN environment variable is required to start the bridge securely.")
}

// Default to 8081 to avoid conflicting with rmediatech running on 8080
port := os.Getenv("PORT")
if port == "" {
port = "8081"
}

mux := http.NewServeMux()

// Go 1.22+ Method-based routing
mux.HandleFunc("GET /api/status", api.StatusHandler)
mux.HandleFunc("POST /api/scan", api.AuthMiddleware(authToken, api.ScanHandler))

// Catch-all to 404 for undefined routes
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusNotFound)
w.Write([]byte(`{"error":"not found"}`))
})

log.Printf("🚀 Intelligence Bridge actively listening on port :%s", port)
if err := http.ListenAndServe(":"+port, mux); err != nil {
log.Fatalf("Server critically terminated: %v", err)
}
}
