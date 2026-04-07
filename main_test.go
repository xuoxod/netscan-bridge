package main

import (
	"encoding/json"
	"testing"
)

func TestParseFrontendCandidate(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		wantErr     bool
		expectedMid string
	}{
		{
			name:        "Logged Mobile Frontend Payload (Numeric MID)",
			input:       `{"candidate":"candidate:50539522 1 udp 2122063615 192.168.1.163 59581 typ host generation 0 ufrag FdNu network-id 3 network-cost 10","sdpMLineIndex":0,"sdpMid":0,"usernameFragment":"FdNu"}`,
			wantErr:     false,
			expectedMid: "0",
		},
		{
			name:        "Standard Compliant (String MID)",
			input:       `{"candidate":"candidate:12345 1 udp 2122063615 10.0.0.1 59581 typ host","sdpMLineIndex":0,"sdpMid":"1"}`,
			wantErr:     false,
			expectedMid: "1",
		},
		{
			name:        "Missing MID entirely",
			input:       `{"candidate":"candidate:67890 1 udp 2122063615 10.0.0.2 59581 typ host","sdpMLineIndex":0}`,
			wantErr:     false,
			expectedMid: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cand, err := ParseFrontendCandidate([]byte(tc.input))
			if (err != nil) != tc.wantErr {
				t.Fatalf("ParseFrontendCandidate() returned err %v, expected err %v", err, tc.wantErr)
			}
			if !tc.wantErr {
				if tc.expectedMid == "" && cand.SDPMid != nil {
					t.Fatalf("Expected nil SDPMid, but got %v", *cand.SDPMid)
				} else if tc.expectedMid != "" {
					if cand.SDPMid == nil {
						t.Fatalf("Expected SDPMid to be '%v', but got nil", tc.expectedMid)
					}
					if *cand.SDPMid != tc.expectedMid {
						t.Fatalf("Expected SDPMid to be '%v', but got '%v'", tc.expectedMid, *cand.SDPMid)
					}
				}
			}
		})
	}
}

func TestExtractLatestSignalingState(t *testing.T) {
	historyJSON := `[
		{"seq":101, "sessionId":"mobile-123", "type":"offer", "data":{"sdp":{"type":"offer","sdp":"old_sdp"}}},
		{"seq":102, "sessionId":"mobile-123", "type":"candidate", "data":{"candidate":{"candidate":"old_candidate","sdpMid":0}}},
		{"seq":103, "sessionId":"mobile-123", "type":"bye", "data":{}},
		{"seq":104, "sessionId":"mobile-456", "type":"offer", "data":{"sdp":{"type":"offer","sdp":"new_sdp"}}},
		{"seq":105, "sessionId":"mobile-456", "type":"candidate", "data":{"candidate":{"candidate":"new_candidate_1","sdpMid":0,"sdpMLineIndex":0}}},
		{"seq":106, "sessionId":"mobile-456", "type":"candidate", "data":{"candidate":{"candidate":"new_candidate_2","sdpMid":0,"sdpMLineIndex":0}}}
	]`

	var msgs []SignalMessage
	if err := json.Unmarshal([]byte(historyJSON), &msgs); err != nil {
		t.Fatalf("Failed to parse simulated history JSON: %v", err)
	}

	lastOffer, candidates, needsReset := ExtractLatestSignalingState(msgs, "my-bridge")

	if needsReset {
		t.Errorf("Expected reset to be false since it ended in a stable offer/candidate state, got true")
	}

	if lastOffer == nil {
		t.Fatal("Expected a valid offer struct, got nil")
	}

	sdpStr, _ := json.Marshal(lastOffer.Data["sdp"])
	if string(sdpStr) != `{"sdp":"new_sdp","type":"offer"}` {
		t.Errorf("Expected exact new offer SDP, got %s", string(sdpStr))
	}

	if len(candidates) != 2 {
		t.Fatalf("Expected exactly 2 valid candidates associated with newest offer, got %d", len(candidates))
	}

	candMap, _ := candidates[0].Data["candidate"].(map[string]interface{})
	if candMap["candidate"] != "new_candidate_1" {
		t.Errorf("First candidate parsed incorrectly: %v", candMap)
	}

	// Test Tear-down Scenario
	terminalJSON := `[
		{"seq":201, "sessionId":"mobile-456", "type":"offer", "data":{"sdp":{"type":"offer","sdp":"active_sdp"}}},
		{"seq":202, "sessionId":"mobile-456", "type":"bye", "data":{}}
	]`
	var teardownMsgs []SignalMessage
	json.Unmarshal([]byte(terminalJSON), &teardownMsgs)

	tOffer, tCands, terminalReset := ExtractLatestSignalingState(teardownMsgs, "my-bridge")
	if !terminalReset {
		t.Error("Expected terminal 'bye' to set reset=true")
	}
	if tOffer != nil || len(tCands) > 0 {
		t.Error("Expected terminal logic to scrub pending offers and candidates to avoid stale associations")
	}
}
