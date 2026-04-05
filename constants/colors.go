package constants

// ANSI text formatting and color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
	Blink     = "\033[5m"
	Reverse   = "\033[7m"
	Hidden    = "\033[8m"

	// Standard Colors
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// High Intensity / Bright Colors
	Gray          = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	// Background Colors
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"

	// Extended 256-Color Palette (Readable on dark backgrounds)
	Orange     = "\033[38;5;208m"
	DeepOrange = "\033[38;5;166m"
	HotPink    = "\033[38;5;205m"
	DeepPink   = "\033[38;5;161m"
	Mauve      = "\033[38;5;140m"
	Lavender   = "\033[38;5;141m"
	DeepYellow = "\033[38;5;214m"
	Silver     = "\033[38;5;145m"
	Gold       = "\033[38;5;220m"
	Copper     = "\033[38;5;172m"
	Bronze     = "\033[38;5;130m"
	Mint       = "\033[38;5;85m"
	SlateBlue  = "\033[38;5;67m"
	DeepBlue   = "\033[38;5;27m"
	Emerald    = "\033[38;5;41m"
	Crimson    = "\033[38;5;160m"
	SoftPurple = "\033[38;5;134m"
)
