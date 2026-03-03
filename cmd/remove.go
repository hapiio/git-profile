package cmd

import (
	"fmt"
	"strings"

	"github.com/hapiio/git-profile/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	var yes bool

	cmd := &cobra.Command{
		Use:     "remove <profile-id> [profile-id...]",
		Aliases: []string{"rm", "delete"},
		Short:   "Remove one or more profiles",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr, err := newManager()
			if err != nil {
				return err
			}
			cfg, err := mgr.Load()
			if err != nil {
				return err
			}

			// validate all IDs before removing any.
			for _, id := range args {
				if _, ok := cfg.Profiles[id]; !ok {
					return fmt.Errorf("profile %q not found", id)
				}
			}

			if !yes {
				msg := fmt.Sprintf("Remove profile(s) %s?", quotedList(args))
				confirmed, err := ui.Confirm(msg)
				if err != nil {
					return err
				}
				if !confirmed {
					ui.Infof("Aborted")
					return nil
				}
			}

			for _, id := range args {
				delete(cfg.Profiles, id)
				ui.Successf("Removed profile %q", id)
			}

			return mgr.Save(cfg)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	rootCmd.AddCommand(cmd)
}

func quotedList(ss []string) string {
	quoted := make([]string, len(ss))
	for i, s := range ss {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return strings.Join(quoted, ", ")
}
