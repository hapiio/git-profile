package cmd

import (
	"fmt"
	"strings"

	"github.com/hapiio/git-profile/internal/config"
	"github.com/hapiio/git-profile/internal/git"
	"github.com/hapiio/git-profile/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	var (
		id     string
		global bool
	)

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import the current git identity as a new profile",
		Long: `Create a new profile from the git identity configured in the current
repository (or the global git config with --global).

This is the fastest way to onboard if you already have git configured.`,
		Example: `  git-profile import --id personal --global
  git-profile import --id work`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var name, email, sshCmd string

			if global {
				name, _ = git.GetGlobalConfig("user.name")
				email, _ = git.GetGlobalConfig("user.email")
				sshCmd, _ = git.GetGlobalConfig("core.sshCommand")
			} else {
				if !git.IsRepo() {
					return fmt.Errorf("not inside a git repository; use --global to import from global config")
				}
				name, _ = git.GetConfig("user.name")
				email, _ = git.GetConfig("user.email")
				sshCmd, _ = git.GetConfig("core.sshCommand")
			}

			if name == "" && email == "" {
				scope := "local"
				if global {
					scope = "global"
				}
				return fmt.Errorf("no git identity found in %s config", scope)
			}

			fmt.Printf("\n%s\n", ui.Bold.Render("Found git identity:"))
			fmt.Printf("  user.name  = %s\n", name)
			fmt.Printf("  user.email = %s\n", email)
			if sshCmd != "" {
				fmt.Printf("  ssh-cmd    = %s\n", sshCmd)
			}
			fmt.Println()

			if id == "" {
				if !ui.IsTTY() {
					return fmt.Errorf("--id is required in non-interactive mode")
				}
				var err error
				if id, err = ui.Input("Profile ID for this identity", ""); err != nil {
					return err
				}
			}
			if id == "" {
				return fmt.Errorf("profile ID is required")
			}

			mgr, err := newManager()
			if err != nil {
				return err
			}
			cfg, err := mgr.Load()
			if err != nil {
				return err
			}

			if _, exists := cfg.Profiles[id]; exists {
				return fmt.Errorf("profile %q already exists; use 'git-profile edit %s' to modify it", id, id)
			}

			cfg.Profiles[id] = config.Profile{
				ID:         id,
				GitUser:    name,
				GitEmail:   email,
				SSHKeyPath: extractSSHKey(sshCmd),
			}

			if err := mgr.Save(cfg); err != nil {
				return err
			}

			ui.Successf("Imported current identity as profile %q", id)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Profile ID for the imported identity")
	cmd.Flags().BoolVar(&global, "global", false, "Import from global git config instead of local repo")
	rootCmd.AddCommand(cmd)
}

// extractSSHKey parses the SSH key path from a core.sshCommand value.
// Handles the format: ssh -i /path/to/key -F /dev/null
func extractSSHKey(sshCmd string) string {
	if sshCmd == "" {
		return ""
	}
	parts := strings.Fields(sshCmd)
	for i, part := range parts {
		if part == "-i" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
