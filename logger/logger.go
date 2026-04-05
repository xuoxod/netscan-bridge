package logger

import (
	"fmt"
	"time"

	"netscan_bridge/constants"
)

// timePrefix generates a dimmed timestamp for the log prefix snippet
func timePrefix() string {
	return fmt.Sprintf("%s%s%s", constants.Gray, time.Now().Format("15:04:05"), constants.Reset)
}

// Info prints a standard diagnostic message
func Info(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s %s[%s%s%s]%s %s\n", timePrefix(), constants.Bold, constants.DeepBlue, constants.EmoInfo, constants.Reset+constants.Bold, constants.Reset, msg)
}

// Success prints a successful operation message
func Success(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s %s[%s%s%s]%s %s%s%s\n", timePrefix(), constants.Bold, constants.Emerald, constants.EmoSuccess, constants.Reset+constants.Bold, constants.Reset, constants.Emerald, msg, constants.Reset)
}

// Warn prints a warning or interception message
func Warn(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s %s[%s%s%s]%s %s%s%s\n", timePrefix(), constants.Bold, constants.DeepYellow, constants.EmoWarn, constants.Reset+constants.Bold, constants.Reset, constants.DeepYellow, msg, constants.Reset)
}

// Error prints a hard failure message
func Error(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s %s[%s%s%s]%s %s%s%s\n", timePrefix(), constants.Bold, constants.Crimson, constants.EmoFail, constants.Reset+constants.Bold, constants.Reset, constants.Crimson, msg, constants.Reset)
}

// Stream prints stdout/live execution telemetry emitted from the rust engine
func Stream(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s %s[%s%s%s]%s %s%s%s\n", timePrefix(), constants.Bold, constants.SlateBlue, constants.EmoZap, constants.Reset+constants.Bold, constants.Reset, constants.Silver, msg, constants.Reset)
}

// System prints low-level P2P/WebRTC states
func System(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s %s[%s⚙%s]%s %s%s%s\n", timePrefix(), constants.Bold, constants.Mauve, constants.Reset+constants.Bold, constants.Reset, constants.Mauve, msg, constants.Reset)
}

// Recon designates scanning/intelligence gathering operations
func Recon(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s %s[%s%s%s]%s %s%s%s\n", timePrefix(), constants.Bold, constants.DeepOrange, constants.EmoSpy, constants.Reset+constants.Bold, constants.Reset, constants.DeepOrange, msg, constants.Reset)
}

// Exploit highlights offensive/active operations
func Exploit(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s %s[%s%s%s]%s %s%s%s\n", timePrefix(), constants.Bold, constants.DeepPink, constants.EmoTarget, constants.Reset+constants.Bold, constants.Reset, constants.HotPink, msg, constants.Reset)
}

// Defense signifies system hardening or protection states
func Defense(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s %s[%s%s%s]%s %s%s%s\n", timePrefix(), constants.Bold, constants.Mint, constants.EmoShield, constants.Reset+constants.Bold, constants.Reset, constants.Mint, msg, constants.Reset)
}

// Custom allows arbitrary emoji and color combinations for specialized alerts
func Custom(emoji, color, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Printf("%s %s[%s%s%s]%s %s%s%s\n", timePrefix(), constants.Bold, color, emoji, constants.Reset+constants.Bold, constants.Reset, color, msg, constants.Reset)
}

// KVList prints an aligned key-value list (safest alternative to rigid tables)
func KVList(title string, items map[string]string) {
	fmt.Printf("\n%s%s❖ %s%s\n", constants.Bold, constants.Gold, title, constants.Reset)
	for k, v := range items {
		fmt.Printf("  %s%s%s: %s\n", constants.Silver, k, constants.Reset, v)
	}
	fmt.Println()
}

// Block outputs multi-line data (like payload dumps) safely indented
func Block(title string, content string) {
	fmt.Printf("\n%s%s❖ %s%s\n", constants.Bold, constants.Copper, title, constants.Reset)
	fmt.Printf("%s%s%s\n\n", constants.Silver, content, constants.Reset)
}

// Banner prints the standalone execution banner cleanly
func Banner(text string) {
	fmt.Printf("\n%s%s%s%s\n", constants.HotPink, constants.Bold, text, constants.Reset)
}
