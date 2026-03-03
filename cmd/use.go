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
		Use:   "use <profile-id>",
		Short: "Apply a profile to this repo (or globally with --global)",
		Example: `  git-profile use work
  git-profile use personal --global`,
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

			p, ok := cfg.Profiles[id]
			if !ok {
				return fmt.Errorf("profile %q not found; run 'git-profile list' to see available profiles", id)
			}

			scope := "local"
			if global {
				scope = "global"
			} else if !git.IsRepo() {
				return fmt.Errorf("not inside a git repository; use --global to apply globally")
			}

			if err := applyProfile(p, scope); err != nil {
				return err
			}

			ui.Successf("Profile %q applied to %s git config", id, scope)
			fmt.Printf("  %s = %s\n", ui.Muted.Render("user.name "), p.GitUser)
			fmt.Printf("  %s = %s\n", ui.Muted.Render("user.email"), p.GitEmail)
			if p.SSHKeyPath != "" {
				fmt.Printf("  %s = %s\n", ui.Muted.Render("ssh-key   "), p.SSHKeyPath)
			}
			if p.GPGKeyID != "" {
				fmt.Printf("  %s = %s (sign=%v)\n", ui.Muted.Render("gpg-key   "), p.GPGKeyID, p.SignCommits)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "Apply to global git config instead of the current repo")
	rootCmd.AddCommand(cmd)
}
