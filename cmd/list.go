package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hapiio/git-profile/internal/git"
	"github.com/hapiio/git-profile/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all configured profiles",
		Args:    cobra.NoArgs,
		RunE:    runList,
	}
	rootCmd.AddCommand(cmd)
}

func runList(_ *cobra.Command, _ []string) error {
	mgr, err := newManager()
	if err != nil {
		return err
	}
	cfg, err := mgr.Load()
	if err != nil {
		return err
	}

	if len(cfg.Profiles) == 0 {
		ui.Infof("No profiles configured yet.")
		fmt.Println()
		fmt.Println("  Get started with:")
		fmt.Printf("    %s\n", ui.Bold.Render("git-profile add --id work --name \"Your Name\" --email you@company.com"))
		return nil
	}

	// detect the currently active profile in this repo.
	activeName, _ := git.GetConfig("user.name")
	activeEmail, _ := git.GetConfig("user.email")

	ids := cfg.SortedIDs()

	wID, wName, wEmail, wSSH := len("PROFILE"), len("USER"), len("EMAIL"), len("SSH KEY")
	for _, id := range ids {
		p := cfg.Profiles[id]
		if len(id) > wID {
			wID = len(id)
		}
		if len(p.GitUser) > wName {
			wName = len(p.GitUser)
		}
		if len(p.GitEmail) > wEmail {
			wEmail = len(p.GitEmail)
		}
		ssh := sshLabel(p.SSHKeyPath)
		if len(ssh) > wSSH {
			wSSH = len(ssh)
		}
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#5C4EE5", Dark: "#9D8EFF"})
	sep := "  "

	header := fmt.Sprintf("  %-*s%s%-*s%s%-*s%s%s",
		wID+2, "PROFILE", sep,
		wName+2, "USER", sep,
		wEmail+2, "EMAIL", sep,
		"SSH KEY",
	)
	fmt.Println(headerStyle.Render(header))
	fmt.Println(headerStyle.Render("  " + strings.Repeat("─", len(header)-2)))

	for _, id := range ids {
		p := cfg.Profiles[id]
		isActive := activeName != "" && p.GitUser == activeName && p.GitEmail == activeEmail

		row := fmt.Sprintf("%-*s%s%-*s%s%-*s%s%s",
			wID+2, id, sep,
			wName+2, p.GitUser, sep,
			wEmail+2, p.GitEmail, sep,
			sshLabel(p.SSHKeyPath),
		)

		if isActive {
			fmt.Printf("%s %s\n", ui.ActiveMarker, ui.Success.Render(row))
		} else {
			fmt.Printf("  %s\n", row)
		}
	}

	fmt.Printf("\n%s\n", ui.Muted.Render(fmt.Sprintf("%d profile(s)  •  config: %s", len(cfg.Profiles), mgr.Path())))
	return nil
}

func sshLabel(path string) string {
	if path == "" {
		return "(default)"
	}
	return path
}
