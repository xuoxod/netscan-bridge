package executor

import (
"bytes"
"context"
"errors"
"fmt"
"os"
"os/exec"
"path/filepath"
"regexp"
)

var execCommandContext = exec.CommandContext

// strict regex defending against shell injection characters. Only CIDR, IP, and Hostnames allowed.
var validTargetRegex = regexp.MustCompile(`^[a-zA-Z0-9\.\-:/]+$`)

// ExecuteScan securely triggers the OS binary within an isolated temporary workspace, 
// grabs the resulting JSON artifact, and tears down the workspace securely.
func ExecuteScan(ctx context.Context, target string, flags ...string) (string, error) {
// 1. Edge Case: Context Timeout check immediately
if err := ctx.Err(); err != nil {
return "", fmt.Errorf("scan aborted early: %w", err)
}

// 2. Security: Deep Sanitization to avoid bash logic slipping through
if !validTargetRegex.MatchString(target) {
return "", errors.New("invalid target format: illegal characters detected")
}

netscanBin := os.Getenv("NETSCAN_BIN_PATH")
if netscanBin == "" {
netscanBin = "netscan"
}

// Create an isolated ephemeral directory for this specific scan execution
tmpDir, err := os.MkdirTemp("", "netscan_execution_*")
if err != nil {
return "", fmt.Errorf("failed to provision isolated workspace: %w", err)
}
// Secure teardown guarantee
defer os.RemoveAll(tmpDir)

// Combine base discovery commands with artifact isolation flags
args := []string{"discover", "-t", target, "--json", "--out-dir", tmpDir}
args = append(args, flags...)

cmd := execCommandContext(ctx, netscanBin, args...)

var stdout bytes.Buffer
var stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr

if err := cmd.Run(); err != nil {
if errors.Is(ctx.Err(), context.DeadlineExceeded) {
return "", fmt.Errorf("context deadline exceeded")
}
return "", fmt.Errorf("scan failed: %w, stderr: %s", err, stderr.String())
}

// Locate the expected JSON artifact
jsonArtifactPath := filepath.Join(tmpDir, "discovered_hosts.json")
fileContent, err := os.ReadFile(jsonArtifactPath)
if err != nil {
// Fallback: If it didn't produce the JSON file, it might just be the raw text output.
// For a purely API-driven bridge, missing the artifact is a systemic error, 
// but we'll return stdout as a resilient fallback in case a non-discovery command was triggered.
if stdout.Len() > 0 {
return stdout.String(), nil
}
return "", fmt.Errorf("scan completed but failed to produce JSON artifact: %w", err)
}

return string(fileContent), nil
}
