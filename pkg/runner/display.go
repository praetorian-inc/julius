package runner

import (
	"fmt"
	"os"
)

// ANSI Color Constants
const (
	ColorRed   = "\033[31m"
	ColorBold  = "\033[1m"
	ColorReset = "\033[0m"
)

const banner = `
 _____           ___
/\___ \         /\_ \    __
\/__/\ \  __  __\//\ \  /\_\  __  __    ____
   _\ \ \/\ \/\ \ \ \ \ \/\ \/\ \/\ \  /',__\
  /\ \_\ \ \ \_\ \ \_\ \_\ \ \ \ \_\ \/\__, ` + "`" + `\
  \ \____/\ \____/ /\____\\ \_\ \____/\/\____/
   \/___/  \/___/  \/____/ \/_/\/___/  \/___/
 
 Vide omnia
 Praetorian Security, Inc.
`

func printBanner(useColor bool) {
	if useColor {
		fmt.Printf("%s%s%s%s\n", ColorBold, ColorRed, banner, ColorReset)
	} else {
		fmt.Printf("%s\n", banner)
	}
}

// isTerminal returns true if stdout is connected to a terminal.
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// isColorEnabled returns true if colored output should be used.
func isColorEnabled(noColor bool) bool {
	return !noColor && isTerminal()
}
