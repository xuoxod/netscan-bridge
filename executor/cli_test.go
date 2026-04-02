package executor

import (
"context"
"fmt"
"os"
"os/exec"
"strings"
"testing"
"time"
)

// mockExecCommandContext allows us to intercept os/exec calls and mock the CLI program.
// We redirect the execution back towards this test binary using a specific env variable.
func mockExecCommandContext(ctx context.Context, command string, args ...string) *exec.Cmd {
cs := []string{"-test.run=TestHelperProcess", "--", command}
cs = append(cs, args...)
cmd := exec.CommandContext(ctx, os.Args[0], cs...)
cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
return cmd
}

// TestHelperProcess acts as our mocked "netscan" binary. It reads the arguments and fakes output.
func TestHelperProcess(t *testing.T) {
if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
return
}
args := os.Args
for len(args) > 0 {
if args[0] == "--" {
args = args[1:]
break
}
args = args[1:]
}
if len(args) == 0 {
os.Exit(1)
}

cmd := args[0]
if cmd == "netscan" {
// Simulate execution failure edge case
for _, arg := range args[1:] {
if arg == "fail-me" {
fmt.Fprint(os.Stderr, "simulate error\n")
os.Exit(2)
}
}

// Simulate successful run
if len(args) >= 3 && args[1] == "discover" && args[2] == "-t" {
target := args[3]
fmt.Fprintf(os.Stdout, `{"network":"%s","status":"scanned"}`, target)
os.Exit(0)
}
}
os.Exit(1)
}

func TestExecuteScan(t *testing.T) {
// Arrange: override execCommandContext for our tests
originalExec := execCommandContext
execCommandContext = mockExecCommandContext
defer func() { execCommandContext = originalExec }()

tests := []struct {
name        string
target      string
flags       []string
timeout     time.Duration
shouldErr   bool
errContains string
wantOutput  string
}{
{
name:       "Sunny Day - Valid Target",
target:     "192.168.1.0/24",
flags:      []string{"--json"},
timeout:    time.Second,
shouldErr:  false,
wantOutput: `{"network":"192.168.1.0/24","status":"scanned"}`,
},
{
name:        "Edge Case - Process Returns Non-Zero Exit Code",
target:      "fail-me",
flags:       nil,
timeout:     time.Second,
shouldErr:   true,
errContains: "exit status 2",
},
{
name:        "Security - Rejects Bash Injection Tactics",
target:      "192.168.1.1; rm -rf /",
flags:       nil,
timeout:     time.Second,
shouldErr:   true,
errContains: "invalid target format",
},
{
name:        "Edge Case - Context Timeout Exceeded",
target:      "10.0.0.0/8",
flags:       nil,
timeout:     0, // instantly timeout
shouldErr:   true,
errContains: "context deadline exceeded",
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
// Act
ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
defer cancel()

// Introduce slight delay if testing timeout manually so exact edge hits
if tt.timeout == 0 {
    // This just simulates something that definitely times out
    ctx, cancel = context.WithTimeout(context.Background(), 1*time.Millisecond)
    defer cancel()
    time.Sleep(5 * time.Millisecond)
}

out, err := ExecuteScan(ctx, tt.target, tt.flags...)

// Assert
if tt.shouldErr {
if err == nil {
t.Fatalf("Expected an error but got none. Output: %v", out)
}
if !strings.Contains(err.Error(), tt.errContains) {
t.Errorf("Expected error to contain %q, but got: %v", tt.errContains, err)
}
} else {
if err != nil {
t.Fatalf("Unexpected error: %v", err)
}
if out != tt.wantOutput {
t.Errorf("Output mismatch. Want %q, got %q", tt.wantOutput, out)
}
}
})
}
}
