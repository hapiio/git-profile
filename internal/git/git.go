// Package git provides thin wrappers around the git command-line tool.
package git

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// SetConfig runs `git config [--local|--global] key value`.
func SetConfig(scope, key, value string) error {
	args := buildScopeArgs(scope)
	args = append(args, key, value)
	if out, err := exec.Command("git", append([]string{"config"}, args...)...).CombinedOutput(); err != nil {
		return fmt.Errorf("git config %s %s: %w\n%s", key, value, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// UnsetConfig runs `git config [--local|--global] --unset key`.
// Returns nil if the key was not present (exit code 5).
func UnsetConfig(scope, key string) error {
	args := buildScopeArgs(scope)
	args = append(args, "--unset", key)
	cmd := exec.Command("git", append([]string{"config"}, args...)...)
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 5 {
			return nil // key was not set — no-op
		}
		return err
	}
	return nil
}

// GetConfig reads a local-scope git config key.
func GetConfig(key string) (string, error) {
	return runGitConfigGet(nil, key)
}

// GetGlobalConfig reads a global-scope git config key.
func GetGlobalConfig(key string) (string, error) {
	return runGitConfigGet([]string{"--global"}, key)
}

func runGitConfigGet(extraArgs []string, key string) (string, error) {
	args := append([]string{"config"}, extraArgs...)
	args = append(args, "--get", key)
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// Dir returns the .git directory path for the current repository.
func Dir() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--git-dir").Output()
	if err != nil {
		return "", fmt.Errorf("not inside a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}

// IsRepo reports whether the working directory is inside a git repository.
func IsRepo() bool {
	_, err := Dir()
	return err == nil
}

func buildScopeArgs(scope string) []string {
	switch scope {
	case "global":
		return []string{"--global"}
	case "local":
		return []string{"--local"}
	default:
		return nil
	}
}
