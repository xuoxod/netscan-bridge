package main

import (
	"bytes"
	"context"
	"encoding/base64"
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

func ExtractLatestSignalingState(msgs []SignalMessage, selfPeerID string) (lastOffer *SignalMessage, candidates []SignalMessage, reset bool) {
	for _, m := range msgs {
		if m.SessionID == selfPeerID {
			continue
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
			candidates = nil
			reset = false
		} else if m.Type == "candidate" {
			candidates = append(candidates, m)
		}
	}
	return lastOffer, candidates, reset
}

func main() {
	var cfg Config

	if b, err := os.ReadFile("config.yaml"); err == nil {
		if err := yaml.Unmarshal(b, &cfg); err != nil {
			log.Fatalf("Failed to parse config.yaml: %v", err)
		}
	}

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
			activeCancel()
			activeCancel = nil
		}
	}

	pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		log.Printf("[WebRTC] Connection State: %s", s.String())
		if s == webrtc.PeerConnectionStateFailed || s == webrtc.PeerConnectionStateClosed {
			log.Println("Connection failed/closed. Exiting to reboot bridge process...")
			os.Exit(0)
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
			log.Println("DEBUG: Local ICE Gathering Completed")
			return
		}
		log.Printf("Generated Local ICE Candidate: %v", c.String())
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
				os.Exit(0)
				return
			}
			actionType, ok := cmd["action"].(string)
			if !ok {
				actionType, _ = cmd["type"].(string)
			}
			if actionType == "scan" || actionType == "discover" || actionType == "weirdpackets" || actionType == "specter" || actionType == "audit" {
				target, _ := cmd["target"].(string)

				var customFlags []string
				if rawFlags, ok := cmd["flags"].([]interface{}); ok {
					for _, f := range rawFlags {
						if fStr, ok := f.(string); ok {
							customFlags = append(customFlags, fStr)
						}
					}
				}

				scanType := actionType
				if actionType == "discover" {
					scanType = "discover"
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
					onStdout := func(line string) {
						safeStr, _ := json.Marshal(line)
						d.SendText(`{"event":"stream", "channel":"stdout", "text":` + string(safeStr) + `}`)
					}

					onStderr := func(line string) {
						safeStr, _ := json.Marshal(line)
						d.SendText(`{"event":"stream", "channel":"stderr", "text":` + string(safeStr) + `}`)
					}

					out, execErr := executor.ExecuteScan(currCtx, target, scanType, onStdout, onStderr, customFlags...)
					if execErr != nil {
						errMsg := fmt.Sprintf(`{"event":"toast", "text":"Execution failed: %v"}`, execErr)
						d.SendText(errMsg)
						return
					}

					log.Printf("✅ Execution complete. Streaming output back... Payload size: %d", len(out))
					resp := map[string]interface{}{
						"event": "scan_complete",
						"type":  scanType,
						"data":  out,
					}
					b, _ := json.Marshal(resp)
					bSize := len(b)
					if bSize > 16384 {
						chunkSize := 16384
						totalChunks := (bSize + chunkSize - 1) / chunkSize
						id := fmt.Sprintf("%d", time.Now().UnixNano())
						for i := 0; i < totalChunks; i++ {
							end := (i + 1) * chunkSize
							if end > bSize {
								end = bSize
							}
							chunkMsg := map[string]interface{}{
								"event": "chunk",
								"id":    id,
								"index": i,
								"total": totalChunks,
								"data":  base64.StdEncoding.EncodeToString(b[i*chunkSize : end]),
							}
							cb, _ := json.Marshal(chunkMsg)
							d.SendText(string(cb))
							time.Sleep(5 * time.Millisecond)
						}
					} else {
						d.SendText(string(b))
					}
				}()
			} else if actionType == "abort" {
				log.Println("Received ABORT command from UI.")
				cancelActiveCommand()
				d.SendText(`{"event":"toast", "text":"Remote execution forcefully aborted."}`)
			} else {
				log.Printf("Unknown command received: %s", actionType)
			}
		})
	})

	go func() {
		lastSeq := 0
		for {
			urlStr := fmt.Sprintf("%s/%s?since=%d", cfg.SignalingURL, url.PathEscape(cfg.RoomID), lastSeq)
			req, _ := http.NewRequest("GET", urlStr, nil)
			req.Header.Set("Authorization", "Bearer "+cfg.Token)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("[Signal] POLL error: %v", err)
				time.Sleep(3 * time.Second)
				continue
			}

			b, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				log.Printf("[Signal] Body read error: %v", err)
				time.Sleep(3 * time.Second)
				continue
			}

			if resp.StatusCode != http.StatusOK {
				log.Printf("[Signal] Non-200 status: %d - %s", resp.StatusCode, string(b))
				time.Sleep(3 * time.Second)
				continue
			}

			var respObj struct {
				Messages []SignalMessage `json:"messages"`
				MaxSeq   int             `json:"maxSeq"`
			}

			var msgs []SignalMessage
			err = json.Unmarshal(b, &respObj)
			if err == nil {
				msgs = respObj.Messages
				if respObj.MaxSeq > lastSeq {
					lastSeq = respObj.MaxSeq
				}
			} else {
				_ = json.Unmarshal(b, &msgs)
			}

			for _, m := range msgs {
				if m.Seq > lastSeq {
					lastSeq = m.Seq
				}
			}

			lastOffer, candidates, reset := ExtractLatestSignalingState(msgs, peerID)

			if reset {
				log.Println("Received 'bye' from remote peer. Tearing down.")
				cancelActiveCommand()
				os.Exit(0)
			}

			if lastOffer != nil {
				log.Println("Received offer from manager")
				sdpRaw, _ := json.Marshal(lastOffer.Data["sdp"])
				var sdp webrtc.SessionDescription
				if err := json.Unmarshal(sdpRaw, &sdp); err == nil {
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
					}
				} else {
					log.Printf("ERROR: Failed parsing SDP raw: %s | err: %v", string(sdpRaw), err)
				}
			}

			for _, m := range candidates {
				cRaw, _ := json.Marshal(m.Data["candidate"])
				cand, err := ParseFrontendCandidate(cRaw)
				if err != nil {
					continue
				}
				if err := pc.AddICECandidate(cand); err == nil {
					log.Printf("Added ICE candidate: %v", cand.Candidate)
				}
			}

			time.Sleep(1500 * time.Millisecond)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	s := <-sigChan
	log.Printf("Caught signal %v, shutting down...", s)
	cancelActiveCommand()
	os.Exit(0)
	pc.Close()
}
