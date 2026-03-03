package cmd

import (
	"fmt"

	"github.com/hapiio/git-profile/internal/git"
	"github.com/hapiio/git-profile/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "ensure",
		Short: "Ensure a profile is applied (used by git hooks)",
		Long: `Applies the correct identity for the current repository:

  1. If a local default is set (gitprofile.default), applies it.
  2. Otherwise, if a global default is set, applies it.
  3. Otherwise, prompts interactively (requires a terminal).

This command is invoked automatically by hooks installed with 'install-hooks'.`,
		Args:   cobra.NoArgs,
		Hidden: false,
		RunE:   runEnsure,
	}
	rootCmd.AddCommand(cmd)
}

func runEnsure(_ *cobra.Command, _ []string) error {
	mgr, err := newManager()
	if err != nil {
		return err
	}
	cfg, err := mgr.Load()
	if err != nil {
		return err
	}

	if len(cfg.Profiles) == 0 {
		return fmt.Errorf("no profiles configured; run 'git-profile add' first")
	}

	// local per-repo default.
	if def, err := git.GetConfig("gitprofile.default"); err == nil && def != "" {
		if p, ok := cfg.Profiles[def]; ok {
			ui.Infof("Applying local default profile %q", def)
			return applyProfile(p, "local")
		}
		ui.Warningf("Local default profile %q not found; falling through", def)
	}

	// global default.
	if gdef, err := git.GetGlobalConfig("gitprofile.default"); err == nil && gdef != "" {
		if p, ok := cfg.Profiles[gdef]; ok {
			ui.Infof("Applying global default profile %q", gdef)
			return applyProfile(p, "local")
		}
		ui.Warningf("Global default profile %q not found; falling through", gdef)
	}

	// interactive fallback — only when a TTY is available.
	if !ui.IsTTY() {
		return fmt.Errorf("no default profile set; run 'git-profile set-default <id>' to configure one")
	}

	return chooseAndApply(cfg)
}
