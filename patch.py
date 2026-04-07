import sys

with open("main.go", "r") as f:
    orig = f.read()

funcs = """
// ParseFrontendCandidate safely unmarshals an ICE Candidate JSON payload originating from a JS frontend
// This addresses strict typing mismatches, such as `sdpMid` transmitting strictly as a number instead of a string
func ParseFrontendCandidate(raw []byte) (webrtc.ICECandidateInit, error) {
var cndMap map[string]interface{}
if err := json.Unmarshal(raw, &cndMap); err != nil {
return webrtc.ICECandidateInit{}, fmt.Errorf("failed raw parse: %w", err)
}

if mid, exists := cndMap["sdpMid"]; exists && mid != nil {
if midNum, isNum := mid.(float64); isNum {
cndMap["sdpMid"] = fmt.Sprintf("%.0f", midNum)
}
}

sanitizedRaw, err := json.Marshal(cndMap)
if err != nil {
return webrtc.ICECandidateInit{}, fmt.Errorf("failed remashal: %w", err)
}

var cand webrtc.ICECandidateInit
if err := json.Unmarshal(sanitizedRaw, &cand); err != nil {
return webrtc.ICECandidateInit{}, fmt.Errorf("failed cast to ICInit: %w", err)
}

return cand, nil
}

// ExtractLatestSignalingState interprets a linear signaling history, returning the most recent active offer and its bundled candidates.
// Returns `reset=true` if the terminal state was a `bye` requiring active peer connection teardown.
func ExtractLatestSignalingState(msgs []SignalMessage, selfPeerID string) (lastOffer *SignalMessage, candidates []SignalMessage, reset bool) {
for _, m := range msgs {
if m.SessionID == selfPeerID {
continue // My own message
}

if m.Type == "bye" {
reset = true
lastOffer = nil
candidates = nil
continue
}

if m.Type == "offer" {
offerCopy := m
lastOffer = &offerCopy
candidates = nil // invalidate older candidates
reset = false
} else if m.Type == "candidate" {
candidates = append(candidates, m)
}
}
return lastOffer, candidates, reset
}

func main() {"""

orig = orig.replace("func main() {", funcs)

old_loop_start = orig.find('for _, m := range msgs {')
old_loop_end = orig.find('time.Sleep(1500 * time.Millisecond)', old_loop_start)

if old_loop_start != -1 and old_loop_end != -1:
    old_loop = orig[old_loop_start:old_loop_end]
    new_loop = """for _, m := range msgs {
if m.Seq > lastSeq {
lastSeq = m.Seq
}
}

lastOffer, candidates, reset := ExtractLatestSignalingState(msgs, peerID)

if reset {
log.Println("Received 'bye' from remote peer. Tearing down.")
cancelActiveCommand()
}

if lastOffer != nil {
log.Println("Received offer from manager")
sdpRaw, _ := json.Marshal(lastOffer.Data["sdp"])
var sdp webrtc.SessionDescription
if err := json.Unmarshal(sdpRaw, &sdp); err == nil {
// Check state to prevent have-remote-offer transitions if already processing or connected
if pc.SignalingState() == webrtc.SignalingStateStable || pc.SignalingState() == webrtc.SignalingStateHaveLocalOffer {
if err := pc.SetRemoteDescription(sdp); err != nil {
log.Printf("Failed setting remote desc: %v", err)
} else {
ans, err := pc.CreateAnswer(nil)
if err != nil {
log.Printf("Failed creating answer: %v", err)
} else if err := pc.SetLocalDescription(ans); err != nil {
log.Printf("Failed setting local desc: %v", err)
} else {
sdpOut, _ := json.Marshal(ans)
var outMap map[string]interface{}
json.Unmarshal(sdpOut, &outMap)
sendSignal("answer", map[string]interface{}{"sdp": outMap})
log.Printf("DEBUG: Answer sent successfully!")
}
}
} else {
log.Printf("Skipping offer, current signaling state: %s", pc.SignalingState())
}
} else {
log.Printf("ERROR: Failed to parse SDP offer: %v\\nRaw payload: %s", err, string(sdpRaw))
}
}

for _, m := range candidates {
cRaw, _ := json.Marshal(m.Data["candidate"])
cand, err := ParseFrontendCandidate(cRaw)
if err != nil {
continue
}
if err := pc.AddICECandidate(cand); err == nil {
log.Printf("DEBUG: Added ICE candidate: %v", cand.Candidate)
}
}
"""
    orig = orig.replace(old_loop, new_loop)

# Patch the ICE Candidate generation
ice_old = """pc.OnICECandidate(func(c *webrtc.ICECandidate) {
if c == nil {
return
}
b, _ := json.Marshal(c.ToJSON())
var raw map[string]interface{}
json.Unmarshal(b, &raw)
sendSignal("candidate", map[string]interface{}{"candidate": raw})
})"""

ice_new = """pc.OnICECandidate(func(c *webrtc.ICECandidate) {
if c == nil {
log.Println("DEBUG: Local ICE Gathering Completed")
return
}
log.Printf("DEBUG: Generated Local ICE Candidate: %v", c.String())
b, _ := json.Marshal(c.ToJSON())
var raw map[string]interface{}
json.Unmarshal(b, &raw)
sendSignal("candidate", map[string]interface{}{"candidate": raw})
})"""

orig = orig.replace(ice_old, ice_new)

with open("main.go", "w") as f:
    f.write(orig)

