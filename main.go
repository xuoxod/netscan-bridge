package main

import (
"bytes"
"context"
"encoding/json"
"fmt"
"io"
"log"
"net/http"
"net/url"
"os"
"os/signal"
"sync"
"syscall"
"time"

"netscan_bridge/executor"

"github.com/pion/webrtc/v3"
"gopkg.in/yaml.v3"
)

type Config struct {
SignalingURL string `yaml:"SIGNALING_URL"`
RoomID       string `yaml:"ROOM_ID"`
Token        string `yaml:"TOKEN"`
}

type SignalMessage struct {
Seq       int                    `json:"seq,omitempty"`
Type      string                 `json:"type"`
SessionID string                 `json:"sessionId"`
Data      map[string]interface{} `json:"data"`
}

func main() {
var cfg Config

// Parse config.yaml if exists
if b, err := os.ReadFile("config.yaml"); err == nil {
if err := yaml.Unmarshal(b, &cfg); err != nil {
log.Fatalf("Failed to parse config.yaml: %v", err)
}
}

// Environment overrides
if envSig := os.Getenv("SIGNALING_URL"); envSig != "" {
cfg.SignalingURL = envSig
}
if envRoom := os.Getenv("ROOM_ID"); envRoom != "" {
cfg.RoomID = envRoom
}
if envToken := os.Getenv("TOKEN"); envToken != "" {
cfg.Token = envToken
}

if cfg.SignalingURL == "" || cfg.RoomID == "" || cfg.Token == "" {
log.Fatal("FATAL: SIGNALING_URL, ROOM_ID, and TOKEN are required.")
}

log.Printf("🚀 Starting Headless WebRTC Intelligence Bridge")
log.Printf("Connecting to %s as %s", cfg.SignalingURL, cfg.RoomID)

peerID := "bridge-" + fmt.Sprintf("%d", time.Now().Unix())

webrtcConfig := webrtc.Configuration{
ICEServers: []webrtc.ICEServer{
{
URLs: []string{"stun:stun.l.google.com:19302"},
},
},
}

pc, err := webrtc.NewPeerConnection(webrtcConfig)
if err != nil {
log.Fatalf("Failed to create PeerConnection: %v", err)
}

var activeCtx context.Context
var activeCancel context.CancelFunc
var mu sync.Mutex

cancelActiveCommand := func() {
mu.Lock()
defer mu.Unlock()
if activeCancel != nil {
log.Println("⚠️ Cancelling active OS execution due to disconnect...")
activeCancel()
activeCancel = nil
}
}

pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
log.Printf("[WebRTC] Connection State: %s", s.String())
if s == webrtc.PeerConnectionStateDisconnected || s == webrtc.PeerConnectionStateFailed || s == webrtc.PeerConnectionStateClosed {
cancelActiveCommand()
}
})

sendSignal := func(msgType string, data map[string]interface{}) {
data["role"] = "host"
msg := SignalMessage{
Type:      msgType,
SessionID: peerID,
Data:      data,
}
b, _ := json.Marshal(msg)
req, _ := http.NewRequest("POST", cfg.SignalingURL+"/"+url.PathEscape(cfg.RoomID), bytes.NewBuffer(b))
req.Header.Set("Authorization", "Bearer "+cfg.Token)
req.Header.Set("Content-Type", "application/json")
resp, err := http.DefaultClient.Do(req)
if err != nil {
log.Printf("[Signal] POST error: %v", err)
return
}
resp.Body.Close()
}

pc.OnICECandidate(func(c *webrtc.ICECandidate) {
if c == nil {
return
}
b, _ := json.Marshal(c.ToJSON())
var raw map[string]interface{}
json.Unmarshal(b, &raw)
sendSignal("candidate", map[string]interface{}{"candidate": raw})
})

pc.OnDataChannel(func(d *webrtc.DataChannel) {
log.Printf("[WebRTC] New DataChannel: %s", d.Label())
d.OnMessage(func(msg webrtc.DataChannelMessage) {
var cmd map[string]interface{}
if err := json.Unmarshal(msg.Data, &cmd); err != nil {
log.Printf("Failed to decode DC message: %v", err)
return
}
action, _ := cmd["type"].(string)
if action == "bye" {
log.Println("Received 'bye' over DataChannel. Tearing down.")
cancelActiveCommand()
return
}
            actionType, ok := cmd["action"].(string)
if !ok {
actionType, _ = cmd["type"].(string)
}
            
if actionType == "scan" || actionType == "discover" || actionType == "weirdpackets" {
target, _ := cmd["target"].(string)
scanType := "discover"
if actionType == "scan" || actionType == "weirdpackets" {
scanType = "scan"
}

mu.Lock()
if activeCancel != nil {
activeCancel()
}
activeCtx, activeCancel = context.WithCancel(context.Background())
currCtx := activeCtx
mu.Unlock()

go func() {
log.Printf("⚡ Executing %s against %s", scanType, target)
d.SendText(`{"event":"toast", "text":"Started ` + scanType + ` against ` + target + `"}`)

out, execErr := executor.ExecuteScan(currCtx, target, scanType)
if execErr != nil {
errMsg := fmt.Sprintf(`{"event":"toast", "text":"Execution failed: %v"}`, execErr)
d.SendText(errMsg)
return
}

// We send the JSON output as a "toast" for simplicity, or we can send it as scan_complete
// Wait, the mobile needs the json payload to render the UI? No, Action Studio just shows it.
log.Println("✅ Execution complete. Streaming output back...")
// For big payloads, string replacement can be heavy, but normally it fits in memory
resp := map[string]interface{}{
"event": "scan_complete",
"data":  out,
}
bResp, _ := json.Marshal(resp)
d.Send(bResp)
}()
}
})
})

// Start signaling polling loop
go func() {
lastSeq := 0
for {
urlStr := fmt.Sprintf("%s/%s?since=%d", cfg.SignalingURL, url.PathEscape(cfg.RoomID), lastSeq)
req, _ := http.NewRequest("GET", urlStr, nil)
req.Header.Set("Authorization", "Bearer "+cfg.Token)
resp, err := http.DefaultClient.Do(req)

if err != nil {
log.Printf("[Signal] GET error: %v", err)
time.Sleep(3 * time.Second)
continue
}

b, _ := io.ReadAll(resp.Body)
resp.Body.Close()

if resp.StatusCode != http.StatusOK {
log.Printf("[Signal] Non-200 status: %d - %s", resp.StatusCode, string(b))
time.Sleep(3 * time.Second)
continue
}

var msgs []SignalMessage
if err := json.Unmarshal(b, &msgs); err != nil {
time.Sleep(1 * time.Second)
continue
}

for _, m := range msgs {
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
}
time.Sleep(1500 * time.Millisecond)
}
}()

// Graceful shutdown
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
s := <-sigChan
log.Printf("Caught signal %v, shutting down...", s)
cancelActiveCommand()
pc.Close()
}
