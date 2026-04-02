package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"netscan_bridge/executor"
)

// RunScan isolates the actual OS execution so we can cleanly mock it during automated TDD sweeps
var RunScan = executor.ExecuteScan

// ScanRequest tightly structures exactly what clients are permitted to transmit
type ScanRequest struct {
	Target   string   `json:"target"`
	ScanType string   `json:"scan_type,omitempty"`
	Flags    []string `json:"flags,omitempty"`
}

// StatusHandler natively ping-backs a healthy 200 JSON payload.
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok"}`)
}

// ScanHandler marshals external network JSON directly into safe execution commands
func ScanHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Ensure clients don't mistakenly use GET to trigger modifying jobs
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, `{"error":"method not allowed"}`)
		return
	}

	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"malformed JSON payload"}`)
		return
	}

	// Our execution engine requires a target, immediately reject unroutable executions
	if req.Target == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"error":"missing required field: target"}`)
		return
	}

	if req.ScanType == "" {
		req.ScanType = "discover"
	}

	// 🚀 Send the validated schema to our sandboxed OS executor
	output, err := RunScan(r.Context(), req.Target, req.ScanType, req.Flags...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// We purposefully bubble the stdout/stderr string error back wrapped in valid JSON
		escapedErr := strings.ReplaceAll(err.Error(), `"`, `\"`)
		fmt.Fprintf(w, `{"error":"%s"}`, escapedErr)
		return
	}

	// Execution succeeded, instantly hand the raw pipeline JSON struct directly the original client
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, output)
}

// AuthMiddleware sits fundamentally in front of targeted handlers, ruthlessly asserting valid tokens
func AuthMiddleware(validToken string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		// Missing Request Auth Structure check
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"error":"unauthorized"}`)
			return
		}

		// Secret Key integrity check
		clientToken := strings.TrimPrefix(authHeader, "Bearer ")
		if clientToken != validToken {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"error":"forbidden"}`)
			return
		}

		// Clear to proceed to specific handler
		next.ServeHTTP(w, r)
	}
}
