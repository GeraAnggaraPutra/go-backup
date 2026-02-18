package constant

// ANSI escape codes for colors and terminal effects
const (
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[90m"
	ColorBold   = "\033[1m"
	ColorReset  = "\033[0m"

	// ANSI codes for clearing lines and moving cursor
	ClearLine = "\033[2K"
	CursorUp  = "\033[%dA"
)

// Icons for different statuses
const (
	IconSuccess = "✔"
	IconError   = "✘"
	IconRunning = "●"
)
