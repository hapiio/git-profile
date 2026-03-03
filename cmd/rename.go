package cmd

import (
	"fmt"

	"github.com/hapiio/git-profile/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:     "rename <old-id> <new-id>",
		Aliases: []string{"mv"},
		Short:   "Rename a profile",
		Example: `  git-profile rename work work-corp`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldID, newID := args[0], args[1]

			mgr, err := newManager()
			if err != nil {
				return err
			}
			cfg, err := mgr.Load()
			if err != nil {
				return err
			}

			p, ok := cfg.Profiles[oldID]
			if !ok {
				return fmt.Errorf("profile %q not found", oldID)
			}
			if _, exists := cfg.Profiles[newID]; exists {
				return fmt.Errorf("profile %q already exists; choose a different name", newID)
			}

			p.ID = newID
			cfg.Profiles[newID] = p
			delete(cfg.Profiles, oldID)

			if err := mgr.Save(cfg); err != nil {
				return err
			}

			ui.Successf("Renamed %q → %q", oldID, newID)
			ui.Warningf("Repos with 'gitprofile.default = %s' must be updated manually:\n  git config gitprofile.default %s", oldID, newID)
			return nil
		},
	}

	rootCmd.AddCommand(cmd)
}
