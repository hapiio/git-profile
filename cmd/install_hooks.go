package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hapiio/git-profile/internal/git"
	"github.com/hapiio/git-profile/internal/ui"
	"github.com/spf13/cobra"
)

const hookMarker = "# managed-by: git-profile"

const hookContent = `#!/bin/sh
# managed-by: git-profile
# Ensures the correct git identity is applied before each commit/push.
# Reinstall with: git-profile install-hooks
git-profile ensure >/dev/null 2>&1 || true
`

func init() {
	cmd := &cobra.Command{
		Use:   "install-hooks",
		Short: "Install git hooks to auto-apply profiles on commit/push",
		Long: `Installs prepare-commit-msg and pre-push hooks in the current repository.

The hooks call 'git-profile ensure' before each commit and push, ensuring the
correct identity is always active.

If a hook already exists and was not installed by git-profile, the git-profile
call is appended rather than overwriting your existing hook.`,
		Args: cobra.NoArgs,
		RunE: runInstallHooks,
	}
	rootCmd.AddCommand(cmd)
}

func runInstallHooks(_ *cobra.Command, _ []string) error {
	gd, err := git.Dir()
	if err != nil {
		return fmt.Errorf("not inside a git repository")
	}

	hooksDir := filepath.Join(gd, "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return fmt.Errorf("creating hooks directory: %w", err)
	}

	hooks := []string{"prepare-commit-msg", "pre-push"}
	for _, name := range hooks {
		path := filepath.Join(hooksDir, name)
		if err := installHook(path, name); err != nil {
			return err
		}
	}

	fmt.Println()
	ui.Infof("Hooks are active. 'git commit' and 'git push' will now enforce git-profile identity.")
	fmt.Printf("  %s\n", ui.Muted.Render("To apply a default automatically: git-profile set-default <id>"))
	return nil
}

func installHook(path, name string) error {
	existing, readErr := os.ReadFile(path)

	switch {
	case readErr == nil && strings.Contains(string(existing), hookMarker):
		// already installed by us, overwrite to pick up any updates.
		if err := os.WriteFile(path, []byte(hookContent), 0o755); err != nil {
			return fmt.Errorf("updating hook %s: %w", name, err)
		}
		ui.Successf("Updated hook: .git/hooks/%s", name)

	case readErr == nil:
		// hook exists but was not installed by us, append our call.
		appended := string(existing) + "\n" + hookMarker + "\ngit-profile ensure >/dev/null 2>&1 || true\n"
		if err := os.WriteFile(path, []byte(appended), 0o755); err != nil {
			return fmt.Errorf("appending to hook %s: %w", name, err)
		}
		ui.Successf("Appended to existing hook: .git/hooks/%s", name)
		ui.Warningf("Existing hook content was preserved. Review .git/hooks/%s if needed.", name)

	default:
		// no hook yet, create fresh.
		if err := os.WriteFile(path, []byte(hookContent), 0o755); err != nil {
			return fmt.Errorf("writing hook %s: %w", name, err)
		}
		ui.Successf("Installed hook: .git/hooks/%s", name)
	}

	return nil
}
