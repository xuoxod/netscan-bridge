package main

import (
	"log"
	"net/http"
	"os"

	"netscan_bridge/api"
)

// CORSMiddleware handles cross-origin requests for local development.
// In production, Caddy reverse-proxies this under the same domain, so CORS isn't needed.
func CORSMiddleware(allowedOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
		}

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	authToken := os.Getenv("NETSCAN_AUTH_TOKEN")
	if authToken == "" {
		log.Fatal("FATAL: NETSCAN_AUTH_TOKEN environment variable is required to start the bridge securely.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// This completely enables our frontend to talk to this backend during dev,
	// but remains tightly closed in production if not explicitly configured.
	allowedOrigin := os.Getenv("ALLOWED_ORIGINS")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/status", api.StatusHandler)
	mux.HandleFunc("POST /api/scan", api.AuthMiddleware(authToken, api.ScanHandler))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not found"}`))
	})

	log.Printf("🚀 Intelligence Bridge actively listening on port :%s", port)
	if allowedOrigin != "" {
		log.Printf("⚠️  CORS Enabled for: %s", allowedOrigin)
	}

	// Wrap the entire mux with our CORS handler
	handler := CORSMiddleware(allowedOrigin, mux)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server critically terminated: %v", err)
	}
}
