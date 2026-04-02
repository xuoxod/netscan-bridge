package executor

import (
"bytes"
"context"
"errors"
"fmt"
"os/exec"
"regexp"
)

var execCommandContext = exec.CommandContext

// strict regex defending against shell injection characters. Only CIDR, IP, and Hostnames allowed.
var validTargetRegex = regexp.MustCompile(`^[a-zA-Z0-9\.\-:/]+$`)

// ExecuteScan securely triggers the OS binary with sanitized inputs.
func ExecuteScan(ctx context.Context, target string, flags ...string) (string, error) {
// 1. Edge Case: Context Timeout check immediately
if err := ctx.Err(); err != nil {
return "", fmt.Errorf("scan aborted early: %w", err)
}

// 2. Security: Deep Sanitization to avoid bash logic slipping through 
if !validTargetRegex.MatchString(target) {
return "", errors.New("invalid target format: illegal characters detected")
}

// 3. Execution: Native isolation parameter array
args := []string{"discover", "-t", target}
args = append(args, flags...)

cmd := execCommandContext(ctx, "netscan", args...)

var stdout bytes.Buffer
var stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr

if err := cmd.Run(); err != nil {
if errors.Is(ctx.Err(), context.DeadlineExceeded) {
return "", fmt.Errorf("context deadline exceeded")
}
// Return specific error along with any standard error buffer from the invoked shell program 
return "", fmt.Errorf("scan failed: %w, stderr: %s", err, stderr.String())
}

return stdout.String(), nil
}
