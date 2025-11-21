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
	} else if scope == "local" {
		// Clear repo-specific sshCommand if present
		_ = exec.Command("git", "config", "--unset", "core.sshCommand").Run()
	}

	return nil
}

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
	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}

	if len(cfg.Profiles) == 0 {
		fmt.Println("No profiles configured yet.")
		return nil
	}

	fmt.Println("Configured profiles:")
	for id, p := range cfg.Profiles {
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

	return nil
}

func cmdChoose(args []string) error {
	_ = args

	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	if len(cfg.Profiles) == 0 {
		return fmt.Errorf("no profiles configured; run `git-profile add` first")
	}

	fmt.Println("Select profile:")
	ids := make([]string, 0, len(cfg.Profiles))
	i := 1
	for id, p := range cfg.Profiles {
		fmt.Printf("  [%d] %s: %s <%s>\n", i, id, p.GitUser, p.GitEmail)
		ids = append(ids, id)
		i++
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

	//alway apply to local repo for choose()
	if err := applyProfile(p, "local"); err != nil {
		return err
	}

	fmt.Printf("Applied profile %q to local repo\n", selectedID)
	return nil
}

func usage() {
	fmt.Println(`git-profile - manage multiple git/GitHub identity profiles

Usage:
  git-profile add     --id <id> --name "<User Name>" --email "email@example.com" [--ssh-key /path/to/key]
  git-profile list
  git-profile use     [--global] <id>
  git-profile current
  git-profile choose  (interactive selector; good for wrapping git commit/push)`,
	)
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
