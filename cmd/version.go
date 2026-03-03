package cmd

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print detailed version information",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			// read version from the root command (set by Execute()).
			ver := rootCmd.Version
			if ver == "" || ver == "dev" {
				if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
					ver = info.Main.Version
				} else {
					ver = "dev"
				}
			}

			bold := lipgloss.NewStyle().Bold(true)
			dim := lipgloss.NewStyle().Faint(true)

			fmt.Printf("%s %s\n", bold.Render("git-profile"), ver)
			fmt.Printf("  %s %s\n", dim.Render("go:    "), runtime.Version())
			fmt.Printf("  %s %s/%s\n", dim.Render("os:    "), runtime.GOOS, runtime.GOARCH)
		},
	}
	rootCmd.AddCommand(cmd)
}
