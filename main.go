package main

import (
	"github.com/spf13/cobra"
	"github.com/wgoodall01/beatsaber-patcher/cmd"
)

func main() {
	// Start cobra
	cmd.Execute()
}

func init() {
	// Disable mousetrap.
	// This lets you run the binary on Windows by double-clicking
	// it from Explorer.
	cobra.MousetrapHelpText = ""
}
