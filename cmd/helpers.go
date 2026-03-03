package cmd

import (
	"fmt"
	"os"

	"github.com/hapiio/git-profile/internal/config"
	"github.com/hapiio/git-profile/internal/git"
)

// newManager returns a config.Manager honouring the --config flag.
func newManager() (*config.Manager, error) {
	return config.NewManager(cfgPath)
}

// applyProfile applies git config values for p at the given scope ("local" or "global").
func applyProfile(p config.Profile, scope string) error {
	if err := git.SetConfig(scope, "user.name", p.GitUser); err != nil {
		return fmt.Errorf("setting user.name: %w", err)
	}
	if err := git.SetConfig(scope, "user.email", p.GitEmail); err != nil {
		return fmt.Errorf("setting user.email: %w", err)
	}

	if p.SSHKeyPath != "" {
		key := expandHome(p.SSHKeyPath)
		sshCmd := fmt.Sprintf("ssh -i %s -F /dev/null", key)
		if err := git.SetConfig(scope, "core.sshCommand", sshCmd); err != nil {
			return fmt.Errorf("setting core.sshCommand: %w", err)
		}
	} else if scope == "local" {
		// remove any leftover local sshCommand so the global value takes effect.
		_ = git.UnsetConfig("local", "core.sshCommand")
	}

	if p.GPGKeyID != "" {
		if err := git.SetConfig(scope, "user.signingkey", p.GPGKeyID); err != nil {
			return fmt.Errorf("setting user.signingkey: %w", err)
		}
		val := "false"
		if p.SignCommits {
			val = "true"
		}
		if err := git.SetConfig(scope, "commit.gpgsign", val); err != nil {
			return fmt.Errorf("setting commit.gpgsign: %w", err)
		}
	}

	return nil
}

func expandHome(path string) string {
	if len(path) > 1 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err == nil {
			return home + path[1:]
		}
	}
	return path
}
