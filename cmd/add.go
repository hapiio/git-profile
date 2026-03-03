package cmd

import (
	"fmt"

	"github.com/hapiio/git-profile/internal/config"
	"github.com/hapiio/git-profile/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	var (
		id          string
		name        string
		email       string
		sshKey      string
		gpgKey      string
		signCommits bool
		force       bool
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new git identity profile",
		Long: `Add a named identity profile (user.name, user.email, optional SSH key and GPG signing).

Provide all flags for scripting, or omit them to use the interactive prompt.`,
		Example: `  # Non-interactive (scriptable)
  git-profile add --id work --name "Jane Dev" --email jane@company.com
  git-profile add --id oss  --name "Jane Dev" --email jane@oss.dev --ssh-key ~/.ssh/id_oss

  # Interactive (run with no flags)
  git-profile add`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Interactive mode when required flags are missing and we have a TTY.
			if (id == "" || name == "" || email == "") && ui.IsTTY() {
				return runAddInteractive(&id, &name, &email, &sshKey, &gpgKey, &signCommits)
			}
			if id == "" {
				return fmt.Errorf("--id is required (or run without flags for interactive mode)")
			}
			if name == "" {
				return fmt.Errorf("--name is required (or run without flags for interactive mode)")
			}
			if email == "" {
				return fmt.Errorf("--email is required (or run without flags for interactive mode)")
			}
			return saveNewProfile(id, name, email, sshKey, gpgKey, signCommits, force)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Profile ID, e.g. work or personal")
	cmd.Flags().StringVar(&name, "name", "", "git user.name")
	cmd.Flags().StringVar(&email, "email", "", "git user.email")
	cmd.Flags().StringVar(&sshKey, "ssh-key", "", "SSH private key path (optional)")
	cmd.Flags().StringVar(&gpgKey, "gpg-key", "", "GPG key ID for commit signing (optional)")
	cmd.Flags().BoolVar(&signCommits, "sign-commits", false, "Enable GPG commit signing")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite an existing profile with the same ID")

	rootCmd.AddCommand(cmd)
}

func runAddInteractive(id, name, email, sshKey, gpgKey *string, signCommits *bool) error {
	var err error
	if *id, err = ui.Input("Profile ID (e.g. work, personal)", *id); err != nil {
		return err
	}
	if *id == "" {
		return fmt.Errorf("profile ID is required")
	}
	if *name, err = ui.Input("Git user.name", *name); err != nil {
		return err
	}
	if *name == "" {
		return fmt.Errorf("name is required")
	}
	if *email, err = ui.Input("Git user.email", *email); err != nil {
		return err
	}
	if *email == "" {
		return fmt.Errorf("email is required")
	}
	if *sshKey, err = ui.Input("SSH key path (optional, leave empty to skip)", *sshKey); err != nil {
		return err
	}
	if *gpgKey, err = ui.Input("GPG key ID (optional, leave empty to skip)", *gpgKey); err != nil {
		return err
	}
	if *gpgKey != "" {
		if *signCommits, err = ui.Confirm("Enable GPG commit signing by default?"); err != nil {
			return err
		}
	}
	return saveNewProfile(*id, *name, *email, *sshKey, *gpgKey, *signCommits, false)
}

func saveNewProfile(id, name, email, sshKey, gpgKey string, signCommits, force bool) error {
	mgr, err := newManager()
	if err != nil {
		return err
	}
	cfg, err := mgr.Load()
	if err != nil {
		return err
	}

	if _, exists := cfg.Profiles[id]; exists && !force {
		return fmt.Errorf("profile %q already exists; use --force to overwrite or 'git-profile edit %s' to modify it", id, id)
	}

	cfg.Profiles[id] = config.Profile{
		ID:          id,
		GitUser:     name,
		GitEmail:    email,
		SSHKeyPath:  sshKey,
		GPGKeyID:    gpgKey,
		SignCommits: signCommits,
	}

	if err := mgr.Save(cfg); err != nil {
		return err
	}

	ui.Successf("Profile %q added", id)
	return nil
}
