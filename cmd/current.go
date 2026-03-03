package cmd

import (
	"fmt"

	"github.com/hapiio/git-profile/internal/git"
	"github.com/hapiio/git-profile/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show the active git identity in this repo",
		Args:  cobra.NoArgs,
		RunE:  runCurrent,
	}
	rootCmd.AddCommand(cmd)
}

func runCurrent(_ *cobra.Command, _ []string) error {
	if !git.IsRepo() {
		return fmt.Errorf("not inside a git repository")
	}

	name, _ := git.GetConfig("user.name")
	email, _ := git.GetConfig("user.email")
	sshCmd, _ := git.GetConfig("core.sshCommand")
	sigKey, _ := git.GetConfig("user.signingkey")
	gpgSign, _ := git.GetConfig("commit.gpgsign")

	fmt.Printf("\n%s\n\n", ui.Bold.Render("Current git identity"))
	printField("user.name ", name, "(not set)")
	printField("user.email", email, "(not set)")
	printField("ssh-key   ", sshCmd, "(default)")
	if sigKey != "" {
		printField("gpg-key   ", sigKey, "")
		printField("gpg-sign  ", gpgSign, "false")
	}

	// show which profile (if any) matches.
	if name != "" || email != "" {
		mgr, err := newManager()
		if err == nil {
			cfg, err := mgr.Load()
			if err == nil {
				for _, p := range cfg.Profiles {
					if p.GitUser == name && p.GitEmail == email {
						fmt.Printf("\n  %s %s\n", ui.Muted.Render("Matched profile:"), ui.Bold.Render(p.ID))
						break
					}
				}
			}
		}
	}

	localDef, _ := git.GetConfig("gitprofile.default")
	globalDef, _ := git.GetGlobalConfig("gitprofile.default")
	if localDef != "" || globalDef != "" {
		fmt.Printf("\n%s\n\n", ui.Bold.Render("Defaults"))
		if localDef != "" {
			printField("local default ", localDef, "")
		}
		if globalDef != "" {
			printField("global default", globalDef, "")
		}
	}

	fmt.Println()
	return nil
}

func printField(key, val, fallback string) {
	label := ui.Muted.Render(key)
	if val != "" {
		fmt.Printf("  %s  %s\n", label, val)
	} else {
		fmt.Printf("  %s  %s\n", label, ui.Muted.Render(fallback))
	}
}
