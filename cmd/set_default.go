package cmd

import (
	"fmt"

	"github.com/hapiio/git-profile/internal/git"
	"github.com/hapiio/git-profile/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	var global bool

	cmd := &cobra.Command{
		Use:   "set-default <profile-id>",
		Short: "Set the default profile for this repo or globally",
		Long: `Stores the default profile ID in git config (gitprofile.default).
This default is used by 'git-profile ensure' and installed git hooks.`,
		Example: `  git-profile set-default work           # per-repo default
  git-profile set-default personal --global  # global default`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			mgr, err := newManager()
			if err != nil {
				return err
			}
			cfg, err := mgr.Load()
			if err != nil {
				return err
			}

			if _, ok := cfg.Profiles[id]; !ok {
				return fmt.Errorf("profile %q not found", id)
			}

			scope := "local"
			if global {
				scope = "global"
			} else if !git.IsRepo() {
				return fmt.Errorf("not inside a git repository; use --global to set a global default")
			}

			if err := git.SetConfig(scope, "gitprofile.default", id); err != nil {
				return err
			}

			ui.Successf("Set %q as %s default profile", id, scope)
			return nil
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "Set as the global default")
	rootCmd.AddCommand(cmd)
}
