package cmd

import (
	"fmt"

	"github.com/hapiio/git-profile/internal/config"
	"github.com/hapiio/git-profile/internal/git"
	"github.com/hapiio/git-profile/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "choose",
		Short: "Interactively pick and apply a profile to this repo",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChoose()
		},
	}
	rootCmd.AddCommand(cmd)
}

// runChoose shows an interactive numbered list and applies the chosen profile locally.
func runChoose() error {
	if !ui.IsTTY() {
		return fmt.Errorf("choose requires an interactive terminal; use 'git-profile use <id>' instead")
	}
	if !git.IsRepo() {
		return fmt.Errorf("not inside a git repository")
	}

	mgr, err := newManager()
	if err != nil {
		return err
	}
	cfg, err := mgr.Load()
	if err != nil {
		return err
	}

	return chooseAndApply(cfg)
}

// chooseAndApply is the shared implementation used by 'choose' and 'ensure'.
func chooseAndApply(cfg *config.Config) error {
	if len(cfg.Profiles) == 0 {
		return fmt.Errorf("no profiles configured; run 'git-profile add' first")
	}

	ids := cfg.SortedIDs()
	opts := make([]string, len(ids))
	for i, id := range ids {
		p := cfg.Profiles[id]
		opts[i] = fmt.Sprintf("%-15s  %s <%s>", ui.Bold.Render(id), p.GitUser, p.GitEmail)
	}

	idx, err := ui.Select("Select a git profile", opts)
	if err != nil {
		return err
	}

	selectedID := ids[idx]
	p := cfg.Profiles[selectedID]

	if err := applyProfile(p, "local"); err != nil {
		return err
	}

	ui.Successf("Profile %q applied to local repo", selectedID)
	return nil
}
