// Package cmd implements the git-profile command-line interface.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgPath string

var rootCmd = &cobra.Command{
	Use:   "git-profile",
	Short: "Manage multiple git identity profiles",
	Long: `git-profile lets you switch between named git identities
(user.name, user.email, SSH key, GPG signing) with a single command.

Get started by adding your first profile:

  git-profile add --id work --name "Jane Dev" --email jane@company.com

Then apply it to the current repository:

  git-profile use work

Or set it as the global default:

  git-profile use work --global`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute runs the root command and exits on error.
func Execute(version, commit, date string) {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate(fmt.Sprintf(
		"git-profile %s (commit: %s, built: %s)\n", version, commit, date,
	))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "",
		"config file path (default: $XDG_CONFIG_HOME/git-profile/config.json)")
}

func RunArgs(args []string) error {
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}
