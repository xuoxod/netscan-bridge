package api

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockRunScanSuccess simulates the Executor returning a valid native engine completion
func mockRunScanSuccess(ctx context.Context, target string, flags ...string) (string, error) {
	return `{"network":"` + target + `","status":"scanned","flags":` + string(rune('0'+len(flags))) + `}`, nil
}

// mockRunScanFailure simulates the Executor failing (e.g. invalid permissions, crashed binary)
func mockRunScanFailure(ctx context.Context, target string, flags ...string) (string, error) {
	return "", errors.New("simulated native engine failure")
}

func TestStatusHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/status", nil)
	rr := httptest.NewRecorder()
	http.HandlerFunc(StatusHandler).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected 200 OK, got %v", status)
	}
	if rr.Body.String() != `{"status":"ok"}` {
		t.Errorf("unexpected body: %v", rr.Body.String())
	}
}

func TestScanHandlerAndAuth(t *testing.T) {
	validToken := "super-secret-key"
	// Wrap our handler with the middleware identically to production
	protectedHandler := AuthMiddleware(validToken, ScanHandler)

	tests := []struct {
		name           string
		method         string
		authHeader     string
		body           string
		mockScan       func(context.Context, string, ...string) (string, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Security Edge Case - Missing Auth Header",
			method:         "POST",
			authHeader:     "",
			body:           `{"target":"192.168.1.0/24"}`,
			mockScan:       mockRunScanSuccess,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized"}`,
		},
		{
			name:           "Security Edge Case - Invalid Auth Token",
			method:         "POST",
			authHeader:     "Bearer hacked-token",
			body:           `{"target":"192.168.1.0/24"}`,
			mockScan:       mockRunScanSuccess,
			expectedStatus: http.StatusForbidden,
			expectedBody:   `{"error":"forbidden"}`,
		},
		{
			name:           "Edge Case - Method Not Allowed (GET instead of POST)",
			method:         "GET",
			authHeader:     "Bearer " + validToken,
			body:           ``,
			mockScan:       mockRunScanSuccess,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   `{"error":"method not allowed"}`,
		},
		{
			name:           "Edge Case - Malformed JSON Payload",
			method:         "POST",
			authHeader:     "Bearer " + validToken,
			body:           `{"target":"192.168.1.0/24", I_AM_BROKEN}`, // Missing quotes
			mockScan:       mockRunScanSuccess,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"malformed JSON payload"}`,
		},
		{
			name:           "Edge Case - Missing Required Target Field",
			method:         "POST",
			authHeader:     "Bearer " + validToken,
			body:           `{"flags":["--json"]}`,
			mockScan:       mockRunScanSuccess,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"missing required field: target"}`,
		},
		{
			name:           "Real-World - Native Engine Execution Failure",
			method:         "POST",
			authHeader:     "Bearer " + validToken,
			body:           `{"target":"10.0.0.1"}`,
			mockScan:       mockRunScanFailure,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"simulated native engine failure"}`,
		},
		{
			name:           "Sunny Day - Successful Authenticated Scan Execution",
			method:         "POST",
			authHeader:     "Bearer " + validToken,
			body:           `{"target":"192.168.1.0/24", "flags":["--json"]}`,
			mockScan:       mockRunScanSuccess,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"network":"192.168.1.0/24","status":"scanned","flags":1}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Hook our mock executor specifically for this test's lifecycle
			originalRunScan := RunScan
			RunScan = tt.mockScan
			defer func() { RunScan = originalRunScan }()

			req, _ := http.NewRequest(tt.method, "/api/scan", bytes.NewBufferString(tt.body))
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			protectedHandler.ServeHTTP(rr, req)

			// Assert HTTP Status Compliance
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, status)
			}

			// Assert JSON Payload Structure
			if strings.TrimSpace(rr.Body.String()) != tt.expectedBody {
				t.Errorf("expected body '%v', got '%v'", tt.expectedBody, strings.TrimSpace(rr.Body.String()))
			}
		})
	}
}
