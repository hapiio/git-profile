package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type Profile struct {
	ID         string `json:"id"`
	GitUser    string `json:"git_user"`
	GitEmail   string `json:"git_email"`
	SSHKeyPath string `json:"ssh_key_path,omitempty"`
}

type Config struct {
	Profiles map[string]Profile `json:"profiles"`
}

// ----- Config file handling -----

func defaultConfigPath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cfgDir, "gitprofile")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func loadConfig() (*Config, string, error) {
	path, err := defaultConfigPath()
	if err != nil {
		return nil, "", err
	}

	cfg := &Config{Profiles: make(map[string]Profile)}

	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, path, nil // empty config
		}
		return nil, "", err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(cfg); err != nil {
		return nil, "", err
	}

	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}

	return cfg, path, nil
}

func saveConfig(cfg *Config, path string) error {
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cfg); err != nil {
		return err
	}

	return os.Rename(tmp, path)
}

// ----- Git config helpers -----

// scope: "global" or anything else (treated as local)
func runGitConfig(scope string, key string, value string) error {
	args := []string{"config"}
	if scope == "global" {
		args = append(args, "--global")
	}
	args = append(args, key, value)

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getGitConfig(key string) (string, error) {
	cmd := exec.Command("git", "config", "--get", key)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func getGitConfigGlobal(key string) (string, error) {
	cmd := exec.Command("git", "config", "--global", "--get", key)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func applyProfile(p Profile, scope string) error {
	if err := runGitConfig(scope, "user.name", p.GitUser); err != nil {
		return fmt.Errorf("setting user.name: %w", err)
	}
	if err := runGitConfig(scope, "user.email", p.GitEmail); err != nil {
		return fmt.Errorf("setting user.email: %w", err)
	}

	if p.SSHKeyPath != "" {
		sshCmd := fmt.Sprintf("ssh -i %s -F /dev/null", p.SSHKeyPath)
		if err := runGitConfig(scope, "core.sshCommand", sshCmd); err != nil {
			return fmt.Errorf("setting core.sshCommand: %w", err)
		}
	} else if scope != "global" {
		// Clear repo-specific sshCommand if present
		_ = exec.Command("git", "config", "--unset", "core.sshCommand").Run()
	}

	return nil
}

func gitDir() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// ----- Commands -----

func cmdAdd(args []string) error {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	id := fs.String("id", "", "Profile ID (e.g. work, personal)")
	name := fs.String("name", "", "Git user.name")
	email := fs.String("email", "", "Git user.email")
	sshKey := fs.String("ssh-key", "", "SSH key path (optional)")
	_ = fs.Parse(args)

	if *id == "" || *name == "" || *email == "" {
		return fmt.Errorf("id, name and email are required")
	}

	cfg, path, err := loadConfig()
	if err != nil {
		return err
	}

	if _, exists := cfg.Profiles[*id]; exists {
		return fmt.Errorf("profile %q already exists", *id)
	}

	p := Profile{
		ID:         *id,
		GitUser:    *name,
		GitEmail:   *email,
		SSHKeyPath: *sshKey,
	}
	cfg.Profiles[*id] = p

	if err := saveConfig(cfg, path); err != nil {
		return err
	}

	fmt.Printf("Added profile %q\n", *id)
	return nil
}

func cmdList(args []string) error {
	_ = args

	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}

	if len(cfg.Profiles) == 0 {
		fmt.Println("No profiles configured yet.")
		return nil
	}

	ids := make([]string, 0, len(cfg.Profiles))
	for id := range cfg.Profiles {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	fmt.Println("Configured profiles:")
	for _, id := range ids {
		p := cfg.Profiles[id]
		ssh := p.SSHKeyPath
		if ssh == "" {
			ssh = "(default SSH)"
		}
		fmt.Printf("  - %s: %s <%s>, ssh=%s\n", id, p.GitUser, p.GitEmail, ssh)
	}
	return nil
}

func cmdUse(args []string) error {
	fs := flag.NewFlagSet("use", flag.ExitOnError)
	scopeGlobal := fs.Bool("global", false, "Apply profile to global git config")
	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		return fmt.Errorf("usage: git-profile use [--global] <profile-id>")
	}
	id := fs.Arg(0)

	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}

	p, ok := cfg.Profiles[id]
	if !ok {
		return fmt.Errorf("profile %q not found", id)
	}

	scope := "local"
	if *scopeGlobal {
		scope = "global"
	}

	if err := applyProfile(p, scope); err != nil {
		return err
	}

	fmt.Printf("Applied profile %q to %s git config\n", id, scope)
	return nil
}

func cmdCurrent(args []string) error {
	_ = args

	name, errName := getGitConfig("user.name")
	email, errEmail := getGitConfig("user.email")

	if errName != nil && errEmail != nil {
		return fmt.Errorf("no user.name or user.email set in this repo")
	}

	fmt.Println("Current git identity (this repo):")
	if errName == nil {
		fmt.Printf("  user.name  = %s\n", name)
	}
	if errEmail == nil {
		fmt.Printf("  user.email = %s\n", email)
	}

	ssh, errSSH := getGitConfig("core.sshCommand")
	if errSSH == nil && ssh != "" {
		fmt.Printf("  core.sshCommand = %s\n", ssh)
	} else {
		fmt.Println("  core.sshCommand = (default)")
	}

	def, errDef := getGitConfig("gitprofile.default")
	if errDef == nil && def != "" {
		fmt.Printf("  gitprofile.default (local) = %s\n", def)
	}
	globalDef, errGDef := getGitConfigGlobal("gitprofile.default")
	if errGDef == nil && globalDef != "" {
		fmt.Printf("  gitprofile.default (global) = %s\n", globalDef)
	}

	return nil
}

func cmdChoose(_ []string) error {
	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	if len(cfg.Profiles) == 0 {
		return fmt.Errorf("no profiles configured; run `git-profile add` first")
	}

	ids := make([]string, 0, len(cfg.Profiles))
	for id := range cfg.Profiles {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	fmt.Println("Select profile:")
	for i, id := range ids {
		p := cfg.Profiles[id]
		fmt.Printf("  [%d] %s: %s <%s>\n", i+1, id, p.GitUser, p.GitEmail)
	}

	fmt.Print("Enter number: ")
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return fmt.Errorf("no selection made")
	}

	var choice int
	_, err = fmt.Sscanf(line, "%d", &choice)
	if err != nil || choice < 1 || choice > len(ids) {
		return fmt.Errorf("invalid selection")
	}

	selectedID := ids[choice-1]
	p := cfg.Profiles[selectedID]

	// Always apply to local repo for choose()
	if err := applyProfile(p, "local"); err != nil {
		return err
	}

	fmt.Printf("Applied profile %q to local repo\n", selectedID)
	return nil
}

// set-default: store default profile in git config
func cmdSetDefault(args []string) error {
	fs := flag.NewFlagSet("set-default", flag.ExitOnError)
	scopeGlobal := fs.Bool("global", false, "Set as global default profile")
	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		return fmt.Errorf("usage: git-profile set-default [--global] <profile-id>")
	}
	id := fs.Arg(0)

	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}

	if _, ok := cfg.Profiles[id]; !ok {
		return fmt.Errorf("profile %q not found", id)
	}

	scope := "local"
	if *scopeGlobal {
		scope = "global"
	}

	if err := runGitConfig(scope, "gitprofile.default", id); err != nil {
		return err
	}

	fmt.Printf("Set %q as %s default profile\n", id, scope)
	return nil
}

// ensure: used by hooks
// 1) If local gitprofile.default exists and matches a profile -> apply it
// 2) Else if global gitprofile.default exists and matches -> apply it
// 3) Else -> interactive choose()
func cmdEnsure(args []string) error {
	_ = args

	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	if len(cfg.Profiles) == 0 {
		return fmt.Errorf("no profiles configured; run `git-profile add` first")
	}

	// 1) Local default
	if def, err := getGitConfig("gitprofile.default"); err == nil && def != "" {
		if p, ok := cfg.Profiles[def]; ok {
			return applyProfile(p, "local")
		}
	}

	// 2) Global default
	if gdef, err := getGitConfigGlobal("gitprofile.default"); err == nil && gdef != "" {
		if p, ok := cfg.Profiles[gdef]; ok {
			return applyProfile(p, "local")
		}
	}

	// 3) Fallback: interactive
	return cmdChoose(nil)
}

// install-hooks: installs prepare-commit-msg & pre-push hooks for this repo
func cmdInstallHooks(args []string) error {
	_ = args

	gd, err := gitDir()
	if err != nil {
		return fmt.Errorf("not a git repo? %w", err)
	}

	hooksDir := filepath.Join(gd, "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return err
	}

	hookContent := `#!/bin/sh
# git-profile hook: ensure correct profile before commit/push
git-profile ensure >/dev/null 2>&1 || true
`

	hooks := []string{"prepare-commit-msg", "pre-push"}

	for _, name := range hooks {
		path := filepath.Join(hooksDir, name)
		if err := os.WriteFile(path, []byte(hookContent), 0o755); err != nil {
			return fmt.Errorf("writing hook %s: %w", name, err)
		}
	}

	fmt.Printf("Installed git-profile hooks in %s\n", hooksDir)
	fmt.Println("From now on, normal `git commit` and `git push` will apply/ask for a profile.")
	return nil
}

// ----- Usage / main -----

func usage() {
	fmt.Println(`git-profile - manage multiple git/GitHub identity profiles

Usage:
  git-profile add         --id <id> --name "<User Name>" --email "email@example.com" [--ssh-key /path/to/key]
  git-profile list
  git-profile use         [--global] <id>
  git-profile current
  git-profile choose
  git-profile set-default [--global] <id>
  git-profile ensure
  git-profile install-hooks

Commands:
  add           Add a new identity profile
  list          List configured profiles
  use           Apply a profile to this repo or globally
  current       Show current git identity and defaults
  choose        Interactively choose a profile and apply locally
  set-default   Set per-repo or global default profile (stored in git config)
  ensure        Apply repo default, then global default, otherwise prompt (used by hooks)
  install-hooks Install hooks so plain 'git commit' and 'git push' call 'git-profile ensure'`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error

	switch cmd {
	case "add":
		err = cmdAdd(args)
	case "list":
		err = cmdList(args)
	case "use":
		err = cmdUse(args)
	case "current":
		err = cmdCurrent(args)
	case "choose":
		err = cmdChoose(args)
	case "set-default":
		err = cmdSetDefault(args)
	case "ensure":
		err = cmdEnsure(args)
	case "install-hooks":
		err = cmdInstallHooks(args)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
