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

	if scanType != "discover" && scanType != "scan" && scanType != "specter" && scanType != "audit" && scanType != "weirdpackets" {
		return "", errors.New("invalid scanType, must be 'discover', 'scan', 'specter', 'audit', or 'weirdpackets'")
	}

	netscanBin := os.Getenv("NETSCAN_BIN_PATH")
	if netscanBin == "" {
		// Attempt to locate the bundled netscan binary securely.
		// We explicitly do NOT fall back to the system PATH ("netscan") to prevent
		// executing stale, globally installed, or unauthorized binaries on production.
		exePath, err := os.Executable()
		if err != nil {
			return "", fmt.Errorf("failed to determine bridge executable path: %w", err)
		}

		exeDir := filepath.Dir(exePath)
		bundledPath := filepath.Join(exeDir, "bin", "netscan")
		altPath := filepath.Join(exeDir, "netscan")

		if _, err := os.Stat(bundledPath); err == nil {
			netscanBin = bundledPath
		} else if _, err := os.Stat(altPath); err == nil {
			netscanBin = altPath
		} else {
			return "", fmt.Errorf("FATAL: bundled 'netscan' binary not found at %s or %s (refusing to use system PATH)", bundledPath, altPath)
		}
	}

	// Create an isolated ephemeral directory for this specific scan execution
	tmpDir, err := os.MkdirTemp("", "netscan_execution_*")
	if err != nil {
		return "", fmt.Errorf("failed to provision isolated workspace: %w", err)
	}
	// Secure teardown guarantee
	defer os.RemoveAll(tmpDir)

	// Combine base discovery commands with artifact isolation flags
	var args []string
	switch scanType {
	case "scan":
		// Map "scan" action to the "recon" engine for single IP targeting
		args = []string{"recon", "--ip", target, "--pt-json", "--out-dir", tmpDir}
	case "specter":
		args = []string{"specter", "-t", target, "--out-dir", tmpDir}
	case "audit":
		logFile := filepath.Join(tmpDir, "audit.jsonl")
		args = []string{"audit", "-t", target, "--log-file", logFile}
	case "weirdpackets":
		// weirdpackets does not output JSON or accept --out-dir
		args = []string{"weirdpackets", "-t", target}
	case "discover":
		args = []string{"discover", "-t", target, "--json", "--out-dir", tmpDir}
	default:
		args = []string{scanType, "-t", target, "--out-dir", tmpDir}
	}
	args = append(args, flags...)

	// Force line-buffering on Linux using stdbuf to prevent Rust fully buffering stdout when piped
	stdbufArgs := []string{"-oL", netscanBin}
	stdbufArgs = append(stdbufArgs, args...)
	cmd := execCommandContext(ctx, "stdbuf", stdbufArgs...)

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
			fmt.Println(line)
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
			fmt.Println(line)
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
		// weirdpackets ends successfully if user aborts or hits limit, or can return errors.
		// Let's just log it.
		if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("scan aborted by user or timeout")
		}
		return "", fmt.Errorf("scan failed: %w, stderr: %s", err, stderrBuf.String())
	}

	// Locate the expected JSON artifact
	var fileContent []byte
	var readErr error

	if scanType == "weirdpackets" {
		// Weirdpackets produces no artifact, stdout/stderr is the output
		if stdoutBuf.Len() > 0 {
			return stdoutBuf.String(), nil
		}
		return "Weirdpackets run completed.", nil
	} else if scanType == "audit" {
		logFile := filepath.Join(tmpDir, "audit.jsonl")
		fileContent, readErr = os.ReadFile(logFile)
		if readErr != nil && stdoutBuf.Len() > 0 {
			return stdoutBuf.String(), nil
		} else if readErr == nil {
			return string(fileContent), nil
		}
		return "Audit run completed.", nil
	} else if scanType == "discover" {
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
		if stdoutBuf.Len() > 0 {
			return stdoutBuf.String(), nil
		}
		return "", fmt.Errorf("scan completed but failed to produce JSON artifact: %w", readErr)
	}

	return string(fileContent), nil
}
