package executor

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync"
)

var execCommandContext = exec.CommandContext

// strict regex defending against shell injection characters. Only CIDR, IP, and Hostnames allowed.
var validTargetRegex = regexp.MustCompile(`^[a-zA-Z0-9\.\-:/]+$`)

// ExecuteScan securely triggers the OS binary within an isolated temporary workspace,
// grabs the resulting JSON artifact, and tears down the workspace securely.
func ExecuteScan(ctx context.Context, target string, scanType string, onStdout func(string), onStderr func(string), flags ...string) (string, error) {
	// 1. Edge Case: Context Timeout check immediately
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("scan aborted early: %w", err)
	}

	// 2. Security: Deep Sanitization to avoid bash logic slipping through
	if !validTargetRegex.MatchString(target) {
		return "", errors.New("invalid target format: illegal characters detected")
	}

	if scanType != "discover" && scanType != "scan" && scanType != "specter" {
		return "", errors.New("invalid scanType, must be 'discover', 'scan', or 'specter'")
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
	args := []string{scanType, "-t", target}
	if scanType == "scan" {
		args = append(args, "--pt-json", "--out-dir", tmpDir)
	} else if scanType == "specter" {
		args = append(args, "--out-dir", tmpDir)
	} else {
		args = append(args, "--json", "--out-dir", tmpDir)
	}
	args = append(args, flags...)

	cmd := execCommandContext(ctx, netscanBin, args...)

	// Enable Graceful Abort if context is cancelled
	cmd.Cancel = func() error {
		return cmd.Process.Signal(os.Interrupt) // Use SIGINT instead of standard SIGKILL
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to attach stdout pipe: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to attach stderr pipe: %w", err)
	}

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(2)

	// Stream stdout line-by-line
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(io.TeeReader(stdoutPipe, &stdoutBuf))
		for scanner.Scan() {
			line := scanner.Text()
			if onStdout != nil {
				onStdout(line)
			}
		}
	}()

	// Stream stderr line-by-line
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(io.TeeReader(stderrPipe, &stderrBuf))
		for scanner.Scan() {
			line := scanner.Text()
			if onStderr != nil {
				onStderr(line)
			}
		}
	}()

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start execution: %w", err)
	}

	wg.Wait() // Ensure all pipes are consumed before waiting on completion

	if err := cmd.Wait(); err != nil {
		if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("scan aborted by user or timeout")
		}
		return "", fmt.Errorf("scan failed: %w, stderr: %s", err, stderrBuf.String())
	}

	// Locate the expected JSON artifact
	var fileContent []byte
	var readErr error

	if scanType == "discover" {
		jsonArtifactPath := filepath.Join(tmpDir, "discovered_hosts.json")
		fileContent, readErr = os.ReadFile(jsonArtifactPath)
	} else if scanType == "scan" || scanType == "specter" {
		// Find scan-*.json or specter_report_*.json
		matchPattern := "scan-*.json"
		if scanType == "specter" {
			matchPattern = "specter_report_*.json"
		}
		matches, err := filepath.Glob(filepath.Join(tmpDir, matchPattern))
		if err != nil || len(matches) == 0 {
			readErr = fmt.Errorf("could not find JSON artifact file matching pattern %s: %v", matchPattern, err)
		} else {
			fileContent, readErr = os.ReadFile(matches[0])
		}
	}

	if readErr != nil {
		// Fallback: If it didn't produce the JSON file, it might just be the raw text output.
		// For a purely API-driven bridge, missing the artifact is a systemic error,
		// but we'll return stdout as a resilient fallback in case a non-discovery command was triggered.
		if stdoutBuf.Len() > 0 {
			return stdoutBuf.String(), nil
		}
		return "", fmt.Errorf("scan completed but failed to produce JSON artifact: %w", readErr)
	}

	return string(fileContent), nil
}
