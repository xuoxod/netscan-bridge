import re

with open("main.go", "r") as f:
    orig_code = f.read()

# 1. Insert Functions
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
"""

code = orig_code.replace("func main() {", funcs + "\nfunc main() {")

# 2. Replace the message handling loop with the robust one from ExtractLatestSignalingState
old_loop = """for _, m := range msgs {
if m.Seq > lastSeq {
lastSeq = m.Seq
}
if m.SessionID == peerID {
continue // My own message
}

if m.Type == "bye" {
log.Println("Received 'bye' from remote peer. Tearing down.")
cancelActiveCommand()
continue
}

if m.Type == "offer" {
log.Println("Received offer from manager")
sdpRaw, _ := json.Marshal(m.Data["sdp"])
var sdp webrtc.SessionDescription
if err := json.Unmarshal(sdpRaw, &sdp); err == nil {
if err := pc.SetRemoteDescription(sdp); err != nil {
log.Printf("Failed setting remote desc: %v", err)
continue
}

ans, err := pc.CreateAnswer(nil)
if err != nil {
log.Printf("Failed creating answer: %v", err)
continue
}
if err := pc.SetLocalDescription(ans); err != nil {
log.Printf("Failed setting local desc: %v", err)
continue
}

sdpOut, _ := json.Marshal(ans)
var outMap map[string]interface{}
json.Unmarshal(sdpOut, &outMap)
sendSignal("answer", map[string]interface{}{"sdp": outMap})
}
} else if m.Type == "candidate" {
candRaw, _ := json.Marshal(m.Data["candidate"])
var cand webrtc.ICECandidateInit
if err := json.Unmarshal(candRaw, &cand); err == nil {
pc.AddICECandidate(cand)
}
}
}"""

new_loop = """// Extract state properly to avoid state machine panic
for _, m := range msgs {
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
log.Printf("DEBUG: Answer completed and sent!")
}
}
} else {
log.Printf("Skipping offer, current signaling state: %s", pc.SignalingState())
}
} else {
log.Printf("ERROR: Failed to unmarshal manager SDP offer: %v\\nsdpRaw: %s", err, string(sdpRaw))
}
}

for _, m := range candidates {
// We only want to process candidates for the current offer
// In a linear log, the `ExtractLatestSignalingState` will already have filtered them
candRaw, err := json.Marshal(m.Data["candidate"])
if err != nil { continue }

cand, err := ParseFrontendCandidate(candRaw)
if err != nil {
log.Printf("Failed parse frontend candidate: %v", err)
continue
}

if err := pc.AddICECandidate(cand); err != nil {
log.Printf("Failed to add ICE candidate: %v", err)
} else {
log.Printf("Added ICE candidate: %v", cand.Candidate)
}
}"""

# Reformat the regex slightly because spacing might be different
code = re.sub(re.escape(old_loop).replace(r'\ ', r'\s*').replace(r'\t', r'\s*'), new_loop, code)

# ALSO add logging to pc.OnICECandidate to confirm what it's generating
ice_gen_old = """pc.OnICECandidate(func(c *webrtc.ICECandidate) {
if c == nil {
return
}
b, _ := json.Marshal(c.ToJSON())
var raw map[string]interface{}
json.Unmarshal(b, &raw)
sendSignal("candidate", map[string]interface{}{"candidate": raw})
})"""

ice_gen_new = """pc.OnICECandidate(func(c *webrtc.ICECandidate) {
if c == nil {
log.Printf("ICE Gathering Complete")
return
}
log.Printf("Generated local ICE Candidate: %v", c.String())
b, _ := json.Marshal(c.ToJSON())
var raw map[string]interface{}
json.Unmarshal(b, &raw)
sendSignal("candidate", map[string]interface{}{"candidate": raw})
})"""

code = code.replace(ice_gen_old, ice_gen_new)

with open("main.go", "w") as f:
    f.write(code)

