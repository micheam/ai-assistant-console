package theme

import "github.com/fatih/color"

var (
	// Basic colors
	white  = color.New(color.FgWhite).SprintFunc()
	gray   = color.New(color.FgHiBlack).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()

	// Text decorations
	Bold      = color.New(color.Bold).SprintFunc()
	Underline = color.New(color.Underline).SprintFunc()

	// Color themes
	Info  = gray
	Reply = blue
	Error = red
)
