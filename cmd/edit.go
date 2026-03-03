package cmd

import (
	"fmt"

	"github.com/hapiio/git-profile/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "edit <profile-id>",
		Short: "Edit an existing profile interactively",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			if !ui.IsTTY() {
				return fmt.Errorf("edit requires an interactive terminal")
			}

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

			fmt.Printf("\n%s %s\n%s\n\n",
				ui.Bold.Render("Editing profile"),
				ui.Primary.Render(id),
				ui.Muted.Render("Press Enter to keep the current value."),
			)

			if p.GitUser, err = ui.Input("Git user.name", p.GitUser); err != nil {
				return err
			}
			if p.GitEmail, err = ui.Input("Git user.email", p.GitEmail); err != nil {
				return err
			}
			if p.SSHKeyPath, err = ui.Input("SSH key path", p.SSHKeyPath); err != nil {
				return err
			}
			if p.GPGKeyID, err = ui.Input("GPG key ID", p.GPGKeyID); err != nil {
				return err
			}
			if p.GPGKeyID != "" {
				prompt := fmt.Sprintf("Enable GPG commit signing? (currently: %v)", p.SignCommits)
				if p.SignCommits, err = ui.Confirm(prompt); err != nil {
					return err
				}
			}

			cfg.Profiles[id] = p

			if err := mgr.Save(cfg); err != nil {
				return err
			}

			ui.Successf("Profile %q updated", id)
			return nil
		},
	}

	rootCmd.AddCommand(cmd)
}
